// Tests for the reconnect retry budget (issues #55 and #56).
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

// refusedPort returns a loopback port with nothing listening on it.
func refusedPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}

// retryClient builds a client with a fast backoff and one auto-connecting
// connection on port, counting client.error and close events.
func retryClient(t *testing.T, port, maxAttempts int) (errs, closes *int32, done chan struct{}) {
	t.Helper()
	cl, err := NewClient(ClientOptions{
		Version:           "2.7",
		Host:              "127.0.0.1",
		IPv4:              ptr(true),
		ConnectionTimeout: ptr(5000),
		RetryLow:          ptr(1),
		RetryHigh:         ptr(5),
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(cl.CloseAll)

	errs, closes = new(int32), new(int32)
	done = make(chan struct{})
	conn, err := cl.CreateConnection(
		ClientListenerOptions{AutoConnect: ptr(false), Port: ptr(port), MaxConnectionAttempts: ptr(maxAttempts)},
		func(*InboundResponse) error { return nil },
	)
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	conn.On("client.error", func(...any) { atomic.AddInt32(errs, 1) })
	conn.On("close", func(...any) {
		atomic.AddInt32(closes, 1)
		close(done)
	})
	// An auto-connecting connection would dial before the listeners above are
	// attached, so start it by hand once they are.
	conn.mu.Lock()
	conn.readyState = declaration.Connecting
	conn.mu.Unlock()
	conn.connect()
	return errs, closes, done
}

// Issue #55: after the last allowed retry the connection must close. With a
// refused port and MaxConnectionAttempts 2 there are two client.error events
// and then a close.
func TestLastRetryClosesConnection(t *testing.T) {
	errs, closes, done := retryClient(t, refusedPort(t), 2)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("no close event after the last retry; client.error=%d", atomic.LoadInt32(errs))
	}
	if got := atomic.LoadInt32(errs); got != 2 {
		t.Fatalf("client.error count = %d, want 2", got)
	}
	if got := atomic.LoadInt32(closes); got != 1 {
		t.Fatalf("close count = %d, want 1", got)
	}
}

// Issue #56: a successful connect resets the retry count, so after the listener
// goes away every outage gets the full MaxConnectionAttempts budget.
func TestRetryCountResetsAfterConnect(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	accepted := make(chan struct{}, 1)
	go func() {
		c, err := l.Accept()
		if err != nil {
			return
		}
		accepted <- struct{}{}
		// Drop the first connection and stop listening so every redial is refused.
		_ = l.Close()
		_ = c.Close()
	}()

	errs, _, done := retryClient(t, port, 3)

	select {
	case <-accepted:
	case <-time.After(5 * time.Second):
		t.Fatalf("listener never accepted the first connection")
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("no close event after the outage; client.error=%d", atomic.LoadInt32(errs))
	}
	if got := atomic.LoadInt32(errs); got != 3 {
		t.Fatalf("client.error count = %d, want 3 (the full budget)", got)
	}
}
