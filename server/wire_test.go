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
	"encoding/base64"
	"net"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/client"
	"github.com/Bugs5382/go-hl7/server"
)

// ipv6Loopback reports whether ::1 is usable on this host; the IPv6 wire tests
// skip when it is not (CI runners with IPv6 disabled).
func ipv6Loopback(t *testing.T) bool {
	t.Helper()
	l, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		return false
	}
	_ = l.Close()
	return true
}

// TestEnd2EndBatchTwoMessages mirrors the no-tls "send batch with two message,
// get proper ACK": one frame carries a Batch of two messages, the client
// receives two ACKs (totalSent 1, totalAck 2).
func TestEnd2EndBatchTwoMessages(t *testing.T) {
	port := freePort(t)
	var acks atomic.Int32
	done := newEventWaiter()

	srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(req *server.InboundRequest, res server.ResponseSender) error {
		if req.GetMessage().Get("MSH.12").String() != "2.7" {
			t.Errorf("MSH.12 = %q, want 2.7", req.GetMessage().Get("MSH.12").String())
		}
		return res.SendResponse("AA")
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, listener.IsListening)

	cli, _ := client.NewClient(client.ClientOptions{Host: "127.0.0.1", IPv4: ptr(true)})
	outbound, err := cli.CreateConnection(client.ClientListenerOptions{Port: ptr(port)}, func(res *client.InboundResponse) error {
		if res.GetMessage().Get("MSA.1").String() == "AA" {
			if acks.Add(1) == 2 {
				done.signal()
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, outbound.IsConnected)

	batch, err := builder.NewBatch(builder.BatchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	batch.Start("")
	msg := makeTestMessage(t, "CONTROL_ID")
	batch.Add(msg, -1)
	batch.Add(msg, -1)
	batch.End()

	if err := outbound.SendMessage(batch); err != nil {
		t.Fatal(err)
	}
	done.wait(t, "two acks")

	if cli.TotalSent() != 1 {
		t.Fatalf("TotalSent = %d, want 1", cli.TotalSent())
	}
	waitFor(t, func() bool { return cli.TotalAck() == 2 })

	_ = outbound.Close()
	_ = listener.Close()
	cli.CloseAll()
}

// TestEnd2EndTLS mirrors the "...tls ...simple" block: a self-signed server
// cert, a client that skips verification, AA round-trip.
func TestEnd2EndTLS(t *testing.T) {
	port := freePort(t)
	certs := tlsTestCerts(t)
	done := newEventWaiter()
	var ackOK atomic.Bool

	srv, err := server.NewServer(&server.ServerOptions{
		BindAddress: ptr("127.0.0.1"),
		IPv4:        ptr(true),
		TLS:         &server.TLSConfig{Cert: certs.cert, Key: certs.key},
	})
	if err != nil {
		t.Fatal(err)
	}
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(req *server.InboundRequest, res server.ResponseSender) error {
		if req.GetMessage().Get("MSH.12").String() != "2.7" {
			t.Errorf("MSH.12 = %q, want 2.7", req.GetMessage().Get("MSH.12").String())
		}
		return res.SendResponse("AA")
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, listener.IsListening)

	reject := false
	cli, _ := client.NewClient(client.ClientOptions{
		Host: "127.0.0.1",
		IPv4: ptr(true),
		TLS:  &client.TLSConfig{RejectUnauthorized: &reject},
	})
	outbound, err := cli.CreateConnection(client.ClientListenerOptions{Port: ptr(port)}, func(res *client.InboundResponse) error {
		ackOK.Store(res.GetMessage().Get("MSA.1").String() == "AA")
		done.signal()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, outbound.IsConnected)

	if err := outbound.SendMessage(makeTestMessage(t, "CONTROL_ID")); err != nil {
		t.Fatal(err)
	}
	done.wait(t, "tls ack")
	if !ackOK.Load() {
		t.Fatalf("MSA.1 was not AA over TLS")
	}

	_ = outbound.Close()
	_ = listener.Close()
	cli.CloseAll()
}

// TestEnd2EndLargeData mirrors the large-encapsulated-data check: an OBX with a
// large base64 payload round-trips and the server reads OBX.3.1.
func TestEnd2EndLargeData(t *testing.T) {
	port := freePort(t)
	done := newEventWaiter()
	var sawPDF atomic.Bool

	srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(req *server.InboundRequest, res server.ResponseSender) error {
		m := req.GetMessage()
		sawPDF.Store(m.Get("MSH.12").String() == "2.7" && m.Get("OBX.3.1").String() == "SOME-PDF")
		return res.SendResponse("AA")
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, listener.IsListening)

	cli, _ := client.NewClient(client.ClientOptions{Host: "127.0.0.1", IPv4: ptr(true)})
	var ackOK atomic.Bool
	outbound, err := cli.CreateConnection(client.ClientListenerOptions{Port: ptr(port)}, func(res *client.InboundResponse) error {
		ackOK.Store(res.GetMessage().Get("MSA.1").String() == "AA")
		done.signal()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, outbound.IsConnected)

	msg := makeTestMessage(t, "CONTROL_ID")
	obx, err := msg.AddSegment("OBX")
	if err != nil {
		t.Fatal(err)
	}
	obx.Set("1", "1")
	obx.Set("2", "ED")
	obx.Set("3.1", "SOME-PDF")
	obx.Set("3.2", "Result Report")
	obx.Set("3.3", "99COD")
	obx.Set("5.2", "application")
	obx.Set("5.3", "pdf")
	obx.Set("5.4", "Base64")
	obx.Set("5.5", largeBase64(600))
	obx.Set("8", "A")
	obx.Set("11", "F")
	obx.Set("14", "20240625103600")

	if err := outbound.SendMessage(msg); err != nil {
		t.Fatal(err)
	}
	done.wait(t, "large ack")
	if !sawPDF.Load() {
		t.Fatalf("server did not read OBX.3.1 = SOME-PDF")
	}
	if !ackOK.Load() {
		t.Fatalf("MSA.1 was not AA")
	}
	if cli.TotalSent() != 1 || cli.TotalAck() != 1 {
		waitFor(t, func() bool { return cli.TotalSent() == 1 && cli.TotalAck() == 1 })
	}

	_ = outbound.Close()
	_ = listener.Close()
	cli.CloseAll()
}

// TestDualStackIPv4 mirrors "IPv4 client -> IPv4 server (loopback)".
func TestDualStackIPv4(t *testing.T) {
	runLoopback(t, "127.0.0.1", "127.0.0.1", ptr(true), ptr(false), ptr(true), ptr(false))
}

// TestDualStackIPv6 mirrors "IPv6 client -> IPv6 server (loopback)".
func TestDualStackIPv6(t *testing.T) {
	if !ipv6Loopback(t) {
		t.Skip("::1 not available on this host")
	}
	runLoopback(t, "::1", "::1", ptr(false), ptr(true), ptr(false), ptr(true))
}

// TestDualStackServerBoth mirrors "dual-stack server accepts both IPv4 and IPv6
// clients": a single dual-stack listener serves a v4 then a v6 client.
func TestDualStackServerBoth(t *testing.T) {
	if !ipv6Loopback(t) {
		t.Skip("::1 not available on this host")
	}
	port := freePort(t)

	srv, _ := server.NewServer(&server.ServerOptions{IPv4: ptr(true), IPv6: ptr(true)})
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(_ *server.InboundRequest, res server.ResponseSender) error {
		return res.SendResponse("AA")
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, listener.IsListening)

	for _, c := range []struct {
		host   string
		v4, v6 *bool
	}{
		{"127.0.0.1", ptr(true), ptr(false)},
		{"::1", ptr(false), ptr(true)},
	} {
		done := newEventWaiter()
		var ackOK atomic.Bool
		cli, _ := client.NewClient(client.ClientOptions{Host: c.host, IPv4: c.v4, IPv6: c.v6})
		outbound, err := cli.CreateConnection(client.ClientListenerOptions{Port: ptr(port)}, func(res *client.InboundResponse) error {
			ackOK.Store(res.GetMessage().Get("MSA.1").String() == "AA")
			done.signal()
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		waitFor(t, outbound.IsConnected)
		if err := outbound.SendMessage(makeTestMessage(t, "CONTROL_ID")); err != nil {
			t.Fatal(err)
		}
		done.wait(t, "ack from "+c.host)
		if !ackOK.Load() {
			t.Fatalf("MSA.1 was not AA from %s", c.host)
		}
		_ = outbound.Close()
		cli.CloseAll()
	}

	_ = listener.Close()
}

// TestDualStackHappyEyeballs mirrors "dual-stack client falls back from IPv6 to
// IPv4 via Happy Eyeballs": the server binds IPv4 only, the dual-stack client
// connects to localhost and must land on the IPv4 address.
func TestDualStackHappyEyeballs(t *testing.T) {
	port := freePort(t)
	done := newEventWaiter()
	var ackOK atomic.Bool

	srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(_ *server.InboundRequest, res server.ResponseSender) error {
		return res.SendResponse("AA")
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, listener.IsListening)

	cli, _ := client.NewClient(client.ClientOptions{Host: "localhost", IPv4: ptr(true), IPv6: ptr(true)})
	outbound, err := cli.CreateConnection(client.ClientListenerOptions{Port: ptr(port), MaxConnectionAttempts: ptr(2)}, func(res *client.InboundResponse) error {
		ackOK.Store(res.GetMessage().Get("MSA.1").String() == "AA")
		done.signal()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, outbound.IsConnected)
	if err := outbound.SendMessage(makeTestMessage(t, "CONTROL_ID")); err != nil {
		t.Fatal(err)
	}
	done.wait(t, "fallback ack")
	if !ackOK.Load() {
		t.Fatalf("MSA.1 was not AA after IPv6->IPv4 fallback")
	}

	_ = outbound.Close()
	_ = listener.Close()
	cli.CloseAll()
}

// runLoopback drives a single AA round-trip between a server bound to bindHost
// and a client dialing dialHost with the given families.
func runLoopback(t *testing.T, bindHost, dialHost string, srvV4, srvV6, cliV4, cliV6 *bool) {
	t.Helper()
	port := freePort(t)
	done := newEventWaiter()
	var ackOK atomic.Bool

	srv, err := server.NewServer(&server.ServerOptions{BindAddress: ptr(bindHost), IPv4: srvV4, IPv6: srvV6})
	if err != nil {
		t.Fatal(err)
	}
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(_ *server.InboundRequest, res server.ResponseSender) error {
		return res.SendResponse("AA")
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, listener.IsListening)

	cli, _ := client.NewClient(client.ClientOptions{Host: dialHost, IPv4: cliV4, IPv6: cliV6})
	outbound, err := cli.CreateConnection(client.ClientListenerOptions{Port: ptr(port)}, func(res *client.InboundResponse) error {
		ackOK.Store(res.GetMessage().Get("MSA.1").String() == "AA")
		done.signal()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, outbound.IsConnected)
	if err := outbound.SendMessage(makeTestMessage(t, "CONTROL_ID")); err != nil {
		t.Fatal(err)
	}
	done.wait(t, "loopback ack")
	if !ackOK.Load() {
		t.Fatalf("MSA.1 was not AA (%s -> %s)", dialHost, bindHost)
	}

	_ = outbound.Close()
	_ = listener.Close()
	cli.CloseAll()
}

// largeBase64 returns a base64 string encoding sizeKB kilobytes of data,
// standing in for the test generateLargeBase64String.
func largeBase64(sizeKB int) string {
	raw := strings.Repeat("X", sizeKB*1024)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}
