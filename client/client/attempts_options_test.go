// Tests for unlimited retries and client-level attempt options (issues #59 and #60).
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
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Bugs5382/go-hl7/client/internal/declaration"
)

// startOnRefusedPort builds a client from opts, adds a connection with lo to a
// refused port, counts its client.error and close events and starts it once the
// listeners are attached.
func startOnRefusedPort(t *testing.T, opts ClientOptions, lo ClientListenerOptions) (errs, closes *int32, done chan struct{}) {
	t.Helper()
	opts.Version, opts.Host, opts.IPv4 = "2.7", "127.0.0.1", ptr(true)
	opts.RetryLow, opts.RetryHigh = ptr(1), ptr(5)
	cl, err := NewClient(opts)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(cl.CloseAll)

	lo.AutoConnect, lo.Port = ptr(false), ptr(refusedPort(t))
	conn, err := cl.CreateConnection(lo, func(*InboundResponse) error { return nil })
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	errs, closes, done = new(int32), new(int32), make(chan struct{})
	conn.On("client.error", func(...any) { atomic.AddInt32(errs, 1) })
	conn.On("close", func(...any) {
		atomic.AddInt32(closes, 1)
		close(done)
	})
	conn.mu.Lock()
	conn.readyState = declaration.Connecting
	conn.mu.Unlock()
	conn.connect()
	return errs, closes, done
}

// Issue #59: MaxConnectionAttempts 0 retries forever, so the connection keeps
// emitting client.error and never closes.
func TestUnlimitedConnectionAttempts(t *testing.T) {
	errs, closes, _ := startOnRefusedPort(t, ClientOptions{}, ClientListenerOptions{MaxConnectionAttempts: ptr(0)})

	// Fifteen errors is past the default cap of ten.
	if !waitUntil(func() bool { return atomic.LoadInt32(errs) >= 15 }) {
		t.Fatalf("expected the retries to continue, got %d client.error events", atomic.LoadInt32(errs))
	}
	if got := atomic.LoadInt32(closes); got != 0 {
		t.Fatalf("close count = %d, want 0 with unlimited retries", got)
	}
}

// Issue #59: the accepted range is 0..50.
func TestMaxConnectionAttemptsRange(t *testing.T) {
	cl, err := NewClient(ClientOptions{Version: "2.7", Host: "127.0.0.1", IPv4: ptr(true)})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(cl.CloseAll)
	for _, tc := range []struct {
		v  int
		ok bool
	}{{0, true}, {1, true}, {50, true}, {51, false}, {-1, false}} {
		_, err := cl.CreateConnection(
			ClientListenerOptions{AutoConnect: ptr(false), Port: ptr(12_400), MaxConnectionAttempts: ptr(tc.v)},
			func(*InboundResponse) error { return nil },
		)
		if tc.ok && err != nil {
			t.Fatalf("MaxConnectionAttempts %d: unexpected error %v", tc.v, err)
		}
		if !tc.ok && (err == nil || !strings.Contains(err.Error(), "maxConnectionAttempts must be a number (0, 50).")) {
			t.Fatalf("MaxConnectionAttempts %d: expected the (0, 50) range error, got %v", tc.v, err)
		}
	}
}

// Issue #60: MaxConnectionAttempts set at client level is the default for every
// connection that does not set its own.
func TestClientLevelMaxConnectionAttempts(t *testing.T) {
	errs, _, done := startOnRefusedPort(t, ClientOptions{MaxConnectionAttempts: ptr(3)}, ClientListenerOptions{})
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("no close event; client.error=%d", atomic.LoadInt32(errs))
	}
	if got := atomic.LoadInt32(errs); got != 3 {
		t.Fatalf("client.error count = %d, want 3", got)
	}
}

// Issue #60: a per-connection value still overrides the client-level one.
func TestConnectionMaxConnectionAttemptsOverridesClient(t *testing.T) {
	errs, _, done := startOnRefusedPort(t, ClientOptions{MaxConnectionAttempts: ptr(5)}, ClientListenerOptions{MaxConnectionAttempts: ptr(2)})
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("no close event; client.error=%d", atomic.LoadInt32(errs))
	}
	if got := atomic.LoadInt32(errs); got != 2 {
		t.Fatalf("client.error count = %d, want 2", got)
	}
}

// Issue #60: client-level values are range checked, and MaxAttempts falls back
// to the client value per connection.
func TestClientLevelAttemptOptions(t *testing.T) {
	base := ClientOptions{Version: "2.7", Host: "127.0.0.1", IPv4: ptr(true)}

	bad := base
	bad.MaxAttempts = ptr(0)
	if _, err := NewClient(bad); err == nil || !strings.Contains(err.Error(), "maxAttempts must be a number (1, 50).") {
		t.Fatalf("MaxAttempts 0: expected the (1, 50) range error, got %v", err)
	}
	bad = base
	bad.MaxConnectionAttempts = ptr(51)
	if _, err := NewClient(bad); err == nil || !strings.Contains(err.Error(), "maxConnectionAttempts must be a number (0, 50).") {
		t.Fatalf("MaxConnectionAttempts 51: expected the (0, 50) range error, got %v", err)
	}

	unlimited := base
	unlimited.MaxConnectionAttempts = ptr(0)
	if _, err := NewClient(unlimited); err != nil {
		t.Fatalf("client-level MaxConnectionAttempts 0: unexpected error %v", err)
	}

	withAttempts := base
	withAttempts.MaxAttempts = ptr(7)
	cl, err := NewClient(withAttempts)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(cl.CloseAll)
	conn, err := cl.CreateConnection(
		ClientListenerOptions{AutoConnect: ptr(false), Port: ptr(12_401)},
		func(*InboundResponse) error { return nil },
	)
	if err != nil {
		t.Fatalf("CreateConnection: %v", err)
	}
	if got := conn.opt.maxAttempts; got != 7 {
		t.Fatalf("connection maxAttempts = %d, want the client-level 7", got)
	}
}
