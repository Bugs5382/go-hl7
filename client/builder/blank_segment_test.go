// Tests for blank segment lines in a parsed message (issue #38).
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

// noPanic runs fn and fails the test if it panics.
func noPanic(t *testing.T, what string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("%s panicked: %v", what, r)
		}
	}()
	fn()
}

// A blank segment line carries no segment, so the parser skips it. The rest
// of the message reads as if the blank line were not there.
func TestMessageSkipsBlankSegmentLines(t *testing.T) {
	const msh = "MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5"
	cases := map[string]string{
		"two CRs in a row":          msh + "\r\rEVN|A01\rPID|1||MRN1",
		"blank LF line":             msh + "\n\nEVN|A01\nPID|1||MRN1",
		"several blank lines":       msh + "\r\r\r\rEVN|A01\r\r\rPID|1||MRN1",
		"whitespace-only line":      msh + "\r  \t \rEVN|A01\rPID|1||MRN1",
		"LF-only line in CR text":   msh + "\r\n\rEVN|A01\rPID|1||MRN1",
		"blank line before trailer": msh + "\rEVN|A01\rPID|1||MRN1\r\r\r",
	}
	for name, text := range cases {
		t.Run(name, func(t *testing.T) {
			m, err := NewMessage(MessageOptions{Text: text})
			if err != nil {
				t.Fatalf("NewMessage: %v", err)
			}
			noPanic(t, "reading the message", func() {
				if got := m.Len(); got != 3 {
					t.Fatalf("expected 3 segments, got %d", got)
				}
				if got := m.Get("EVN.1").String(); got != "A01" {
					t.Fatalf("EVN.1: got %q want %q", got, "A01")
				}
				if got := m.Get("PID.3").String(); got != "MRN1" {
					t.Fatalf("PID.3: got %q want %q", got, "MRN1")
				}
				if got := m.TotalSegment("EVN"); got != 1 {
					t.Fatalf("TotalSegment(EVN): got %d want 1", got)
				}
				m.ForEach(func(v HL7Node, i int) {
					if strings.TrimSpace(v.String()) == "" {
						t.Fatalf("segment %d is blank", i)
					}
				})
			})
			enc := m.String()
			nl := m.Delimiters()[:1]
			if strings.Contains(enc, nl+nl) {
				t.Fatalf("encoding still holds a blank segment line: %q", enc)
			}
			reencodeStable(t, text)
		})
	}
}

// A message with only blank lines after MSH is just the MSH segment.
func TestMessageOnlyBlankLinesAfterMSH(t *testing.T) {
	m, err := NewMessage(MessageOptions{Text: "MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5\r\r \r"})
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	noPanic(t, "reading the message", func() {
		if got := m.Len(); got != 1 {
			t.Fatalf("expected 1 segment, got %d", got)
		}
		if got := m.Get("MSH.9.1").String(); got != "ADT" {
			t.Fatalf("MSH.9.1: got %q want %q", got, "ADT")
		}
	})
}

// Writing to a message that had blank lines appends after the real segments,
// not after a blank one.
func TestMessageSetAfterBlankSegmentLines(t *testing.T) {
	m, err := NewMessage(MessageOptions{Text: "MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5\r\rEVN|A01"})
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	noPanic(t, "writing the message", func() {
		m.Set("PID.3", "MRN9")
	})
	want := "MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5\rEVN|A01\rPID|||MRN9"
	if got := m.String(); got != want {
		t.Fatalf("encoding:\n got %q\nwant %q", got, want)
	}
}

// Batches and file batches share the segment parsing, so their blank lines are
// skipped too. A CR LF file read from disk or a buffer turns every line ending
// into two CRs, which left a blank segment after each line.
func TestBatchAndFileSkipBlankSegmentLines(t *testing.T) {
	const msh = "MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5"
	t.Run("batch", func(t *testing.T) {
		b, err := NewBatch(BatchOptions{Text: "BHS|^~\\&|A\r\r" + msh + "\rEVN|A01\r\r" + msh + "\rEVN|A02\r \rBTS|2"})
		if err != nil {
			t.Fatalf("NewBatch: %v", err)
		}
		noPanic(t, "reading the batch", func() {
			if got := b.Len(); got != 6 {
				t.Fatalf("expected 6 segments, got %d", got)
			}
			if got := len(b.Messages()); got != 2 {
				t.Fatalf("expected 2 messages, got %d", got)
			}
			if got := b.Get("BTS.1").String(); got != "2" {
				t.Fatalf("BTS.1: got %q want %q", got, "2")
			}
		})
	})
	t.Run("file buffer with CR LF line endings", func(t *testing.T) {
		text := "FHS|^~\\&|A\r\nBHS|^~\\&|A\r\n" + msh + "\r\nEVN|A01\r\nBTS|1\r\nFTS|1\r\n"
		f, err := NewFileBatch(FileOptions{FileBuffer: []byte(text)})
		if err != nil {
			t.Fatalf("NewFileBatch: %v", err)
		}
		noPanic(t, "reading the file batch", func() {
			if got := f.Len(); got != 6 {
				t.Fatalf("expected 6 segments, got %d", got)
			}
			if got := len(f.Messages()); got != 1 {
				t.Fatalf("expected 1 message, got %d", got)
			}
			if got := f.Get("FTS.1").String(); got != "1" {
				t.Fatalf("FTS.1: got %q want %q", got, "1")
			}
		})
	})
}
