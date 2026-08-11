package server_test

/*
MIT License

Copyright (c) 2026 Shane

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
*/

import (
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Bugs5382/go-hl7/client/hl7/metadata"
	"github.com/Bugs5382/go-hl7/server"
)

// versionProbe sends one MLLP-framed HL7 message whose MSH.12 is mshVersion to a
// listener built from opts, and reports whether the handler ran and the ACK the
// listener wrote back. A raw socket is used (rather than the go-hl7 client)
// because the client enforces its own version on send and would reject a
// deliberate mismatch before it ever left the wire.
func versionProbe(t *testing.T, opts server.ListenerOptions, mshVersion, controlID string) (handlerRan bool, ack string) {
	t.Helper()

	port := freePort(t)
	opts.Port = ptr(port)

	var handlerCalls atomic.Int32
	srv, err := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	if err != nil {
		t.Fatal(err)
	}
	listener, err := srv.CreateInbound(opts, func(_ *server.InboundRequest, res server.ResponseSender) error {
		handlerCalls.Add(1)
		return res.SendResponse("AA")
	})
	if err != nil {
		t.Fatalf("CreateInbound: %v", err)
	}
	defer func() { _ = listener.Close() }()
	waitFor(t, listener.IsListening)

	body := joinSegs(
		`MSH|^~\&|EPIC|HOSP|RECV|RFAC|20240101000000||ADT^A08|`+controlID+`|P|`+mshVersion,
		`EVN|A08|20240101000000`,
	)
	const VT, FS, CR = 0x0b, 0x1c, 0x0d
	framed := append([]byte{VT}, append([]byte(body), FS, CR)...)

	raw, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = raw.Close() }()

	got := make(chan string, 1)
	go func() {
		buf := make([]byte, 4096)
		n, _ := raw.Read(buf)
		got <- string(buf[:n])
	}()
	if _, err := raw.Write(framed); err != nil {
		t.Fatal(err)
	}

	select {
	case ack = <-got:
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for ack")
	}
	return handlerCalls.Load() > 0, ack
}

// TestInboundConstructionGuards covers the CreateInbound construction guards for
// the relaxed version constraint: the version forms are mutually exclusive, the
// allow-list may not be empty or contain an unknown version, and a missing
// version without AcceptAnyVersion stays an error.
func TestInboundConstructionGuards(t *testing.T) {
	newSrv := func() *server.Server {
		s, err := server.NewServer(nil)
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	noop := func(*server.InboundRequest, server.ResponseSender) error { return nil }

	t.Run("acceptAnyVersion combined with Version is rejected", func(t *testing.T) {
		_, err := newSrv().CreateInbound(server.ListenerOptions{Port: ptr(4000), Version: "2.7", AcceptAnyVersion: true}, noop)
		if err == nil || err.Error() != "acceptAnyVersion cannot be combined with an explicit version or version list." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("acceptAnyVersion combined with Versions is rejected", func(t *testing.T) {
		_, err := newSrv().CreateInbound(server.ListenerOptions{Port: ptr(4000), Versions: []string{"2.5"}, AcceptAnyVersion: true}, noop)
		if err == nil || err.Error() != "acceptAnyVersion cannot be combined with an explicit version or version list." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("empty allow-list without a version is rejected", func(t *testing.T) {
		_, err := newSrv().CreateInbound(server.ListenerOptions{Port: ptr(4000), Versions: []string{}}, noop)
		if err == nil || err.Error() != "version is not defined." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("missing version without acceptAnyVersion is rejected", func(t *testing.T) {
		_, err := newSrv().CreateInbound(server.ListenerOptions{Port: ptr(4000)}, noop)
		if err == nil || err.Error() != "version is not defined." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("unknown element in the allow-list is rejected", func(t *testing.T) {
		_, err := newSrv().CreateInbound(server.ListenerOptions{Port: ptr(4000), Versions: []string{"2.5", "9.9"}}, noop)
		if err == nil || err.Error() != "version is not a valid HL7 version." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("valid allow-list constructs", func(t *testing.T) {
		in, err := newSrv().CreateInbound(server.ListenerOptions{Port: ptr(freePort(t)), Versions: []string{"2.3.1", "2.4", "2.5"}}, noop)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = in.Close()
	})

	t.Run("acceptAnyVersion alone constructs", func(t *testing.T) {
		in, err := newSrv().CreateInbound(server.ListenerOptions{Port: ptr(freePort(t)), AcceptAnyVersion: true}, noop)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_ = in.Close()
	})
}

// TestInboundAcceptAnyVersion asserts a listener built with AcceptAnyVersion
// accepts every known HL7 version, and still AR-rejects an unknown/garbage
// MSH.12.
func TestInboundAcceptAnyVersion(t *testing.T) {
	for _, v := range metadata.KnownVersions {
		v := string(v)
		t.Run("accepts "+v, func(t *testing.T) {
			ran, ack := versionProbe(t, server.ListenerOptions{AcceptAnyVersion: true}, v, "ANY_"+strings.ReplaceAll(v, ".", "_"))
			if !ran {
				t.Fatalf("handler did not run for accepted version %s", v)
			}
			if !strings.Contains(ack, "MSA|AA|") {
				t.Fatalf("ack for %s was not an AA: %q", v, ack)
			}
		})
	}

	t.Run("AR-rejects an unknown version", func(t *testing.T) {
		ran, ack := versionProbe(t, server.ListenerOptions{AcceptAnyVersion: true}, "9.9", "ANY_GARBAGE")
		if ran {
			t.Fatal("handler ran for an unknown version")
		}
		if !strings.Contains(ack, "MSA|AR|ANY_GARBAGE") {
			t.Fatalf("ack was not an AR for the unknown version: %q", ack)
		}
	})
}

// TestInboundVersionAllowList asserts an allow-list listener accepts every
// member of its set and AR-rejects a version outside it.
func TestInboundVersionAllowList(t *testing.T) {
	allow := server.ListenerOptions{Versions: []string{"2.3.1", "2.4", "2.5"}}

	for _, v := range []string{"2.3.1", "2.4", "2.5"} {
		t.Run("accepts "+v, func(t *testing.T) {
			ran, ack := versionProbe(t, allow, v, "LIST_"+strings.ReplaceAll(v, ".", "_"))
			if !ran {
				t.Fatalf("handler did not run for allow-listed version %s", v)
			}
			if !strings.Contains(ack, "MSA|AA|") {
				t.Fatalf("ack for %s was not an AA: %q", v, ack)
			}
		})
	}

	t.Run("AR-rejects a version outside the allow-list", func(t *testing.T) {
		ran, ack := versionProbe(t, allow, "2.7", "LIST_OUTSIDE")
		if ran {
			t.Fatal("handler ran for a version outside the allow-list")
		}
		if !strings.Contains(ack, "MSA|AR|LIST_OUTSIDE") {
			t.Fatalf("ack was not an AR for the out-of-set version: %q", ack)
		}
	})
}

// TestInboundSingleVersionBackwardCompatible asserts the original single-Version
// form still accepts its version and AR-rejects any other.
func TestInboundSingleVersionBackwardCompatible(t *testing.T) {
	single := server.ListenerOptions{Version: "2.7"}

	t.Run("accepts the pinned version", func(t *testing.T) {
		ran, ack := versionProbe(t, single, "2.7", "SINGLE_MATCH")
		if !ran || !strings.Contains(ack, "MSA|AA|") {
			t.Fatalf("pinned version not accepted: ran=%v ack=%q", ran, ack)
		}
	})

	t.Run("AR-rejects a different version", func(t *testing.T) {
		ran, ack := versionProbe(t, single, "2.5", "SINGLE_MISMATCH")
		if ran {
			t.Fatal("handler ran for a mismatched version")
		}
		if !strings.Contains(ack, "MSA|AR|SINGLE_MISMATCH") {
			t.Fatalf("ack was not an AR for the mismatch: %q", ack)
		}
	})
}
