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
	"bytes"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Bugs5382/go-hl7/server"
)

// sendRawFrame dials the listener, writes one MLLP frame carrying body, and
// returns the connection so the caller can keep it open until it is done.
func sendRawFrame(t *testing.T, port int, body []byte) net.Conn {
	t.Helper()
	conn, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		t.Fatal(err)
	}
	frame := append([]byte{0x0B}, body...)
	frame = append(frame, 0x1C, 0x0D)
	if _, err := conn.Write(frame); err != nil {
		_ = conn.Close()
		t.Fatal(err)
	}
	return conn
}

// receiveName starts a listener with the given charset, sends one framed ADT
// whose PID-5 is the raw bytes in nameBytes, and returns the decoded PID-5 the
// handler saw.
func receiveName(t *testing.T, charset string, nameBytes []byte) string {
	t.Helper()
	port := freePort(t)
	srv, err := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	if err != nil {
		t.Fatal(err)
	}
	got := make(chan string, 1)
	listener, err := srv.CreateInbound(server.ListenerOptions{Version: "2.7", Port: ptr(port), Charset: charset},
		func(req *server.InboundRequest, res server.ResponseSender) error {
			got <- req.GetMessage().Get("PID.5.1").String()
			return res.SendResponse("AA")
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	waitFor(t, listener.IsListening)

	body := append([]byte("MSH|^~\\&|A|B|C|D|20240101000000||ADT^A01|C1|P|2.7\rPID|1||123||"), nameBytes...)
	body = append(body, '\r')
	conn := sendRawFrame(t, port, body)
	defer func() { _ = conn.Close() }()

	select {
	case v := <-got:
		return v
	case <-time.After(5 * time.Second):
		t.Fatal("handler was not called")
		return ""
	}
}

// TestListenerCharsetISO88591 checks that a listener configured for
// ISO-8859-1 decodes a Latin-1 framed message (issue #50).
func TestListenerCharsetISO88591(t *testing.T) {
	latin1 := []byte{'M', 0xFC, 'l', 'l', 'e', 'r', '^', 'H', 'a', 'n', 's'} // Müller^Hans
	for _, charset := range []string{"iso-8859-1", "8859/1"} {
		if got := receiveName(t, charset, latin1); got != "Müller" {
			t.Errorf("charset %q: PID.5 = %q, want %q", charset, got, "Müller")
		}
	}
}

// TestListenerCharsetDefaultUTF8 checks that an empty Charset still decodes UTF-8.
func TestListenerCharsetDefaultUTF8(t *testing.T) {
	if got := receiveName(t, "", []byte("Müller^Hans")); got != "Müller" {
		t.Errorf("PID.5 = %q, want %q", got, "Müller")
	}
}

// TestListenerCharsetUnknownRejected checks that an unknown charset fails at
// listener creation, not at connect.
func TestListenerCharsetUnknownRejected(t *testing.T) {
	srv, err := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	if err != nil {
		t.Fatal(err)
	}
	in, err := srv.CreateInbound(server.ListenerOptions{Version: "2.7", Port: ptr(freePort(t)), Charset: "not-a-charset"},
		func(*server.InboundRequest, server.ResponseSender) error { return nil })
	if err == nil {
		_ = in.Close()
		t.Fatal("expected an error for an unknown charset")
	}
	if !strings.Contains(err.Error(), "not-a-charset") {
		t.Errorf("error %q should name the charset", err)
	}
}

// readAckBytes starts a listener with the given charset, sends one framed ADT
// whose MSH-3 is the raw bytes in appBytes, and returns the raw bytes of the
// ACK the listener wrote back (issue #61).
func readAckBytes(t *testing.T, charset string, appBytes []byte) []byte {
	t.Helper()
	port := freePort(t)
	srv, err := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	if err != nil {
		t.Fatal(err)
	}
	listener, err := srv.CreateInbound(server.ListenerOptions{Version: "2.7", Port: ptr(port), Charset: charset},
		func(req *server.InboundRequest, res server.ResponseSender) error {
			return res.SendResponse("AA")
		})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()
	waitFor(t, listener.IsListening)

	body := append([]byte("MSH|^~\\&|"), appBytes...)
	body = append(body, []byte("|B|C|D|20240101000000||ADT^A01|C1|P|2.7\rPID|1||123\r")...)
	conn := sendRawFrame(t, port, body)
	defer func() { _ = conn.Close() }()

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var got []byte
	buf := make([]byte, 4096)
	for !bytes.Contains(got, []byte{0x1C, 0x0D}) {
		n, err := conn.Read(buf)
		got = append(got, buf[:n]...)
		if err != nil {
			t.Fatalf("reading ACK: %v (got %q)", err, got)
		}
	}
	return got
}

// TestAckWrittenInListenerCharset checks that an ACK is encoded in the
// listener's charset, so a Latin-1 sender name round-trips (issue #61).
func TestAckWrittenInListenerCharset(t *testing.T) {
	got := readAckBytes(t, "iso-8859-1", []byte{'M', 0xFC, 'l', 'l', 'e', 'r'})
	if !bytes.Contains(got, []byte{'|', 'M', 0xFC, 'l', 'l', 'e', 'r', '|'}) {
		t.Errorf("ACK is not Latin-1 encoded: %q", got)
	}
	if bytes.Contains(got, []byte{0xC3, 0xBC}) {
		t.Errorf("ACK contains UTF-8 bytes: %q", got)
	}
}

// TestAckDefaultsToUTF8 checks that an empty Charset still writes UTF-8 ACKs.
func TestAckDefaultsToUTF8(t *testing.T) {
	got := readAckBytes(t, "", []byte("Müller"))
	if !bytes.Contains(got, []byte("|Müller|")) {
		t.Errorf("ACK is not UTF-8 encoded: %q", got)
	}
}
