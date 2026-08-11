package ed_test

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
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/ed"
)

// samplePDF is a small PDF-like payload. It deliberately contains bytes that
// base64 maps to "+" and "/" plus "=" padding, none of which are HL7
// delimiters, proving the encoded data is delimiter-safe.
var samplePDF = []byte("%PDF-1.4\n1 0 obj<< /Type /Catalog >>endobj\xff\xfe\xfd\x00~^|&\\")

const baseMessage = "MSH|^~\\&|APP|FAC|RAPP|RFAC|20240101||ORU^R01|MSGID|P|2.7"

// TestEncodeWireForm asserts the encoded value carries the five ED components
// and the Base64-encoded data.
func TestEncodeWireForm(t *testing.T) {
	value := ed.Encode(samplePDF, "application", "PDF")
	wantB64 := base64.StdEncoding.EncodeToString(samplePDF)
	want := "^application^PDF^Base64^" + wantB64
	if value != want {
		t.Fatalf("Encode value mismatch\n got: %q\nwant: %q", value, want)
	}
	if got := strings.Count(value, "^"); got != 4 {
		t.Fatalf("expected 4 component separators (5 components), got %d in %q", got, value)
	}
}

// TestRoundTripThroughMessage exercises the acceptance path: encode bytes, set
// them on OBX-5, serialize, parse, decode, and compare byte-for-byte.
func TestRoundTripThroughMessage(t *testing.T) {
	msg, err := builder.NewMessage(builder.MessageOptions{Text: baseMessage})
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	if _, err := msg.AddSegment("OBX"); err != nil {
		t.Fatalf("AddSegment: %v", err)
	}
	msg.Set("OBX.2", "ED")
	msg.Set("OBX.5", ed.Encode(samplePDF, "application", "PDF"))

	wire := msg.String()
	if !strings.Contains(wire, "^application^PDF^Base64^") {
		t.Fatalf("wire form missing ED components:\n%q", wire)
	}
	if !strings.Contains(wire, base64.StdEncoding.EncodeToString(samplePDF)) {
		t.Fatalf("wire form missing Base64 data:\n%q", wire)
	}

	reparsed, err := builder.NewMessage(builder.MessageOptions{Text: wire})
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if got := reparsed.Get("OBX.2").String(); got != "ED" {
		t.Fatalf("OBX.2 = %q, want ED", got)
	}

	decoded, err := ed.Decode(reparsed.Get("OBX.5").Raw())
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if !bytes.Equal(decoded.Data, samplePDF) {
		t.Fatalf("round-trip payload mismatch\n got: %v\nwant: %v", decoded.Data, samplePDF)
	}
	if decoded.Type != "application" || decoded.Subtype != "PDF" || decoded.Encoding != ed.EncodingBase64 {
		t.Fatalf("metadata mismatch: %+v", decoded)
	}
	if decoded.Source != "" {
		t.Fatalf("expected empty source, got %q", decoded.Source)
	}
}

// TestDecodeMetadata checks each metadata component is parsed from the wire
// value.
func TestDecodeMetadata(t *testing.T) {
	value := "SENDER^application^PDF^Base64^" + base64.StdEncoding.EncodeToString([]byte("hi"))
	e, err := ed.Decode(value)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if e.Source != "SENDER" || e.Type != "application" || e.Subtype != "PDF" || e.Encoding != "Base64" {
		t.Fatalf("metadata mismatch: %+v", e)
	}
	if string(e.Data) != "hi" {
		t.Fatalf("data = %q, want %q", e.Data, "hi")
	}
}

// TestEncodingsRoundTrip covers the three HL7 Table 0299 encodings.
func TestEncodingsRoundTrip(t *testing.T) {
	payload := []byte("some raw \x00 bytes")
	for _, enc := range []string{ed.EncodingBase64, ed.EncodingHex, ed.EncodingNone} {
		t.Run(enc, func(t *testing.T) {
			marshaled, err := newED(payload, enc).Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			got, err := ed.Decode(marshaled)
			if err != nil {
				t.Fatalf("Decode: %v", err)
			}
			if !bytes.Equal(got.Data, payload) {
				t.Fatalf("%s round-trip mismatch: %v vs %v", enc, got.Data, payload)
			}
		})
	}
}

// newED builds an ED with the given payload and encoding for the table test.
func newED(payload []byte, enc string) ed.ED {
	return ed.ED{Type: "application", Subtype: "octet-stream", Encoding: enc, Data: payload}
}

// TestUnsupportedEncoding proves both Marshal and Decode reject an encoding
// outside HL7 Table 0299.
func TestUnsupportedEncoding(t *testing.T) {
	if _, err := (ed.ED{Encoding: "Gzip", Data: []byte("x")}).Marshal(); err == nil {
		t.Fatal("Marshal: expected error for unsupported encoding")
	}
	if _, err := ed.Decode("^application^PDF^Gzip^data"); err == nil {
		t.Fatal("Decode: expected error for unsupported encoding")
	}
}

// TestDecodeMalformedBase64 checks the payload-decode error path.
func TestDecodeMalformedBase64(t *testing.T) {
	if _, err := ed.Decode("^application^PDF^Base64^not!valid!base64"); err == nil {
		t.Fatal("expected error decoding malformed base64")
	}
}
