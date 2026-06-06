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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/client"
	"github.com/Bugs5382/go-hl7/server"
)

// These tests mirror the client/server sanity + no-tls/tls blocks of
// __tests__/client/hl7.end2end.test.ts and __tests__/server/hl7.end2end.test.ts:
// a real Client talks to a real Server over an ephemeral localhost port.

func ptr[T any](v T) *T { return &v }

// makeTestMessage builds the minimal ADT^A01 used across the end2end suites.
func makeTestMessage(t *testing.T, controlID string) *builder.Message {
	t.Helper()
	m, err := builder.NewMessage(builder.MessageOptions{
		Text: "MSH|^~\\&|||||20240101000000||ADT^A01|" + controlID + "|D|2.7",
	})
	if err != nil {
		t.Fatalf("build message: %v", err)
	}
	return m
}

// freePort grabs an ephemeral localhost port and releases it for the server to
// rebind (the Go equivalent of portfinder.getPortPromise).
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}

func TestEnd2EndSimpleConnect(t *testing.T) {
	port := freePort(t)

	srv, err := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	if err != nil {
		t.Fatal(err)
	}
	var gotVersion atomic.Value
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(req *server.InboundRequest, res server.ResponseSender) error {
		gotVersion.Store(req.GetMessage().Get("MSH.12").String())
		return res.SendResponse("AA")
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, listener.IsListening)

	cli, err := client.NewClient(client.ClientOptions{Host: "127.0.0.1", IPv4: ptr(true)})
	if err != nil {
		t.Fatal(err)
	}
	var ackOK atomic.Bool
	done := newEventWaiter()
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
		t.Fatalf("send: %v", err)
	}
	done.wait(t, "ack")

	if !ackOK.Load() {
		t.Fatalf("MSA.1 was not AA")
	}
	if v, _ := gotVersion.Load().(string); v != "2.7" {
		t.Fatalf("server saw MSH.12 = %q, want 2.7", v)
	}
	if cli.TotalSent() != 1 {
		t.Fatalf("TotalSent = %d, want 1", cli.TotalSent())
	}
	// allow the acknowledged counter to settle
	waitFor(t, func() bool { return cli.TotalAck() == 1 })

	_ = outbound.Close()
	_ = listener.Close()
	cli.CloseAll()
}

func TestEnd2EndSendTwiceNoAck(t *testing.T) {
	port := freePort(t)

	srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
	var totalSent atomic.Int32
	listener, err := srv.CreateInbound(server.ListenerOptions{Port: ptr(port)}, func(req *server.InboundRequest, res server.ResponseSender) error {
		totalSent.Add(1)
		return res.SendResponse("AA")
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, listener.IsListening)

	cli, _ := client.NewClient(client.ClientOptions{Host: "127.0.0.1", IPv4: ptr(true)})
	outbound, err := cli.CreateConnection(client.ClientListenerOptions{Port: ptr(port), WaitAck: ptr(false)}, func(*client.InboundResponse) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, outbound.IsConnected)

	if err := outbound.SendMessage(makeTestMessage(t, "CONTROL_ID_1")); err != nil {
		t.Fatal(err)
	}
	if err := outbound.SendMessage(makeTestMessage(t, "CONTROL_ID_2")); err != nil {
		t.Fatal(err)
	}

	waitFor(t, func() bool { return totalSent.Load() == 2 })
	waitFor(t, func() bool { return cli.TotalSent() == 2 })
	waitFor(t, func() bool { return cli.TotalAck() == 2 })

	_ = outbound.Close()
	_ = listener.Close()
	cli.CloseAll()
}

func TestEnd2EndQueueAutoConnectFalse(t *testing.T) {
	cli, _ := client.NewClient(client.ClientOptions{Host: "127.0.0.1", IPv4: ptr(true)})
	// autoConnect false + a port with no server: the message must queue rather
	// than send. (mocks _connect; here the dial just fails in the
	// background and the message stays queued as pending.)
	outbound, err := cli.CreateConnection(
		client.ClientListenerOptions{AutoConnect: ptr(false), Port: ptr(9), MaxConnectionAttempts: ptr(1)},
		func(*client.InboundResponse) error { return nil },
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := outbound.SendMessage(makeTestMessage(t, "CONTROL_ID")); err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { return cli.TotalPending() == 1 })

	_ = outbound.Close()
	cli.CloseAll()
}

// waitFor polls cond until true or the deadline, mirroring the implicit
// settling the spec gets from its event loop / dfd.promise awaits.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("condition not met before deadline")
}

// eventWaiter is a tiny once-style latch for the On(event) wait helpers.
type eventWaiter struct {
	once sync.Once
	ch   chan struct{}
}

func newEventWaiter() *eventWaiter { return &eventWaiter{ch: make(chan struct{})} }

func (w *eventWaiter) signal(_ ...any) { w.once.Do(func() { close(w.ch) }) }

func (w *eventWaiter) wait(t *testing.T, what string) {
	t.Helper()
	select {
	case <-w.ch:
	case <-time.After(10 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}
