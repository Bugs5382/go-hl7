// Tests for the segment separator a parsed message picks (issue #36).
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

// reencodeStable parses text, encodes it, parses the encoding and encodes it
// again. Both parses must agree: the same segment count and the same encoding.
func reencodeStable(t *testing.T, text string) *Message {
	t.Helper()
	first := mustMessage(t, text)
	enc := first.String()
	second, err := NewMessage(MessageOptions{Text: enc})
	if err != nil {
		t.Fatalf("re-parse of %q failed: %v", enc, err)
	}
	if first.Len() != second.Len() {
		t.Fatalf("segment count changed on re-parse: %d then %d\ninput:   %q\nencoded: %q", first.Len(), second.Len(), text, enc)
	}
	if enc2 := second.String(); enc2 != enc {
		t.Fatalf("re-encoding is not stable\nfirst:  %q\nsecond: %q", enc, enc2)
	}
	return first
}

func TestMessageOnlyTrailingCRSplitsOnLF(t *testing.T) {
	// LF separates the segments. The only CR is the trailing one, which is
	// trimmed, so it must not decide the separator.
	text := "MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5\nEVN|A01\nPID|1||MRN1\r"
	m := reencodeStable(t, text)
	if m.Len() != 3 {
		t.Fatalf("expected 3 segments split on LF, got %d", m.Len())
	}
	if got := m.Get("EVN.1").String(); got != "A01" {
		t.Fatalf("EVN.1: got %q want %q", got, "A01")
	}
	if got := m.Get("PID.3").String(); got != "MRN1" {
		t.Fatalf("PID.3: got %q want %q", got, "MRN1")
	}
}

func TestMessageSeparatorAfterTrimming(t *testing.T) {
	cases := map[string]string{
		"single segment, trailing CR":     "MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5\r",
		"LF segments, trailing CR LF":     "MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5\nEVN|A01\r\n",
		"CR segments, trailing CR":        "MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5\rEVN|A01\r",
		"CR segments with a stray LF":     "MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5\rEVN|A01\nX\r",
		"LF segments, no trailing marker": "MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5\nEVN|A01",
	}
	for name, text := range cases {
		t.Run(name, func(t *testing.T) {
			reencodeStable(t, text)
		})
	}
}

// segmentCount returns the number of segments in m. It reports false when
// reading the segments panics, which an empty segment line does today (issue
// #38); that is a separate bug from the separator choice tested here.
func segmentCount(m *Message) (n int, ok bool) {
	defer func() {
		if recover() != nil {
			n, ok = 0, false
		}
	}()
	return m.Len(), true
}

// FuzzMessageReencode checks that for every input that parses, parsing its
// encoding gives the same message: parse(encode(parse(x))) equals parse(x).
func FuzzMessageReencode(f *testing.F) {
	seeds := []string{
		"MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5\rEVN|A01\rPID|1||MRN1",
		"MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5\nEVN|A01\nPID|1||MRN1\r",
		"MSH|^~\\&|A|FAC|||20260101000000||ADT^A01|1|P|2.5\nEVN|A01\r\n",
	}
	for _, s := range seeds {
		f.Add(s)
		f.Add(strings.ReplaceAll(s, "\r", "\n"))
	}

	f.Fuzz(func(t *testing.T, text string) {
		if len(strings.TrimSpace(text)) <= 3 {
			t.Skip("text shorter than a header panics, see issue #39")
		}
		first, err := NewMessage(MessageOptions{Text: text})
		if err != nil {
			return
		}
		n1, ok := segmentCount(first)
		if !ok {
			t.Skip("empty segment line, see issue #38")
		}
		enc := first.String()
		second, err := NewMessage(MessageOptions{Text: enc})
		if err != nil {
			t.Fatalf("re-encoded message no longer parses: %v\ninput:   %q\nencoded: %q", err, text, enc)
		}
		n2, ok := segmentCount(second)
		if !ok || n1 != n2 {
			t.Fatalf("segment count changed on re-parse: %d then %d (ok=%v)\ninput:   %q\nencoded: %q", n1, n2, ok, text, enc)
		}
		if enc2 := second.String(); enc2 != enc {
			t.Fatalf("re-encoding is not stable\ninput:  %q\nfirst:  %q\nsecond: %q", text, enc, enc2)
		}
	})
}
