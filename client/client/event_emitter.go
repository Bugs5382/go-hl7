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

import "sync"

// eventEmitter is a minimal Go EventEmitter, which both Client and Connection
// embed. It is a minimal On(event, handler) / Once / emit over string event
// names (documented adaptation per spec 2.8). Handlers take a variadic []any
// argument list so the same event set maps directly (numbers, errors, or no
// args). It is safe for concurrent use because the socket data/close callbacks
// fire from the connection's read goroutine while user code subscribes from
// another.
type eventEmitter struct {
	mu        sync.Mutex
	listeners map[string][]*emitterHandler
}

// emitterHandler wraps a registered callback and whether it is a one-shot
// (the once) listener.
type emitterHandler struct {
	fn   func(args ...any)
	once bool
}

// On registers handler for the named event, mirroring the
// EventEmitter.on. It returns the emitter so registrations can be chained the
// way the spec returns `this`.
func (e *eventEmitter) On(name string, handler func(args ...any)) *eventEmitter {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.listeners == nil {
		e.listeners = map[string][]*emitterHandler{}
	}
	e.listeners[name] = append(e.listeners[name], &emitterHandler{fn: handler})
	return e
}

// Once registers a one-shot handler that is removed after its first
// invocation, mirroring the EventEmitter.once.
func (e *eventEmitter) Once(name string, handler func(args ...any)) *eventEmitter {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.listeners == nil {
		e.listeners = map[string][]*emitterHandler{}
	}
	e.listeners[name] = append(e.listeners[name], &emitterHandler{fn: handler, once: true})
	return e
}

// RemoveAllListeners drops every registered handler (or only those for the
// given event when a name is supplied), mirroring the
// EventEmitter.removeAllListeners.
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

// emit fires every handler registered for the named event with args, removing
// one-shot handlers afterward. It mirrors the EventEmitter.emit. Handlers
// run with the emitter unlocked so a handler may re-subscribe (the end2end
// "close" handler registers a follow-on "connection" listener).
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
	// Snapshot the current handlers and strip the one-shots before releasing
	// the lock so re-entrant emit/On calls behave like the once semantics.
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
