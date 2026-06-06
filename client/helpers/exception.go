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

// Package helpers carries the client package's error hierarchy and shared
// constants. It mirrors the helpers/ folder (exception.ts, constants.ts).
package helpers

import (
	"errors"
	"fmt"
)

// Sentinel errors back the HL7Error hierarchy so callers can match with
// errors.Is where the reference used `instanceof`. Go necessity: the spec throws typed
// Error subclasses; Go has no exceptions, so each error type unwraps to one of
// these sentinels.
var (
	// ErrFatal indicates a fatal failure of a connection (HL7FatalError, 500).
	ErrFatal = errors.New("hl7: fatal")
	// ErrParser indicates a parser failure (HL7ParserError, 404).
	ErrParser = errors.New("hl7: parser")
	// ErrValidation indicates a validator failure (HL7ValidationError, 404).
	ErrValidation = errors.New("hl7: validation")
)

// HL7Error is the parent of the HL7 error hierarchy. It carries a numeric code
// and a stable name, mirroring the HL7Error class.
type HL7Error struct {
	// Code is the numeric error code (e.g. 500, 404).
	Code int
	// Name is the stable error name (e.g. "HL7FatalError").
	Name string
	// Msg is the human-readable message.
	Msg string
	// sentinel is the wrapped sentinel for errors.Is matching.
	sentinel error
}

// Error implements the error interface.
func (e *HL7Error) Error() string { return e.Msg }

// Unwrap returns the wrapped sentinel so errors.Is(err, ErrParser) and friends
// resolve through the typed error.
func (e *HL7Error) Unwrap() error { return e.sentinel }

// HL7FatalError indicates a fatal failure of a connection. It mirrors
// the HL7FatalError (code 500).
type HL7FatalError struct{ HL7Error }

// NewHL7FatalError constructs an HL7FatalError with the given message.
func NewHL7FatalError(message string) *HL7FatalError {
	return &HL7FatalError{HL7Error{Code: 500, Name: "HL7FatalError", Msg: message, sentinel: ErrFatal}}
}

// HL7ParserError indicates a failure while parsing an HL7 message. It mirrors
// the HL7ParserError (code 404).
type HL7ParserError struct{ HL7Error }

// NewHL7ParserError constructs an HL7ParserError with the given message.
func NewHL7ParserError(message string) *HL7ParserError {
	return &HL7ParserError{HL7Error{Code: 404, Name: "HL7ParserError", Msg: message, sentinel: ErrParser}}
}

// HL7ValidationError indicates a failure of the spec-driven field validator.
// It mirrors the HL7ValidationError (code 404).
type HL7ValidationError struct{ HL7Error }

// NewHL7ValidationError constructs an HL7ValidationError with the given message.
func NewHL7ValidationError(message string) *HL7ValidationError {
	return &HL7ValidationError{HL7Error{Code: 404, Name: "HL7ValidationError", Msg: message, sentinel: ErrValidation}}
}

// NewHL7FatalErrorf is a formatting convenience for NewHL7FatalError.
func NewHL7FatalErrorf(format string, args ...any) *HL7FatalError {
	return NewHL7FatalError(fmt.Sprintf(format, args...))
}

// NewHL7ParserErrorf is a formatting convenience for NewHL7ParserError.
func NewHL7ParserErrorf(format string, args ...any) *HL7ParserError {
	return NewHL7ParserError(fmt.Sprintf(format, args...))
}

// NewHL7ValidationErrorf is a formatting convenience for NewHL7ValidationError.
func NewHL7ValidationErrorf(format string, args ...any) *HL7ValidationError {
	return NewHL7ValidationError(fmt.Sprintf(format, args...))
}
