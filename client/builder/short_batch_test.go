// Tests for batch and file-batch text too short to hold a header (issue #63).
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

// A bare BHS or FHS header has no field separator, so the constructors return
// an HL7FatalError instead of panicking.
func TestBatchTextShorterThanHeader(t *testing.T) {
	for _, text := range []string{"BHS", "BHS\r\n", "BHS  "} {
		t.Run("batch "+text, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("NewBatch(%q) panicked: %v", text, r)
				}
			}()
			b, err := NewBatch(BatchOptions{Text: text})
			if err == nil || !errors.Is(err, helpers.ErrFatal) {
				t.Fatalf("NewBatch(%q): expected an HL7FatalError, got %v", text, err)
			}
			if b != nil {
				t.Fatalf("NewBatch(%q): expected a nil batch", text)
			}
		})
	}
	for _, text := range []string{"FHS", "FHS\r\n", "FHS  "} {
		t.Run("file "+text, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("NewFileBatch(%q) panicked: %v", text, r)
				}
			}()
			f, err := NewFileBatch(FileOptions{Text: text})
			if err == nil || !errors.Is(err, helpers.ErrFatal) {
				t.Fatalf("NewFileBatch(%q): expected an HL7FatalError, got %v", text, err)
			}
			if f != nil {
				t.Fatalf("NewFileBatch(%q): expected a nil file batch", text)
			}
		})
	}
}

// ParseMessages reports a batch or file with no messages as an error where
// Messages panics.
func TestParseMessagesNoPanic(t *testing.T) {
	b, err := NewBatch(BatchOptions{Text: "BHS|^~\\&|A|B|C|D|20240101"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.ParseMessages(); err == nil || !errors.Is(err, helpers.ErrFatal) {
		t.Fatalf("Batch.ParseMessages: expected an HL7FatalError, got %v", err)
	}
	f, err := NewFileBatch(FileOptions{Text: "FHS|^~\\&|A|B"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.ParseMessages(); err == nil || !errors.Is(err, helpers.ErrFatal) {
		t.Fatalf("FileBatch.ParseMessages: expected an HL7FatalError, got %v", err)
	}
}

// ParseMessages returns the carried messages like Messages does.
func TestParseMessagesReturnsMessages(t *testing.T) {
	text := "BHS|^~\\&|A|B\rMSH|^~\\&|A|B|C|D|2024||ADT^A01|1|P|2.5\rBTS|1"
	b, err := NewBatch(BatchOptions{Text: text})
	if err != nil {
		t.Fatal(err)
	}
	msgs, err := b.ParseMessages()
	if err != nil || len(msgs) != 1 {
		t.Fatalf("expected 1 message and no error, got %d, %v", len(msgs), err)
	}
}
