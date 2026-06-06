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

// Package utils carries the server package's error hierarchy and option
// normalization, mirroring the utils/ folder (exception.ts,
// normalize.ts, constants.ts).
package utils

import "errors"

// Sentinel errors back the server error hierarchy so callers can match with
// errors.Is where the reference used `instanceof` (the same Go necessity the
// client helpers package documents).
var (
	// ErrServer indicates a server failure (HL7ServerError, code 500).
	ErrServer = errors.New("hl7: server")
	// ErrListener indicates a listener failure (HL7ListenerError).
	ErrListener = errors.New("hl7: listener")
)

// HL7ServerError mirrors the HL7ServerError (code 500). It is the
// MSA.1-validation / option failure error.
type HL7ServerError struct {
	// Code is the numeric error code (500).
	Code int
	// Msg is the human-readable message.
	Msg string
}

// Error implements the error interface.
func (e *HL7ServerError) Error() string { return e.Msg }

// Unwrap returns the server sentinel for errors.Is matching.
func (e *HL7ServerError) Unwrap() error { return ErrServer }

// NewHL7ServerError constructs an HL7ServerError with the given message.
func NewHL7ServerError(message string) *HL7ServerError {
	return &HL7ServerError{Code: 500, Msg: message}
}

// HL7ListenerError mirrors the HL7ListenerError. It is the
// inbound-listener option/parse failure error.
type HL7ListenerError struct {
	// Msg is the human-readable message.
	Msg string
}

// Error implements the error interface.
func (e *HL7ListenerError) Error() string { return e.Msg }

// Unwrap returns the listener sentinel for errors.Is matching.
func (e *HL7ListenerError) Unwrap() error { return ErrListener }

// NewHL7ListenerError constructs an HL7ListenerError with the given message.
func NewHL7ListenerError(message string) *HL7ListenerError {
	return &HL7ListenerError{Msg: message}
}
