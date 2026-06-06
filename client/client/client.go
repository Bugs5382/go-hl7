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

package client

import "sync"

// Client is the main entry point: it holds the remote-server connection
// defaults and opens per-port Connections. It mirrors the Client, which
// extends EventEmitter (here via the embedded eventEmitter); the "limitExceeded"
// event is re-emitted from connections.
type Client struct {
	eventEmitter
	// connections holds every Connection created off this client (the
	// _connections).
	connections []*Connection
	// opt is the validated client option set (the _opt).
	opt validatedClientOptions
	// stats tracks the aggregate counters the spec exposes (the stats).
	stats clientStats
	mu    sync.Mutex
}

// clientStats mirrors the Client.stats aggregate counters.
type clientStats struct {
	totalAck     int
	totalPending int
	totalSent    int
}

// NewClient creates a client to a remote server, validating the options.
// It mirrors the `new Client(properties)`; the spec throws on bad options, Go
// returns the error (the client tests catch these by message).
func NewClient(properties ClientOptions) (*Client, error) {
	opt, err := normalizeClientOptions(properties)
	if err != nil {
		return nil, err
	}
	return &Client{opt: opt}, nil
}

// CloseAll closes every connection and clears the list, mirroring the
// closeAll.
func (c *Client) CloseAll() {
	c.mu.Lock()
	conns := c.connections
	c.connections = nil
	c.mu.Unlock()
	for _, connection := range conns {
		_ = connection.Close()
	}
}

// CreateConnection opens a connection to a specified port, wiring the
// per-connection stat events back to the client aggregate counters and
// re-emitting limitExceeded. It mirrors the createConnection; the spec throws on
// bad per-port options, Go returns the error.
func (c *Client) CreateConnection(properties ClientListenerOptions, callback OutboundHandler) (*Connection, error) {
	outbound, err := newConnection(c, properties, callback)
	if err != nil {
		return nil, err
	}

	outbound.On("client.acknowledged", func(args ...any) {
		c.mu.Lock()
		c.stats.totalAck = args[0].(int)
		c.mu.Unlock()
	})
	outbound.On("client.sent", func(args ...any) {
		c.mu.Lock()
		c.stats.totalSent = args[0].(int)
		c.mu.Unlock()
	})
	outbound.On("client.pending", func(args ...any) {
		c.mu.Lock()
		c.stats.totalPending = args[0].(int)
		c.mu.Unlock()
	})
	outbound.On("client.limitExceeded", func(args ...any) {
		c.emit("limitExceeded", args...)
	})

	c.mu.Lock()
	c.connections = append(c.connections, outbound)
	c.mu.Unlock()

	return outbound, nil
}

// GetHost returns the configured host, mirroring the getHost.
func (c *Client) GetHost() string { return c.opt.host }

// TotalAck returns the lifetime acknowledged count, mirroring the totalAck.
func (c *Client) TotalAck() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stats.totalAck
}

// TotalPending returns the count of messages pending a (re)connection,
// mirroring the totalPending.
func (c *Client) TotalPending() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stats.totalPending
}

// TotalSent returns the lifetime sent count, mirroring the totalSent.
func (c *Client) TotalSent() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stats.totalSent
}
