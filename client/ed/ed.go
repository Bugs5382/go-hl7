// Package ed encodes and decodes the HL7 v2 ED (Encapsulated Data) datatype,
// the value carried by OBX-5 when OBX-2 is "ED" (for example a Base64 PDF).
//
// An ED value is a five-component structure joined by the HL7 component
// delimiter:
//
//	source ^ type-of-data ^ data-subtype ^ encoding ^ data
//
// for example "^application^PDF^Base64^<base64>". The delimiters are real
// component separators, so an OBX-5 ED value legitimately holds five
// components. The data component is encoded per the encoding component
// (HL7 Table 0299); Base64 and hexadecimal use alphabets that exclude every
// HL7 delimiter, so the payload is delimiter-safe and survives serialization
// without escaping.
//
// This package is a pure codec over the wire string. Wire it to the message
// model with the builder package:
//
//	msg.Set("OBX.2", "ED")
//	msg.Set("OBX.5", ed.Encode(pdfBytes, "application", "PDF"))
//	// ... serialize, transmit, parse ...
//	value, err := ed.Decode(msg.Get("OBX.5").Raw())
//
// Get("OBX.5").Raw() returns the joined five-component wire value; .String()
// resolves to the first component only, so Raw() is the value to decode.
package ed

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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

// Encoding values from HL7 Table 0299 (encoding).
const (
	// EncodingNone ("A") means the data component is carried verbatim, not
	// encoded.
	EncodingNone = "A"
	// EncodingHex ("Hex") means the data component is hexadecimal.
	EncodingHex = "Hex"
	// EncodingBase64 ("Base64") means the data component is RFC 4648 base64.
	EncodingBase64 = "Base64"
)

// componentSeparator is the standard HL7 component delimiter used to join the
// five ED components on the wire.
const componentSeparator = "^"

// ED is a decoded HL7 Encapsulated Data value. It holds the four metadata
// components and the raw, decoded payload bytes; Marshal re-encodes Data
// according to Encoding.
type ED struct {
	// Source is ED.1, the source application (HL7 HD). Usually empty.
	Source string
	// Type is ED.2, the type of data (HL7 Table 0191), e.g. "application".
	Type string
	// Subtype is ED.3, the data subtype (HL7 Table 0291), e.g. "PDF".
	Subtype string
	// Encoding is ED.4, the encoding (HL7 Table 0299); one of the Encoding*
	// constants.
	Encoding string
	// Data is ED.5, the raw decoded payload bytes.
	Data []byte
}

// Encode builds a Base64 ED value string carrying data with the given type and
// subtype and an empty source. The result is "^type^subtype^Base64^<base64>".
// Base64 encoding never fails, so no error is returned.
func Encode(data []byte, typ, subtype string) string {
	value, _ := ED{
		Type:     typ,
		Subtype:  subtype,
		Encoding: EncodingBase64,
		Data:     data,
	}.Marshal()
	return value
}

// Marshal serializes the ED to its five-component wire value, encoding Data
// according to Encoding. It returns an error for an unsupported encoding.
func (e ED) Marshal() (string, error) {
	var payload string
	switch e.Encoding {
	case EncodingBase64:
		payload = base64.StdEncoding.EncodeToString(e.Data)
	case EncodingHex:
		payload = hex.EncodeToString(e.Data)
	case EncodingNone, "":
		payload = string(e.Data)
	default:
		return "", fmt.Errorf("hl7 ed: unsupported encoding %q", e.Encoding)
	}
	return strings.Join(
		[]string{e.Source, e.Type, e.Subtype, e.Encoding, payload},
		componentSeparator,
	), nil
}

// Decode parses a five-component ED value string and decodes the data component
// per its declared encoding. It returns an error for an unsupported encoding or
// a malformed payload. Missing trailing components decode to their zero values.
func Decode(value string) (ED, error) {
	parts := strings.Split(value, componentSeparator)
	component := func(i int) string {
		if i < len(parts) {
			return parts[i]
		}
		return ""
	}

	e := ED{
		Source:   component(0),
		Type:     component(1),
		Subtype:  component(2),
		Encoding: component(3),
	}
	payload := component(4)

	switch e.Encoding {
	case EncodingBase64:
		raw, err := base64.StdEncoding.DecodeString(payload)
		if err != nil {
			return ED{}, fmt.Errorf("hl7 ed: decode base64 data: %w", err)
		}
		e.Data = raw
	case EncodingHex:
		raw, err := hex.DecodeString(payload)
		if err != nil {
			return ED{}, fmt.Errorf("hl7 ed: decode hex data: %w", err)
		}
		e.Data = raw
	case EncodingNone, "":
		e.Data = []byte(payload)
	default:
		return ED{}, fmt.Errorf("hl7 ed: unsupported encoding %q", e.Encoding)
	}
	return e, nil
}
