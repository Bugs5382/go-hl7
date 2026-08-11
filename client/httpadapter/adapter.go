// Package httpadapter is an optional, dependency-free bridge that lets a
// standard net/http application send HL7 messages over MLLP and expose the
// result to HTTP callers. It is the Go analog of the node fastify-hl7 plugin:
// where that plugin decorates a Fastify instance with an HL7 client, this
// package hands a net/http app a small set of http.Handler values wired to a
// go-hl7 outbound Connection.
//
// The adapter is intentionally minimal. It holds a bound sender, exposes a
// POST handler that forwards a request body as an HL7 message and returns the
// ACK, and a GET handler that reports connection health. It adds no third-party
// dependency and nothing else in the module imports it.
//
// Wiring (the ack handler must be registered on the Connection at creation, so
// the adapter is created first and bound after):
//
//	a := httpadapter.New(httpadapter.WithTimeout(5 * time.Second))
//	conn, err := c.CreateConnection(
//		client.ClientListenerOptions{Port: ptr(3000)},
//		a.AckHandler(), // feed ACKs back to the adapter
//	)
//	if err != nil {
//		// handle
//	}
//	a.Bind(conn)
//
//	mux := http.NewServeMux()
//	mux.Handle("POST /hl7", a.SendHandler())
//	mux.Handle("GET /healthz", a.HealthHandler())
//
// Note on parity: fastify-hl7 also decorates the app with message/batch
// builders and inbound-server management. In Go those needs are met by
// importing client/hl7, client/builder, and server directly, so this adapter
// deliberately does not re-export them.
package httpadapter

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
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/client"
)

// ContentTypeHL7 is the media type used for HL7 v2 (ER7) request and response
// bodies.
const ContentTypeHL7 = "application/hl7-v2"

// defaultTimeout bounds how long SendHandler waits for an ACK before giving up.
const defaultTimeout = 10 * time.Second

var (
	// ErrNotBound is returned when a message is sent before Bind has supplied a
	// sender.
	ErrNotBound = errors.New("httpadapter: adapter is not bound to a connection")
	// ErrAckTimeout is returned when no ACK arrives within the configured
	// timeout.
	ErrAckTimeout = errors.New("httpadapter: timed out waiting for ACK")
)

// Sender is the minimal subset of *client.Connection the adapter needs. It is
// satisfied by *client.Connection and can be faked in tests.
type Sender interface {
	// SendMessage transmits an HL7 message over the established connection.
	SendMessage(message client.MessageItem) error
	// IsConnected reports whether the underlying socket is currently up.
	IsConnected() bool
	// GetPort returns the remote port the connection targets.
	GetPort() int
}

// Adapter bridges net/http and a go-hl7 outbound Connection. It serializes
// sends so a single in-flight message is correlated with the next ACK. The
// zero value is not usable; construct one with New.
type Adapter struct {
	mu      sync.Mutex // serializes send/ack correlation
	sender  Sender
	acks    chan *client.InboundResponse
	timeout time.Duration
}

// Option configures an Adapter.
type Option func(*Adapter)

// WithTimeout sets how long SendHandler waits for an ACK. A non-positive value
// is ignored and the default (10s) is kept.
func WithTimeout(d time.Duration) Option {
	return func(a *Adapter) {
		if d > 0 {
			a.timeout = d
		}
	}
}

// New creates an Adapter. Register AckHandler on the Connection when creating
// it, then call Bind with that Connection.
func New(opts ...Option) *Adapter {
	a := &Adapter{
		// Buffered by one so AckHandler never blocks the connection read loop
		// and a synchronous ACK can land before the sender's caller reads it.
		acks:    make(chan *client.InboundResponse, 1),
		timeout: defaultTimeout,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Bind attaches the sender (normally the *client.Connection) that SendHandler
// transmits through. It may be called again to rebind.
func (a *Adapter) Bind(sender Sender) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sender = sender
}

// AckHandler returns a client.OutboundHandler to register when creating the
// Connection. It forwards each ACK to the waiting SendHandler; if no request is
// waiting the ACK is dropped rather than blocking the connection.
func (a *Adapter) AckHandler() client.OutboundHandler {
	return func(res *client.InboundResponse) error {
		select {
		case a.acks <- res:
		default:
		}
		return nil
	}
}

// exchange sends msg and waits for the matching ACK, honoring ctx and the
// configured timeout. It holds the lock for the whole round trip so only one
// message is in flight at a time.
func (a *Adapter) exchange(ctx context.Context, msg client.MessageItem) (*client.InboundResponse, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.sender == nil {
		return nil, ErrNotBound
	}

	// Drop any stale ACK left over from a prior timed-out request.
	select {
	case <-a.acks:
	default:
	}

	if err := a.sender.SendMessage(msg); err != nil {
		return nil, err
	}

	timer := time.NewTimer(a.timeout)
	defer timer.Stop()

	select {
	case res := <-a.acks:
		return res, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return nil, ErrAckTimeout
	}
}

// SendHandler returns an http.Handler that reads an HL7 v2 message from the
// request body, sends it over MLLP, and writes the ACK back as the response
// body. It responds 405 to non-POST, 400 on an unparseable body, 503 when the
// adapter is not bound, 504 on ACK timeout, and 502 on any other send failure.
func (a *Adapter) SendHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		msg, err := builder.NewMessage(builder.MessageOptions{Text: string(body)})
		if err != nil {
			http.Error(w, "invalid HL7 message: "+err.Error(), http.StatusBadRequest)
			return
		}

		res, err := a.exchange(r.Context(), msg)
		if err != nil {
			switch {
			case errors.Is(err, ErrNotBound):
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
			case errors.Is(err, ErrAckTimeout), errors.Is(err, context.DeadlineExceeded):
				http.Error(w, err.Error(), http.StatusGatewayTimeout)
			case errors.Is(err, context.Canceled):
				// Client went away; nothing useful to send.
				return
			default:
				http.Error(w, "failed to send HL7 message: "+err.Error(), http.StatusBadGateway)
			}
			return
		}

		w.Header().Set("Content-Type", ContentTypeHL7)
		_, _ = io.WriteString(w, res.GetMessage().String())
	})
}

// health is the JSON shape reported by HealthHandler.
type health struct {
	Connected bool `json:"connected"`
	Port      int  `json:"port"`
}

// HealthHandler returns an http.Handler that reports connection status as JSON.
// It responds 200 when the connection is up and 503 otherwise (including when
// the adapter is not yet bound).
func (a *Adapter) HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		a.mu.Lock()
		sender := a.sender
		a.mu.Unlock()

		body := health{}
		if sender != nil {
			body.Connected = sender.IsConnected()
			body.Port = sender.GetPort()
		}

		w.Header().Set("Content-Type", "application/json")
		if !body.Connected {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(body)
	})
}
