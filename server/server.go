package server

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

import "github.com/Bugs5382/go-hl7/server/utils"

// Re-export the option/error types so callers use server.ServerOptions,
// server.ListenerOptions, server.HL7ListenerError, etc., mirroring the
// barrel that re-exports them from the server package root.
type (
	// ServerOptions configures a Server (see utils.ServerOptions).
	ServerOptions = utils.ServerOptions
	// ListenerOptions configures a createInbound listener (see
	// utils.ListenerOptions).
	ListenerOptions = utils.ListenerOptions
	// TLSConfig configures server TLS (see utils.TLSConfig).
	TLSConfig = utils.TLSConfig
	// MSHOverride is an ACK MSH-field override (see utils.MSHOverride).
	MSHOverride = utils.MSHOverride
	// HL7ServerError is the server error type (see utils.HL7ServerError).
	HL7ServerError = utils.HL7ServerError
	// HL7ListenerError is the listener error type (see utils.HL7ListenerError).
	HL7ListenerError = utils.HL7ListenerError
)

// StringOverride builds a literal MSH override (see utils.StringOverride).
var StringOverride = utils.StringOverride

// FuncOverride builds a computed MSH override (see utils.FuncOverride).
var FuncOverride = utils.FuncOverride

// Server starts listeners on network addresses. It mirrors the reference's
// Server (an EventEmitter) via the embedded EventEmitter.
type Server struct {
	EventEmitter
	opt utils.NormalizedServerOptions
}

// NewServer constructs a Server, validating the options. It mirrors the
// `new Server(properties)`; the spec throws on bad options, Go returns the error.
// Pass nil for the defaults (IPv4-only on 0.0.0.0).
func NewServer(properties *ServerOptions) (*Server, error) {
	opt, err := utils.NormalizeServerOptions(properties)
	if err != nil {
		return nil, err
	}
	return &Server{opt: opt}, nil
}

// CreateInbound creates and starts an inbound listener on a port, mirroring
// the createInbound. The spec throws on bad per-listener options, Go returns the
// error.
func (s *Server) CreateInbound(properties ListenerOptions, callback InboundHandler) (*Inbound, error) {
	return newInbound(s, properties, callback)
}
