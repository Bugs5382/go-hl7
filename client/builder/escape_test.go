// Tests for write-side delimiter escaping and its round-trip with unescape.
package builder

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
	"testing"
)

// baseMessage is a minimal, well-formed message with the standard encoding
// characters used as a starting point for write tests.
const baseMessage = "MSH|^~\\&|APP|FAC|RAPP|RFAC|20240101||ADT^A01|MSGID|P|2.7"

// TestEscapeRoundTripLeaf verifies that a leaf value carrying every delimiter is
// encoded on write and decoded to the original bytes on read, and that the wire
// form carries the expected escape sequences.
func TestEscapeRoundTripLeaf(t *testing.T) {
	m := mustMessage(t, baseMessage)

	// A sub-component is a leaf: it can hold no sub-structure, so every embedded
	// delimiter (field, component, repetition, sub-component and the escape
	// character itself) is data and must be escaped.
	const value = "a|b^c~d&e\\f"
	m.Set("ZZZ.1.1.1", value)

	wire := m.String()
	for _, want := range []string{`\F\`, `\S\`, `\T\`, `\R\`, `\E\`} {
		if !strings.Contains(wire, want) {
			t.Fatalf("serialized wire form is missing %q\nwire: %q", want, wire)
		}
	}
	// The raw delimiter bytes must not survive unescaped inside the value.
	if strings.Contains(wire, "ZZZ|") && strings.Contains(strings.SplitN(wire, "ZZZ", 2)[1], "^c") {
		t.Fatalf("value delimiters leaked unescaped into the wire form: %q", wire)
	}

	got := mustMessage(t, wire).Get("ZZZ.1.1.1").String()
	if got != value {
		t.Fatalf("round-trip mismatch:\n got: %q\nwant: %q", got, value)
	}
}

// TestEscapeRoundTripCustomDelimiters proves the escape characters are read from
// the message (not hard-coded), so a message declaring custom encoding
// characters still round-trips.
func TestEscapeRoundTripCustomDelimiters(t *testing.T) {
	// field "!", component "+", repetition "?", escape "#", sub-component "@".
	const custom = "MSH!+?#@!APP!FAC!RAPP!RFAC!20240101!!ADT+A01!MSGID!P!2.7"
	m := mustMessage(t, custom)

	const value = "a!b+c?d@e#f"
	m.Set("ZZZ.1.1.1", value)

	wire := m.String()
	for _, want := range []string{"#F#", "#S#", "#T#", "#R#", "#E#"} {
		if !strings.Contains(wire, want) {
			t.Fatalf("custom-delimiter wire form is missing %q\nwire: %q", want, wire)
		}
	}

	got := mustMessage(t, wire).Get("ZZZ.1.1.1").String()
	if got != value {
		t.Fatalf("custom-delimiter round-trip mismatch:\n got: %q\nwant: %q", got, value)
	}
}

// TestEscapePreservesCompositeLiterals verifies the level-aware behavior: a
// component separator written at the field level is structural and must stay
// raw, while a field separator at the same level is data and must be escaped.
func TestEscapePreservesCompositeLiterals(t *testing.T) {
	m := mustMessage(t, baseMessage)
	m.Set("PID.5", "DOE^JANE")

	wire := m.String()
	if !strings.Contains(wire, "DOE^JANE") {
		t.Fatalf("component separator was escaped in a field-level composite\nwire: %q", wire)
	}

	reparsed := mustMessage(t, wire)
	if last := reparsed.Get("PID.5.1").String(); last != "DOE" {
		t.Fatalf("PID.5.1 = %q, want %q", last, "DOE")
	}
	if first := reparsed.Get("PID.5.2").String(); first != "JANE" {
		t.Fatalf("PID.5.2 = %q, want %q", first, "JANE")
	}

	// A field separator embedded in a field value is data and must be escaped
	// even though it is written at the field level.
	m.Set("PID.3", "A|B")
	if got := mustMessage(t, m.String()).Get("PID.3.1").String(); got != "A|B" {
		t.Fatalf("PID.3.1 = %q, want %q", got, "A|B")
	}
}

// TestEscapeEdgeCases covers empty and delimiter-free values plus the direct
// escape/unescape inverse relationship.
func TestEscapeEdgeCases(t *testing.T) {
	m := mustMessage(t, baseMessage)

	t.Run("empty value round-trips", func(t *testing.T) {
		m.Set("ZZZ.1.1.1", "")
		if got := mustMessage(t, m.String()).Get("ZZZ.1.1.1").String(); got != "" {
			t.Fatalf("empty value = %q, want empty", got)
		}
	})

	t.Run("delimiter-free value is untouched", func(t *testing.T) {
		const plain = "PlainValue123"
		m.Set("ZZZ.1.1.1", plain)
		wire := m.String()
		if strings.Contains(wire, `\`) && strings.Contains(wire, "PlainValue") {
			// no escape sequence should be introduced for a clean value
			if strings.Contains(wire, `\E\`) || strings.Contains(wire, `\F\`) {
				t.Fatalf("clean value was escaped: %q", wire)
			}
		}
		if got := mustMessage(t, wire).Get("ZZZ.1.1.1").String(); got != plain {
			t.Fatalf("plain value = %q, want %q", got, plain)
		}
	})

	t.Run("escape and unescape are inverses", func(t *testing.T) {
		root := &m.RootBase
		for _, in := range []string{"", "no delims", "a|b", "a^b~c&d", "\\", "|^~&\\"} {
			if out := root.unescape(root.escape(in)); out != in {
				t.Fatalf("unescape(escape(%q)) = %q, want %q", in, out, in)
			}
		}
	})
}

// TestEscapeDoesNotTouchEncodingCharacters guards the MSH separator fields: the
// field separator and encoding characters must never be escaped.
func TestEscapeDoesNotTouchEncodingCharacters(t *testing.T) {
	m := mustMessage(t, baseMessage)
	m.Set("PID.3", "MRN1")
	wire := m.String()
	if !strings.HasPrefix(wire, "MSH|^~\\&|") {
		t.Fatalf("MSH encoding characters were altered\nwire: %q", wire)
	}
}
