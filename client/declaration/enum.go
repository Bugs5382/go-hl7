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

// Package declaration carries the client package's enums and pure type/contract
// declarations. It mirrors the declaration/ folder (enum.ts,
// validationRule.ts, deferred.ts).
package declaration

// Delimiters identifies a kind of HL7 delimiter. Each member is used during
// tree construction to give each level its own index value. It mirrors
// the Delimiters enum, preserving the same member order.
type Delimiters int

const (
	// DelimiterSegment is usually each line of the overall HL7 message.
	DelimiterSegment Delimiters = iota
	// DelimiterField is the field of each segment, usually separated with a |.
	DelimiterField
	// DelimiterComponent is usually within each Field, separated by ^.
	DelimiterComponent
	// DelimiterRepetition is usually within each Component, separated by &.
	DelimiterRepetition
	// DelimiterEscape is the escape string used within the code.
	DelimiterEscape
	// DelimiterSubComponent is usually within each Field, separated by ~.
	DelimiterSubComponent
)

// ReadyState is the state of the connection to the server. It tracks the
// connecting state and the auto-reconnect phase. It mirrors the reference's
// ReadyState enum, preserving the same member order.
type ReadyState int

const (
	// Connecting indicates the client is trying to connect to the server.
	Connecting ReadyState = iota
	// Connected indicates the client is connected to the server.
	Connected
	// Open indicates the client is open but not yet trying to connect.
	Open
	// Closing indicates the client is closing the connection.
	Closing
	// Closed indicates the client connection is closed.
	Closed
)
