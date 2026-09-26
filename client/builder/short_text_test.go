// Tests for message text too short to hold an MSH header (issue #39).
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
	"errors"
	"testing"

	"github.com/Bugs5382/go-hl7/client/helpers"
)

// newMessageNoPanic calls NewMessage and fails the test if it panics.
func newMessageNoPanic(t *testing.T, text string) (m *Message, err error) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("NewMessage(%q) panicked: %v", text, r)
		}
	}()
	return NewMessage(MessageOptions{Text: text})
}

// Text of 1 to 3 bytes cannot hold "MSH" and a field separator, so NewMessage
// returns an HL7FatalError instead of panicking.
func TestMessageTextShorterThanHeader(t *testing.T) {
	// "MSH\r\n" and "MSH  " are longer than 3 bytes, but their only
	// "separator" is whitespace that trimming removes.
	cases := []string{"0", " ", "M", "MS", "MSH", "XYZ", "\r\n\r", "MSH\r\n", "MSH  "}
	for _, text := range cases {
		t.Run(text, func(t *testing.T) {
			m, err := newMessageNoPanic(t, text)
			if err == nil {
				t.Fatalf("NewMessage(%q): expected an error, got a message with %d segments", text, m.Len())
			}
			if !errors.Is(err, helpers.ErrFatal) {
				t.Fatalf("NewMessage(%q): expected an HL7FatalError, got %v", text, err)
			}
			if m != nil {
				t.Fatalf("NewMessage(%q): expected a nil message with the error", text)
			}
		})
	}
}

// Text that does not start with MSH is rejected at every length, not only at
// three bytes and up.
func TestMessageTextNotStartingWithMSH(t *testing.T) {
	for _, text := range []string{"X", "XY", "XYZ|", "EVN|A01"} {
		t.Run(text, func(t *testing.T) {
			_, err := newMessageNoPanic(t, text)
			if !errors.Is(err, helpers.ErrFatal) {
				t.Fatalf("NewMessage(%q): expected an HL7FatalError, got %v", text, err)
			}
		})
	}
}

// Lengths 0 through 4: empty text still builds an empty message, 1 to 3 bytes
// fail, and "MSH|" (the shortest header) parses.
func TestMessageTextLengthsZeroToFour(t *testing.T) {
	cases := []struct {
		text    string
		wantErr bool
	}{
		{"", false},
		{"M", true},
		{"MS", true},
		{"MSH", true},
		{"MSH|", false},
	}
	for _, tc := range cases {
		t.Run("len"+string(rune('0'+len(tc.text))), func(t *testing.T) {
			m, err := newMessageNoPanic(t, tc.text)
			if tc.wantErr {
				if !errors.Is(err, helpers.ErrFatal) {
					t.Fatalf("NewMessage(%q): expected an HL7FatalError, got %v", tc.text, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewMessage(%q): unexpected error %v", tc.text, err)
			}
			noPanic(t, "reading the message", func() {
				_ = m.Len()
				_ = m.Get("MSH.9.1").String()
				_ = m.String()
			})
			if tc.text == "MSH|" && m.Get("MSH.1").String() != "|" {
				t.Fatalf("MSH.1: got %q want %q", m.Get("MSH.1").String(), "|")
			}
		})
	}
}
