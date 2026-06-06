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
	"github.com/Bugs5382/go-hl7/client/utils"
)

// These cover the Batch/FileBatch sanity and edge-case behavior.

func TestBatchRejectsSingleMSH(t *testing.T) {
	_, err := builder.NewBatch(builder.BatchOptions{
		Text: "MSH|^~\\&|||||20081231||ADT^A01^ADT_A01|12345||2.7\rEVN||20081231",
	})
	if err == nil || err.Error() != "Unable to process a single MSH as a batch. Use Message." {
		t.Fatalf("err = %v, want single-MSH rejection", err)
	}
}

func TestIsBatchIsFileOnOutputs(t *testing.T) {
	m, _ := builder.NewMessage(builder.MessageOptions{MessageHeader: mshHeader("CONTROL_ID")})
	if utils.IsBatch(m.String()) {
		t.Fatalf("isBatch should be false for a single Message")
	}
	if utils.IsFile(m.String()) {
		t.Fatalf("isFile should be false for a single Message")
	}

	b, _ := builder.NewBatch(builder.BatchOptions{})
	b.Start("")
	b.End()
	if utils.IsFile(b.String()) {
		t.Fatalf("isFile should be false for a Batch output")
	}
	if !utils.IsBatch(b.String()) {
		t.Fatalf("isBatch should be true for a Batch output")
	}

	f, _ := builder.NewFileBatch(builder.FileOptions{})
	f.Start()
	f.End()
	if !utils.IsFile(f.String()) {
		t.Fatalf("isFile should be true for a FileBatch output")
	}
}

func TestBatchParsesBHSHeader(t *testing.T) {
	hl7Batch := "BHS|^~\\&|||||20231208\rMSH|^~\\&|||||20231208||ADT^A01^ADT_A01|CONTROL_ID||2.7\rEVN||20081231\rEVN||20081231\rBTS|1"
	b, err := builder.NewBatch(builder.BatchOptions{Text: hl7Batch})
	if err != nil {
		t.Fatal(err)
	}
	if got := b.String()[:8]; got != `BHS|^~\&` {
		t.Fatalf("prefix = %q", got)
	}

	messages := b.Messages()
	if len(messages) != 1 {
		t.Fatalf("messages = %d", len(messages))
	}
	count := 0
	messages[0].Get("EVN").ForEach(func(seg builder.HL7Node, _ int) {
		if seg.Name() != "EVN" {
			t.Fatalf("name = %q", seg.Name())
		}
		count++
	})
	if count != 2 {
		t.Fatalf("EVN count = %d", count)
	}
}

func TestBatchNonStandardBHSDelimiters(t *testing.T) {
	hl7 := "BHS:-+?*:::::20231208\rMSH|^~\\&|||||20231208||ADT^A01^ADT_A01|CONTROL_ID||2.7\rEVN||20081231\rEVN||20081231\rBTS|1"
	b, err := builder.NewBatch(builder.BatchOptions{Text: hl7})
	if err != nil {
		t.Fatal(err)
	}
	if got := b.String()[:8]; got != "BHS:-+?*" {
		t.Fatalf("prefix = %q", got)
	}
}

func TestMessageRejectsMultiMSH(t *testing.T) {
	multi := "MSH|^~\\&|||||20081231||ADT^A01^ADT_A01|12345||2.7\rEVN||20081231\rMSH|^~\\&|||||20081231||ADT^A01^ADT_A01|12345||2.7\rEVN||20081231"
	_, err := builder.NewMessage(builder.MessageOptions{Text: multi})
	if err == nil || err.Error() != "Multiple MSH segments found. Use Batch." {
		t.Fatalf("err = %v, want multi-MSH rejection", err)
	}
}

func TestMultiMSHWithoutBHSParsesAsBatch(t *testing.T) {
	multi := "MSH|^~\\&|||||20081231||ADT^A01^ADT_A01|12345||2.7\rEVN||20081231\rMSH|^~\\&|||||20081231||ADT^A01^ADT_A01|12345||2.7\rEVN||20081231"
	b, err := builder.NewBatch(builder.BatchOptions{Text: multi})
	if err != nil {
		t.Fatal(err)
	}
	messages := b.Messages()
	if len(messages) != 2 {
		t.Fatalf("messages = %d, want 2", len(messages))
	}
	for _, m := range messages {
		count := 0
		m.Get("EVN").ForEach(func(seg builder.HL7Node, _ int) { count++ })
		if count != 1 {
			t.Fatalf("EVN count = %d, want 1", count)
		}
	}
}

func TestMessageHelpers(t *testing.T) {
	t.Run("totalSegment counts segments by name", func(t *testing.T) {
		m, _ := builder.NewMessage(builder.MessageOptions{MessageHeader: mshHeader("X")})
		_, _ = m.AddSegment("EVN")
		_, _ = m.AddSegment("EVN")
		_, _ = m.AddSegment("PID")
		if m.TotalSegment("EVN") != 2 || m.TotalSegment("PID") != 1 || m.TotalSegment("NOPE") != 0 {
			t.Fatalf("totals EVN=%d PID=%d NOPE=%d", m.TotalSegment("EVN"), m.TotalSegment("PID"), m.TotalSegment("NOPE"))
		}
	})

	t.Run("getFirstSegment / getLastSegment return the ends", func(t *testing.T) {
		m, _ := builder.NewMessage(builder.MessageOptions{MessageHeader: mshHeader("X")})
		_, _ = m.AddSegment("EVN")
		last, _ := m.AddSegment("PID")
		if m.GetFirstSegment().Name() != "MSH" {
			t.Fatalf("first = %q", m.GetFirstSegment().Name())
		}
		if m.GetLastSegment() != last {
			t.Fatalf("last segment mismatch")
		}
	})

	t.Run("addSegment with empty path errors", func(t *testing.T) {
		m, _ := builder.NewMessage(builder.MessageOptions{MessageHeader: mshHeader("X")})
		if _, err := m.AddSegment(""); err == nil || !strings.Contains(err.Error(), "Missing segment path") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("addSegment with a multi-segment path errors", func(t *testing.T) {
		m, _ := builder.NewMessage(builder.MessageOptions{MessageHeader: mshHeader("X")})
		if _, err := m.AddSegment("PID.1"); err == nil || !strings.Contains(err.Error(), "Invalid segment") {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestFileBatchNormalizesLFToCR(t *testing.T) {
	lf := "MSH|^~\\&|||||20081231||ADT^A01^ADT_A01|12345||2.7\nEVN||20081231"
	f, err := builder.NewFileBatch(builder.FileOptions{FileBuffer: []byte(lf)})
	if err != nil {
		t.Fatal(err)
	}
	raw := f.String()
	if strings.Contains(raw, `\r`) {
		t.Fatalf("raw contains literal \\r: %q", raw)
	}
	if !strings.Contains(raw, "\r") {
		t.Fatalf("raw missing CR: %q", raw)
	}
}
