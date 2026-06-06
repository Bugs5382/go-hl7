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

// Package server implements the HL7 MLLP server: Server, Inbound,
// InboundRequest, SendResponse, and option normalization. It is a thin
// TCP/TLS listener over the client package's Message parser and MLLP codec,
// building the AA/AE/AR + CA/CR/CE ACK matrix.
package server

import "sync"

// eventEmitter is the server package's minimal On/Once/emit EventEmitter, which
// Server, Inbound, and SendResponse all embed. Same documented Go adaptation as
// the client package: string event names, handlers take a variadic []any, and
// it is concurrency-safe so per-socket data goroutines and user subscriptions
// can race.
type eventEmitter struct {
	mu        sync.Mutex
	listeners map[string][]*emitterHandler
}

type emitterHandler struct {
	fn   func(args ...any)
	once bool
}

// On registers handler for the named event (the on).
func (e *eventEmitter) On(name string, handler func(args ...any)) *eventEmitter {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.listeners == nil {
		e.listeners = map[string][]*emitterHandler{}
	}
	e.listeners[name] = append(e.listeners[name], &emitterHandler{fn: handler})
	return e
}

// Once registers a one-shot handler (the once).
func (e *eventEmitter) Once(name string, handler func(args ...any)) *eventEmitter {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.listeners == nil {
		e.listeners = map[string][]*emitterHandler{}
	}
	e.listeners[name] = append(e.listeners[name], &emitterHandler{fn: handler, once: true})
	return e
}

// RemoveAllListeners drops all handlers, or those for a single event (the
// removeAllListeners).
func (e *eventEmitter) RemoveAllListeners(name ...string) *eventEmitter {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(name) == 0 {
		e.listeners = nil
		return e
	}
	if e.listeners != nil {
		delete(e.listeners, name[0])
	}
	return e
}

// emit fires every handler for the named event with args, stripping one-shots
// (the emit). Handlers run unlocked so re-entrant subscription is safe.
func (e *eventEmitter) emit(name string, args ...any) bool {
	e.mu.Lock()
	if e.listeners == nil {
		e.mu.Unlock()
		return false
	}
	handlers := e.listeners[name]
	if len(handlers) == 0 {
		e.mu.Unlock()
		return false
	}
	snapshot := make([]*emitterHandler, len(handlers))
	copy(snapshot, handlers)
	remaining := handlers[:0:0]
	for _, h := range handlers {
		if !h.once {
			remaining = append(remaining, h)
		}
	}
	if len(remaining) == 0 {
		delete(e.listeners, name)
	} else {
		e.listeners[name] = remaining
	}
	e.mu.Unlock()

	for _, h := range snapshot {
		h.fn(args...)
	}
	return true
}
