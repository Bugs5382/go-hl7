package builder_test

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

	"github.com/Bugs5382/go-hl7/client/builder"
)

// These tests cover low-level Segment, SegmentList, SubComponent, Component,
// FieldRepetition behavior and the RootBase escape/unescape paths. Set splits
// into Set/SetIndex; array values pass as a []any to Set.

func freshSegment(t *testing.T) (*builder.Message, *builder.Segment) {
	t.Helper()
	m, _ := builder.NewMessage(builder.MessageOptions{MessageHeader: &builder.MessageHeader{MSH9_1: "ADT", MSH9_2: "A01"}})
	seg, _ := m.AddSegment("PID")
	return m, seg
}

func ctn(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected %q to contain %q", s, sub)
	}
}

func TestSegmentSet(t *testing.T) {
	t.Run("string path + array value writes sub-components", func(t *testing.T) {
		m, seg := freshSegment(t)
		seg.Set("3", []any{"MRN1", "MRN2", "MRN3"})
		if got := m.Get("PID.3.1").String(); got != "MRN1" {
			t.Fatalf("PID.3.1 = %q", got)
		}
		if got := m.Get("PID.3.2").String(); got != "MRN2" {
			t.Fatalf("PID.3.2 = %q", got)
		}
		if got := m.Get("PID.3.3").String(); got != "MRN3" {
			t.Fatalf("PID.3.3 = %q", got)
		}
	})

	t.Run("numeric path + array value writes a repeated field", func(t *testing.T) {
		_, seg := freshSegment(t)
		seg.SetIndex(5, []any{"DOE", "JANE"})
		ctn(t, seg.String(), "DOE")
		ctn(t, seg.String(), "JANE")
	})

	t.Run("numeric path + scalar writes a single field", func(t *testing.T) {
		m, seg := freshSegment(t)
		seg.SetIndex(3, "MRN1")
		ctn(t, m.Get("PID.3").String(), "MRN1")
	})
}

func TestSegmentList(t *testing.T) {
	withTwo := func(t *testing.T) *builder.Message {
		m, _ := builder.NewMessage(builder.MessageOptions{MessageHeader: &builder.MessageHeader{MSH9_1: "ADT", MSH9_2: "A01"}})
		s1, _ := m.AddSegment("EVN")
		s1.Set("1", "A01")
		s2, _ := m.AddSegment("EVN")
		s2.Set("1", "A04")
		return m
	}

	t.Run("get EVN toString returns the first segment", func(t *testing.T) {
		ctn(t, withTwo(t).Get("EVN").String(), "EVN|A01")
	})
	t.Run("get EVN.1 reads via the list to the first segment", func(t *testing.T) {
		if got := withTwo(t).Get("EVN.1").String(); got != "A01" {
			t.Fatalf("EVN.1 = %q", got)
		}
	})
	t.Run("set EVN.1 writes through the list to the first segment", func(t *testing.T) {
		m := withTwo(t)
		m.Set("EVN.1", "A08")
		if got := m.Get("EVN.1").String(); got != "A08" {
			t.Fatalf("EVN.1 = %q", got)
		}
	})
}

func TestCompositeParsing(t *testing.T) {
	text := "MSH|^~\\&|||||20240101000000||ADT^A01|CTRL|D|2.7\r" +
		"PID|||MRN1^^^FAC&SYS&IDTYPE~MRN2^^^FAC2&SYS2"

	t.Run("Component read returns the named sub-component", func(t *testing.T) {
		m, _ := builder.NewMessage(builder.MessageOptions{Text: text})
		if got := m.Get("PID.3.4.2").String(); got != "SYS" {
			t.Fatalf("PID.3.4.2 = %q", got)
		}
	})
	t.Run("FieldRepetition read traverses repetition then sub-paths", func(t *testing.T) {
		m, _ := builder.NewMessage(builder.MessageOptions{Text: text})
		if got := m.Get("PID.3").Index(1).Index(0).String(); got != "MRN2" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("SubComponent isEmpty is true for missing depths", func(t *testing.T) {
		m, _ := builder.NewMessage(builder.MessageOptions{Text: text})
		if !m.Get("PID.3.4.99").IsEmpty() {
			t.Fatal("expected empty")
		}
	})
	t.Run("SubComponent isEmpty is false when value exists", func(t *testing.T) {
		m, _ := builder.NewMessage(builder.MessageOptions{Text: text})
		if m.Get("PID.3.4.1").IsEmpty() {
			t.Fatal("expected non-empty")
		}
	})
}

func TestRootBaseEscape(t *testing.T) {
	t.Run("unescape passes through text with no escape character", func(t *testing.T) {
		m, _ := builder.NewMessage(builder.MessageOptions{MessageHeader: &builder.MessageHeader{MSH9_1: "ADT", MSH9_2: "A01"}})
		m.Set("PID.5.1", "DOE")
		if got := m.Get("PID.5.1").String(); got != "DOE" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("unescape decodes F/S/R/T/E sequences", func(t *testing.T) {
		raw := "MSH|^~\\&|||||20240101000000||ADT^A01|CTRL|D|2.7\rNTE|1|L|A\\F\\B\\S\\C\\R\\D\\T\\E\\E\\F"
		m, _ := builder.NewMessage(builder.MessageOptions{Text: raw})
		out := m.Get("NTE.3").String()
		ctn(t, out, "A|B")
		ctn(t, out, "^C")
		ctn(t, out, "~D")
		ctn(t, out, "&E")
	})
	t.Run("unescape decodes hex sequences to characters", func(t *testing.T) {
		raw := "MSH|^~\\&|||||20240101000000||ADT^A01|CTRL|D|2.7\rNTE|1|L|\\X4869\\"
		m, _ := builder.NewMessage(builder.MessageOptions{Text: raw})
		if got := m.Get("NTE.3").String(); got != "Hi" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("unescape drops unsupported control codes", func(t *testing.T) {
		raw := "MSH|^~\\&|||||20240101000000||ADT^A01|CTRL|D|2.7\rNTE|1|L|A\\C\\B\\H\\C\\M\\D\\N\\E\\Z\\"
		m, _ := builder.NewMessage(builder.MessageOptions{Text: raw})
		if got := m.Get("NTE.3").String(); got != "ABCDE" {
			t.Fatalf("got %q", got)
		}
	})
}
