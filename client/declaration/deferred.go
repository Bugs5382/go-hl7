package declaration

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

// Deferred is a one-shot promise handle, mirroring the Deferred<T>
// (declaration/deferred.ts: { promise, resolve, reject }). Go has no Promise,
// so Resolve/Reject signal completion exactly once and Wait blocks until then,
// returning the resolved value or the rejection error. It is the connection's
// _onConnect gate.
type Deferred[T any] struct {
	mu     sync.Mutex
	done   chan struct{}
	value  T
	err    error
	closed bool
}

// NewDeferred constructs a fresh, unsettled Deferred.
func NewDeferred[T any]() *Deferred[T] {
	return &Deferred[T]{done: make(chan struct{})}
}

// Resolve completes the deferred with value. Subsequent calls are no-ops,
// matching a settled Promise.
func (d *Deferred[T]) Resolve(value T) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return
	}
	d.value = value
	d.closed = true
	close(d.done)
}

// Reject completes the deferred with reason. Subsequent calls are no-ops.
func (d *Deferred[T]) Reject(reason error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.closed {
		return
	}
	d.err = reason
	d.closed = true
	close(d.done)
}

// Wait blocks until the deferred is settled and returns its value/error,
// standing in for `await dfd.promise`.
func (d *Deferred[T]) Wait() (T, error) {
	<-d.done
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.value, d.err
}

// Done exposes the completion channel so callers can select on settlement.
func (d *Deferred[T]) Done() <-chan struct{} { return d.done }
