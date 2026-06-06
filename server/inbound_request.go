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

package server

import (
	"net"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/server/utils"
)

// InboundRequestProps carries the construction inputs for an InboundRequest,
// mirroring the InboundRequestProps { socket?, type }.
type InboundRequestProps struct {
	// Socket is the connection that produced this message, exposed so handlers
	// can read peer/local details (the socket).
	Socket net.Conn
	// Type is the source type: "message", "batch", or "file" (the type).
	Type string
}

// InboundRequest wraps a parsed inbound message handed to the inbound handler.
// It mirrors the InboundRequest.
type InboundRequest struct {
	fromType string
	message  *builder.Message
	socket   net.Conn
}

// NewInboundRequest constructs an InboundRequest, mirroring the constructor.
func NewInboundRequest(message *builder.Message, properties InboundRequestProps) *InboundRequest {
	return &InboundRequest{
		fromType: properties.Type,
		message:  message,
		socket:   properties.Socket,
	}
}

// GetMessage returns the stored message, mirroring the getMessage. The spec
// throws HL7ListenerError when the message is undefined; Go panics with the
// same error (this mirrors the throw, the EmptyNode-style structural-access
// contract).
func (r *InboundRequest) GetMessage() *builder.Message {
	if r.message != nil {
		return r.message
	}
	panic(utils.NewHL7ListenerError("Message is not defined."))
}

// GetSocket returns the underlying connection, mirroring the getSocket. The spec
// throws when undefined; Go panics with the same error.
func (r *InboundRequest) GetSocket() net.Conn {
	if r.socket != nil {
		return r.socket
	}
	panic(utils.NewHL7ListenerError("Socket is not defined."))
}

// GetType returns the source type, mirroring the getType.
func (r *InboundRequest) GetType() string { return r.fromType }
