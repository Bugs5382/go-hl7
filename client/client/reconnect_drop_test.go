// Tests for reconnect after a drop and the connection timer (issues #57 and #58).
package client

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
	"sync/atomic"
	"testing"
	"time"

	"github.com/Bugs5382/go-hl7/client/internal/declaration"
)

// reconnectHarness is a loopback listener that counts accepts, paired with a
// connection that has not dialed yet so listeners can be attached first.
type reconnectHarness struct {
	conn     *Connection
	accepts  int32
	errors   int32
	timeouts int32
	closes   int32
	listener net.Listener
}

// newReconnectHarness listens on a loopback port and builds a connection to it
// with the given ConnectionTimeout (nil keeps the default 0) and a fast backoff.
func newReconnectHarness(t *testing.T, connectionTimeout *int, dropFirst bool) *reconnectHarness {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = l.Close() })
	h := &reconnectHarness{listener: l}

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			n := atomic.AddInt32(&h.accepts, 1)
			if dropFirst && n == 1 {
				_ = c.Close()
				continue
			}
			t.Cleanup(func() { _ = c.Close() })
		}
	}()

	cl, err := NewClient(ClientOptions{
		Version:           "2.7",
		Host:              "127.0.0.1",
		IPv4:              ptr(true),
		ConnectionTimeout: connectionTimeout,
		RetryLow:          ptr(1),
		RetryHigh:         ptr(5),
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(cl.CloseAll)

	conn, err := cl.CreateConnection(
		ClientListenerOptions{AutoConnect: ptr(false), Port: ptr(l.Addr().(*net.TCPAddr).Port)},
		func(*InboundResponse) error { return nil },
	)
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	conn.On("client.error", func(...any) { atomic.AddInt32(&h.errors, 1) })
	conn.On("client.timeout", func(...any) { atomic.AddInt32(&h.timeouts, 1) })
	conn.On("close", func(...any) { atomic.AddInt32(&h.closes, 1) })
	h.conn = conn
	return h
}

// start dials once the listeners are attached.
func (h *reconnectHarness) start() {
	h.conn.mu.Lock()
	h.conn.readyState = declaration.Connecting
	h.conn.mu.Unlock()
	h.conn.connect()
}

// waitUntil polls cond for up to two seconds.
func waitUntil(cond func() bool) bool {
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return cond()
}

// Issue #57: with the default ConnectionTimeout of 0 a dropped connection emits
// client.error and reconnects.
func TestDropReconnectsWithDefaultConnectionTimeout(t *testing.T) {
	h := newReconnectHarness(t, nil, true)
	h.start()

	if !waitUntil(func() bool { return atomic.LoadInt32(&h.accepts) >= 2 }) {
		t.Fatalf("no reconnect after a drop; accepts=%d", atomic.LoadInt32(&h.accepts))
	}
	if atomic.LoadInt32(&h.errors) < 1 {
		t.Fatalf("expected a client.error for the drop, got %d", atomic.LoadInt32(&h.errors))
	}
}

// Issue #57: a refused port with ConnectionTimeout 0 keeps retrying instead of
// stopping after one error.
func TestRefusedPortRetriesWithDefaultConnectionTimeout(t *testing.T) {
	h := newReconnectHarness(t, nil, false)
	_ = h.listener.Close()
	h.start()

	if !waitUntil(func() bool { return atomic.LoadInt32(&h.errors) >= 3 }) {
		t.Fatalf("expected repeated client.error events, got %d", atomic.LoadInt32(&h.errors))
	}
}

// Issue #57: an explicit Close() does not reconnect.
func TestCloseDoesNotReconnect(t *testing.T) {
	h := newReconnectHarness(t, nil, false)
	h.start()
	if !waitUntil(func() bool { return atomic.LoadInt32(&h.accepts) >= 1 }) {
		t.Fatalf("listener never accepted the connection")
	}
	if !waitUntil(func() bool { return h.conn.IsConnected() }) {
		t.Fatalf("connection never became connected")
	}

	_ = h.conn.Close()
	time.Sleep(300 * time.Millisecond)

	if got := atomic.LoadInt32(&h.accepts); got != 1 {
		t.Fatalf("accepts = %d, want 1 (no reconnect after Close)", got)
	}
	if got := atomic.LoadInt32(&h.errors); got != 0 {
		t.Fatalf("client.error = %d, want 0 after an explicit Close", got)
	}
}

// Issue #58: ConnectionTimeout limits one attempt only. A healthy open
// connection is left alone: no timeout, no error, no redial.
func TestOpenConnectionIsLeftAlone(t *testing.T) {
	h := newReconnectHarness(t, ptr(200), false)
	h.start()

	time.Sleep(900 * time.Millisecond)

	if got := atomic.LoadInt32(&h.timeouts); got != 0 {
		t.Fatalf("client.timeout = %d, want 0 on an open connection", got)
	}
	if got := atomic.LoadInt32(&h.errors); got != 0 {
		t.Fatalf("client.error = %d, want 0 on an open connection", got)
	}
	if got := atomic.LoadInt32(&h.accepts); got != 1 {
		t.Fatalf("accepts = %d, want 1 (no redial)", got)
	}
}
