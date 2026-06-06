// MIT License
//
// Copyright (c) 2026 Shane
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package server_test

import (
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/client"
	"github.com/Bugs5382/go-hl7/client/utils"
	"github.com/Bugs5382/go-hl7/server"
)

// These mirror __tests__/server/hl7.issues.test.ts: the over-the-wire portions
// of PR #134 (getSocket), issue #130 (custom ACK), issue #132 (concurrency /
// split frame), and issue #133 (throughput).

// TestPR134GetSocketOverWire mirrors the createInbound handler receiving a real
// net.Conn through req.getSocket(), with addresses available on both ends.
func TestPR134GetSocketOverWire(t *testing.T) {
	port := freePort(t)
	done := newEventWaiter()
	var sockOK atomic.Bool

	srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(req *server.InboundRequest, res server.ResponseSender) error {
		sock := req.GetSocket()
		_, okLocal := sock.LocalAddr().(*net.TCPAddr)
		_, okRemote := sock.RemoteAddr().(*net.TCPAddr)
		sockOK.Store(okLocal && okRemote)
		return res.SendResponse("AA")
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, listener.IsListening)

	cli, _ := client.NewClient(client.ClientOptions{Host: "127.0.0.1", IPv4: ptr(true)})
	outbound, err := cli.CreateConnection(client.ClientListenerOptions{Port: ptr(port)}, func(*client.InboundResponse) error {
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
	done.wait(t, "ack")
	if !sockOK.Load() {
		t.Fatalf("socket did not expose TCP local/remote addresses")
	}

	_ = outbound.Close()
	_ = listener.Close()
	cli.CloseAll()
}

// TestIssue130CustomACK mirrors sendCustomResponse delivering a verbatim,
// vendor-shaped ACK (with extra MSA fields and an ERR segment) to the client.
func TestIssue130CustomACK(t *testing.T) {
	port := freePort(t)
	done := newEventWaiter()
	var (
		gotMSH3 atomic.Value
		gotMSH4 atomic.Value
		gotMSA1 atomic.Value
		gotMSA3 atomic.Value
		gotERR  atomic.Value
	)

	srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(req *server.InboundRequest, res server.ResponseSender) error {
		original := req.GetMessage()
		id := original.Get("MSH.10").String()
		date := utils.CreateHL7Date(time.Now(), "")
		customAck, err := builder.NewMessage(builder.MessageOptions{Text: joinSegs(
			`MSH|^~\&|CUSTOM_APP|CUSTOM_FAC|REMOTE_APP|REMOTE_FAC|`+date+`||ACK^A01|RESP_`+id+`|P|2.5`,
			`MSA|AA|`+id+`|Custom message accepted|||VENDOR_CODE`,
			`ERR|||0^Message accepted^HL70357|I`,
		)})
		if err != nil {
			return err
		}
		if err := res.SendCustomResponse(customAck); err != nil {
			return err
		}
		stored := res.GetAckMessage()
		if stored.Get("MSH.3").String() != "CUSTOM_APP" ||
			stored.Get("MSA.3").String() != "Custom message accepted" ||
			stored.Get("MSA.6").String() != "VENDOR_CODE" {
			t.Errorf("stored ACK mismatch: %q", stored.String())
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, listener.IsListening)

	cli, _ := client.NewClient(client.ClientOptions{Host: "127.0.0.1", IPv4: ptr(true)})
	outbound, err := cli.CreateConnection(client.ClientListenerOptions{Port: ptr(port)}, func(res *client.InboundResponse) error {
		got := res.GetMessage()
		gotMSH3.Store(got.Get("MSH.3").String())
		gotMSH4.Store(got.Get("MSH.4").String())
		gotMSA1.Store(got.Get("MSA.1").String())
		gotMSA3.Store(got.Get("MSA.3").String())
		gotERR.Store(got.Get("ERR.3.2").String())
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
	done.wait(t, "ack")

	checks := []struct {
		path, got, want string
	}{
		{"MSH.3", loadStr(&gotMSH3), "CUSTOM_APP"},
		{"MSH.4", loadStr(&gotMSH4), "CUSTOM_FAC"},
		{"MSA.1", loadStr(&gotMSA1), "AA"},
		{"MSA.3", loadStr(&gotMSA3), "Custom message accepted"},
		{"ERR.3.2", loadStr(&gotERR), "Message accepted"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Fatalf("%s = %q, want %q", c.path, c.got, c.want)
		}
	}

	_ = outbound.Close()
	_ = listener.Close()
	cli.CloseAll()
}

// TestIssue130CustomACKRawString mirrors sendCustomResponse accepting a raw HL7
// string as well as a Message.
func TestIssue130CustomACKRawString(t *testing.T) {
	port := freePort(t)
	done := newEventWaiter()
	var gotMSH3, gotMSA1 atomic.Value

	srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(req *server.InboundRequest, res server.ResponseSender) error {
		id := req.GetMessage().Get("MSH.10").String()
		date := utils.CreateHL7Date(time.Now(), "")
		raw := joinSegs(
			`MSH|^~\&|RAW|FAC|R|RF|`+date+`||ACK^A01|RAW_`+id+`|P|2.5`,
			`MSA|AA|`+id,
		)
		return res.SendCustomResponse(raw)
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, listener.IsListening)

	cli, _ := client.NewClient(client.ClientOptions{Host: "127.0.0.1", IPv4: ptr(true)})
	outbound, err := cli.CreateConnection(client.ClientListenerOptions{Port: ptr(port)}, func(res *client.InboundResponse) error {
		got := res.GetMessage()
		gotMSH3.Store(got.Get("MSH.3").String())
		gotMSA1.Store(got.Get("MSA.1").String())
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
	done.wait(t, "ack")
	if v, _ := gotMSH3.Load().(string); v != "RAW" {
		t.Fatalf("MSH.3 = %q, want RAW", v)
	}
	if v, _ := gotMSA1.Load().(string); v != "AA" {
		t.Fatalf("MSA.1 = %q, want AA", v)
	}

	_ = outbound.Close()
	_ = listener.Close()
	cli.CloseAll()
}

// TestIssue130CustomACKResponseSent mirrors sendCustomResponse emitting
// 'response.sent' on the listener.
func TestIssue130CustomACKResponseSent(t *testing.T) {
	port := freePort(t)
	responseSent := newEventWaiter()
	ackReceived := newEventWaiter()

	srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(req *server.InboundRequest, res server.ResponseSender) error {
		id := req.GetMessage().Get("MSH.10").String()
		date := utils.CreateHL7Date(time.Now(), "")
		ack, err := builder.NewMessage(builder.MessageOptions{Text: joinSegs(
			`MSH|^~\&|A|F|R|RF|`+date+`||ACK|X`+id+`|P|2.5`,
			`MSA|AA|`+id,
		)})
		if err != nil {
			return err
		}
		return res.SendCustomResponse(ack)
	})
	if err != nil {
		t.Fatal(err)
	}
	listener.On("response.sent", responseSent.signal)
	waitFor(t, listener.IsListening)

	cli, _ := client.NewClient(client.ClientOptions{Host: "127.0.0.1", IPv4: ptr(true)})
	outbound, err := cli.CreateConnection(client.ClientListenerOptions{Port: ptr(port)}, func(*client.InboundResponse) error {
		ackReceived.signal()
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, outbound.IsConnected)

	if err := outbound.SendMessage(makeTestMessage(t, "CONTROL_ID")); err != nil {
		t.Fatal(err)
	}
	responseSent.wait(t, "response.sent")
	ackReceived.wait(t, "ack")

	_ = outbound.Close()
	_ = listener.Close()
	cli.CloseAll()
}

// TestIssue132Concurrency mirrors two simultaneous clients not interleaving each
// other's per-socket data buffers.
func TestIssue132Concurrency(t *testing.T) {
	port := freePort(t)
	const expected = 6
	var mu sync.Mutex
	seen := map[string]struct{}{}
	allDone := newEventWaiter()
	var dataErrors atomic.Int32

	srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(req *server.InboundRequest, res server.ResponseSender) error {
		id := req.GetMessage().Get("MSH.10").String()
		mu.Lock()
		seen[id] = struct{}{}
		n := len(seen)
		mu.Unlock()
		if err := res.SendResponse("AA"); err != nil {
			return err
		}
		if n == expected {
			allDone.signal()
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	listener.On("data.error", func(...any) { dataErrors.Add(1) })
	waitFor(t, listener.IsListening)

	cliA, _ := client.NewClient(client.ClientOptions{Host: "127.0.0.1", IPv4: ptr(true)})
	cliB, _ := client.NewClient(client.ClientOptions{Host: "127.0.0.1", IPv4: ptr(true)})
	outA, err := cliA.CreateConnection(client.ClientListenerOptions{Port: ptr(port), WaitAck: ptr(false)}, func(*client.InboundResponse) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	outB, err := cliB.CreateConnection(client.ClientListenerOptions{Port: ptr(port), WaitAck: ptr(false)}, func(*client.InboundResponse) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, outA.IsConnected)
	waitFor(t, outB.IsConnected)

	var wg sync.WaitGroup
	sends := []struct {
		out *client.Connection
		id  string
	}{
		{outA, "A_1"}, {outB, "B_1"},
		{outA, "A_2"}, {outB, "B_2"},
		{outA, "A_3"}, {outB, "B_3"},
	}
	for _, s := range sends {
		wg.Add(1)
		go func(out *client.Connection, id string) {
			defer wg.Done()
			if err := out.SendMessage(makeTestMessage(t, id)); err != nil {
				t.Errorf("send %s: %v", id, err)
			}
		}(s.out, s.id)
	}
	wg.Wait()
	allDone.wait(t, "all messages")

	if dataErrors.Load() != 0 {
		t.Fatalf("data.error fired %d times", dataErrors.Load())
	}
	mu.Lock()
	defer mu.Unlock()
	for _, id := range []string{"A_1", "A_2", "A_3", "B_1", "B_2", "B_3"} {
		if _, ok := seen[id]; !ok {
			t.Fatalf("missing control id %q (seen=%v)", id, seen)
		}
	}

	_ = outA.Close()
	_ = outB.Close()
	_ = listener.Close()
	cliA.CloseAll()
	cliB.CloseAll()
}

// TestIssue132SplitFrame mirrors a single MLLP frame split across many tiny TCP
// writes still parsing cleanly.
func TestIssue132SplitFrame(t *testing.T) {
	port := freePort(t)
	ackReceived := make(chan string, 1)
	var dataErrors atomic.Int32

	srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(_ *server.InboundRequest, res server.ResponseSender) error {
		return res.SendResponse("AA")
	})
	if err != nil {
		t.Fatal(err)
	}
	listener.On("data.error", func(...any) { dataErrors.Add(1) })
	waitFor(t, listener.IsListening)

	big := make([]byte, 8*1024)
	for i := range big {
		big[i] = 'X'
	}
	bodyText := joinSegs(
		`MSH|^~\&|EPIC|HOSP|RECV|RFAC|20240101000000||ADT^A08|FRAG_001|P|2.5`,
		`EVN|A08|20240101000000`,
		`PID|1||MRN12345^^^HOSP^MR||DOE^JANE^A||19800101|F`,
		`OBX|1|TX|NOTE^Long Note^L||`+string(big)+`||||||F`,
	)
	const VT, FS, CR = 0x0b, 0x1c, 0x0d
	framed := append([]byte{VT}, append([]byte(bodyText), FS, CR)...)

	raw, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		buf := make([]byte, 4096)
		n, _ := raw.Read(buf)
		ackReceived <- string(buf[:n])
	}()

	const chunkSize = 64
	for i := 0; i < len(framed); i += chunkSize {
		end := i + chunkSize
		if end > len(framed) {
			end = len(framed)
		}
		if _, err := raw.Write(framed[i:end]); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond)
	}

	select {
	case ack := <-ackReceived:
		if !strings.Contains(ack, "MSA|AA|FRAG_001") {
			t.Fatalf("ack missing MSA|AA|FRAG_001: %q", ack)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for ack")
	}
	if dataErrors.Load() != 0 {
		t.Fatalf("data.error fired %d times", dataErrors.Load())
	}

	_ = raw.Close()
	_ = listener.Close()
}

// TestIssue133Throughput mirrors the listener processing a burst of messages
// without dropping any.
func TestIssue133Throughput(t *testing.T) {
	port := freePort(t)
	const total = 200
	var mu sync.Mutex
	seen := map[string]struct{}{}
	allDone := newEventWaiter()
	var dataErrors atomic.Int32

	srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(req *server.InboundRequest, res server.ResponseSender) error {
		id := req.GetMessage().Get("MSH.10").String()
		mu.Lock()
		seen[id] = struct{}{}
		n := len(seen)
		mu.Unlock()
		if err := res.SendResponse("AA"); err != nil {
			return err
		}
		if n == total {
			allDone.signal()
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	listener.On("data.error", func(...any) { dataErrors.Add(1) })
	waitFor(t, listener.IsListening)

	cli, _ := client.NewClient(client.ClientOptions{Host: "127.0.0.1", IPv4: ptr(true)})
	outbound, err := cli.CreateConnection(client.ClientListenerOptions{Port: ptr(port), WaitAck: ptr(false)}, func(*client.InboundResponse) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, outbound.IsConnected)

	start := time.Now()
	for i := 0; i < total; i++ {
		if err := outbound.SendMessage(makeTestMessage(t, "PERF_"+strconv.Itoa(i))); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}
	allDone.wait(t, "all messages")
	elapsed := time.Since(start)

	if dataErrors.Load() != 0 {
		t.Fatalf("data.error fired %d times", dataErrors.Load())
	}
	mu.Lock()
	got := len(seen)
	mu.Unlock()
	if got != total {
		t.Fatalf("seen = %d, want %d", got, total)
	}
	if listener.TotalMessage() < total {
		t.Fatalf("TotalMessage = %d, want >= %d", listener.TotalMessage(), total)
	}
	if elapsed > 10*time.Second {
		t.Fatalf("elapsed %v, want < 10s", elapsed)
	}

	_ = outbound.Close()
	_ = listener.Close()
	cli.CloseAll()
}

// joinSegs joins HL7 segment lines with the segment terminator \r.
func joinSegs(lines ...string) string {
	return strings.Join(lines, "\r")
}

// loadStr reads a string previously stored in an atomic.Value, or "" if unset.
func loadStr(v *atomic.Value) string {
	s, _ := v.Load().(string)
	return s
}
