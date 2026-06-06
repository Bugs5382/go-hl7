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
	"testing"

	"github.com/Bugs5382/go-hl7/client/builder"
)

// These tests cover ValueNode value coercions and the EmptyNode contract,
// exercised through the public Message API. The value coercions return (T, ok),
// and EmptyNode panics only on structural access (Raw/Name/Path/ForEach/Read)
// while the coercions return (zero, false).

func sampleMessage(t *testing.T) *builder.Message {
	t.Helper()
	m, err := builder.NewMessage(builder.MessageOptions{
		MessageHeader: &builder.MessageHeader{MSH9_1: "ADT", MSH9_2: "A01"},
	})
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	m.Set("OBX.1", "1")
	m.Set("OBX.2", "NM")
	m.Set("OBX.5", "98.6")
	m.Set("OBX.11", "Y")
	m.Set("OBX.13", "42")
	m.Set("OBX.14", "20240101010159+0100")
	return m
}

func TestValueNodeCoercions(t *testing.T) {
	t.Run("String returns the field value", func(t *testing.T) {
		m := sampleMessage(t)
		if got := m.Get("OBX.5").String(); got != "98.6" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("Int parses int-shaped fields", func(t *testing.T) {
		m := sampleMessage(t)
		if v, ok := m.Get("OBX.13").Int(); !ok || v != 42 {
			t.Fatalf("got %d %v", v, ok)
		}
	})
	t.Run("Float parses decimal-shaped fields", func(t *testing.T) {
		m := sampleMessage(t)
		if v, ok := m.Get("OBX.5").Float(); !ok || v < 98.5 || v > 98.7 {
			t.Fatalf("got %v %v", v, ok)
		}
	})
	t.Run("Bool accepts Y/N", func(t *testing.T) {
		m := sampleMessage(t)
		m.Set("OBX.11", "Y")
		if v, ok := m.Get("OBX.11").Bool(); !ok || !v {
			t.Fatalf("Y -> %v %v", v, ok)
		}
		m.Set("OBX.11", "N")
		if v, ok := m.Get("OBX.11").Bool(); !ok || v {
			t.Fatalf("N -> %v %v", v, ok)
		}
	})
	t.Run("Bool ok is false for non Y/N", func(t *testing.T) {
		m := sampleMessage(t)
		m.Set("OBX.11", "MAYBE")
		if _, ok := m.Get("OBX.11").Bool(); ok {
			t.Fatal("expected ok=false for MAYBE")
		}
	})
	t.Run("Date parses YYYYMMDDHHMMSS+TZ to UTC instant", func(t *testing.T) {
		m := sampleMessage(t)
		d, ok := m.Get("OBX.14").Date()
		if !ok {
			t.Fatal("expected ok")
		}
		if got := d.UTC().Format("2006-01-02T15:04:05Z"); got != "2024-01-01T00:01:59Z" {
			t.Fatalf("got %s", got)
		}
	})
	t.Run("Date without timezone returns a local-time Date", func(t *testing.T) {
		m := sampleMessage(t)
		m.Set("OBX.14", "20240101010159")
		d, ok := m.Get("OBX.14").Date()
		if !ok {
			t.Fatal("expected ok")
		}
		if d.Year() != 2024 || d.Month() != 1 || d.Day() != 1 || d.Hour() != 1 || d.Minute() != 1 || d.Second() != 59 {
			t.Fatalf("got %v", d)
		}
	})
	t.Run("Date handles short YYYYMMDD with zero time", func(t *testing.T) {
		m := sampleMessage(t)
		m.Set("OBX.14", "20240315")
		d, ok := m.Get("OBX.14").Date()
		if !ok {
			t.Fatal("expected ok")
		}
		if d.Year() != 2024 || d.Month() != 3 || d.Day() != 15 || d.Hour() != 0 {
			t.Fatalf("got %v", d)
		}
	})
}

func expectPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	fn()
}

func TestEmptyNode(t *testing.T) {
	node := &builder.EmptyNode{}

	t.Run("missing path on a Message returns an empty value", func(t *testing.T) {
		m, _ := builder.NewMessage(builder.MessageOptions{MessageHeader: &builder.MessageHeader{MSH9_1: "ADT", MSH9_2: "A01"}})
		missing := m.Get("ZZZ.99.99")
		if missing.String() != "" {
			t.Fatalf("String %q", missing.String())
		}
		if missing.Len() != 0 {
			t.Fatalf("Len %d", missing.Len())
		}
		if missing.Exists("anything") {
			t.Fatal("Exists true")
		}
		if !missing.IsEmpty() {
			t.Fatal("IsEmpty false")
		}
	})

	t.Run("get returns itself, set returns itself", func(t *testing.T) {
		if node.Get("any") != node {
			t.Fatal("Get")
		}
		if node.Set("any", "value") != node {
			t.Fatal("Set")
		}
	})

	t.Run("write returns itself (no-op)", func(t *testing.T) {
		if node.Write([]string{"x"}, "value") != node {
			t.Fatal("Write")
		}
	})

	t.Run("String empty, Len 0, IsEmpty true", func(t *testing.T) {
		if node.String() != "" || node.Len() != 0 || !node.IsEmpty() {
			t.Fatal("contract")
		}
	})

	t.Run("Exists false, ToArray empty", func(t *testing.T) {
		if node.Exists("anything") || len(node.ToArray()) != 0 {
			t.Fatal("contract")
		}
	})

	t.Run("Name panics (not implemented)", func(t *testing.T) { expectPanic(t, func() { node.Name() }) })
	t.Run("ForEach panics", func(t *testing.T) { expectPanic(t, func() { node.ForEach(func(builder.HL7Node, int) {}) }) })
	t.Run("Raw panics", func(t *testing.T) { expectPanic(t, func() { node.Raw() }) })
	t.Run("Read panics, Path panics", func(t *testing.T) {
		expectPanic(t, func() { node.Read([]string{"x"}) })
		expectPanic(t, func() { node.Path() })
	})

	t.Run("coercions return ok=false (Go adaptation)", func(t *testing.T) {
		if _, ok := node.Date(); ok {
			t.Fatal("Date ok")
		}
		if _, ok := node.Int(); ok {
			t.Fatal("Int ok")
		}
		if _, ok := node.Float(); ok {
			t.Fatal("Float ok")
		}
		if _, ok := node.Bool(); ok {
			t.Fatal("Bool ok")
		}
	})
}

func TestSegmentListAndSubComponent(t *testing.T) {
	t.Run("repeated segments yield a usable list", func(t *testing.T) {
		m, _ := builder.NewMessage(builder.MessageOptions{MessageHeader: &builder.MessageHeader{MSH9_1: "ADT", MSH9_2: "A01"}})
		seg1, _ := m.AddSegment("EVN")
		seg1.Set("1", "A01")
		seg2, _ := m.AddSegment("EVN")
		seg2.Set("1", "A04")

		list := m.Get("EVN")
		if list.Name() != "EVN" {
			t.Fatalf("Name %q", list.Name())
		}
		count := 0
		list.ForEach(func(builder.HL7Node, int) { count++ })
		if count != 2 {
			t.Fatalf("count %d", count)
		}
	})

	t.Run("sub-component access exercises unescape path", func(t *testing.T) {
		m, _ := builder.NewMessage(builder.MessageOptions{
			Text: "MSH|^~\\&|||||20240101000000||ADT^A01|CTRL_SUB|D|2.7\rPID|||MRN1^^^FAC&SYS&IDTYPE",
		})
		if got := m.Get("PID.3.4.1").String(); got != "FAC" {
			t.Fatalf("PID.3.4.1 = %q", got)
		}
		if got := m.Get("PID.3.4.2").String(); got != "SYS" {
			t.Fatalf("PID.3.4.2 = %q", got)
		}
		if got := m.Get("PID.3.4.3").String(); got != "IDTYPE" {
			t.Fatalf("PID.3.4.3 = %q", got)
		}
	})
}
