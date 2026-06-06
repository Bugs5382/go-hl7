// Package client provides Client, Connection, InboundResponse, and option
// normalization. It connects to a remote HL7 MLLP/TCP (or TLS) server, frames
// messages with the MLLP codec, parses ACKs, and exposes an EventEmitter-style
// event set via On(event, handler).
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

	"github.com/Bugs5382/go-hl7/client/builder"
)

// InboundResponse wraps the parsed response a server sends back after a
// message. The raw string is trimmed and parsed into a Message that the
// outbound handler reads (e.g. MSA.1 for the ACK code).
type InboundResponse struct {
	message *builder.Message
}

// NewInboundResponse parses data (trimmed) into a Message, returning an error
// on a malformed body.
func NewInboundResponse(data string) (*InboundResponse, error) {
	message, err := builder.NewMessage(builder.MessageOptions{Text: strings.TrimRight(data, " \t\r\n")})
	if err != nil {
		return nil, err
	}
	return &InboundResponse{message: message}, nil
}

// GetMessage returns the parsed response Message.
func (r *InboundResponse) GetMessage() *builder.Message { return r.message }
