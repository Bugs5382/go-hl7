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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/Bugs5382/go-hl7/client/builder"
)

func newFileBatchStarted(t *testing.T) *builder.FileBatch {
	t.Helper()
	f, err := builder.NewFileBatch(builder.FileOptions{})
	if err != nil {
		t.Fatal(err)
	}
	f.Start()
	f.Set("FHS.7", "20081231")
	return f
}

// --- basic file basics ---

func TestFileBatchWrapsMessage(t *testing.T) {
	f := newFileBatchStarted(t)
	m := newControlMessage(t, "CONTROL_ID", "20081231")
	f.AddMessage(m)
	f.End()

	want := strings.Join([]string{
		`FHS|^~\&|||||20081231`,
		`MSH|^~\&|||||20081231||ADT^A01^ADT_A01|CONTROL_ID|D|2.7`,
		"FTS|1",
	}, "\r")
	if got := f.String(); got != want {
		t.Fatalf("toString = %q want %q", got, want)
	}
}

func TestFileBatchTenMessages(t *testing.T) {
	f := newFileBatchStarted(t)
	for i := 0; i < 10; i++ {
		m := newControlMessage(t, "CONTROL_ID"+strconv.Itoa(i+1), "20081231")
		f.AddMessage(m)
	}
	f.End()

	lines := []string{`FHS|^~\&|||||20081231`}
	for i := 0; i < 10; i++ {
		lines = append(lines, `MSH|^~\&|||||20081231||ADT^A01^ADT_A01|CONTROL_ID`+strconv.Itoa(i+1)+`|D|2.7`)
	}
	lines = append(lines, "FTS|10")
	want := strings.Join(lines, "\r")
	if got := f.String(); got != want {
		t.Fatalf("toString = %q want %q", got, want)
	}
}

func TestFileBatchAddBatch(t *testing.T) {
	f := newFileBatchStarted(t)
	b, _ := builder.NewBatch(builder.BatchOptions{})
	b.Start("")
	b.Set("BHS.7", "20081231")
	b.End()
	if err := f.AddBatch(b); err != nil {
		t.Fatalf("AddBatch: %v", err)
	}
	f.End()

	want := strings.Join([]string{
		`FHS|^~\&|||||20081231`,
		`BHS|^~\&|||||20081231`,
		"BTS|0",
		"FTS|1",
	}, "\r")
	if got := f.String(); got != want {
		t.Fatalf("toString = %q want %q", got, want)
	}
}

func TestFileBatchWrapsBatchWithMessage(t *testing.T) {
	f := newFileBatchStarted(t)
	b, _ := builder.NewBatch(builder.BatchOptions{})
	b.Start("")
	b.Set("BHS.7", "20081231")
	m := newControlMessage(t, "CONTROL_ID", "20081231")
	b.Add(m, -1)
	b.End()
	if err := f.AddBatch(b); err != nil {
		t.Fatalf("AddBatch: %v", err)
	}
	f.End()

	want := strings.Join([]string{
		`FHS|^~\&|||||20081231`,
		`BHS|^~\&|||||20081231`,
		`MSH|^~\&|||||20081231||ADT^A01^ADT_A01|CONTROL_ID|D|2.7`,
		"BTS|1",
		"FTS|1",
	}, "\r")
	if got := f.String(); got != want {
		t.Fatalf("toString = %q want %q", got, want)
	}
}

func TestFileBatchRoutesMessagesIntoOpenBatch(t *testing.T) {
	f := newFileBatchStarted(t)
	b, _ := builder.NewBatch(builder.BatchOptions{})
	b.Start("")
	b.Set("BHS.7", "20081231")
	b.Add(newControlMessage(t, "CONTROL_ID1", "20081231"), -1)
	b.End()
	if err := f.AddBatch(b); err != nil {
		t.Fatalf("AddBatch: %v", err)
	}

	f.AddMessage(newControlMessage(t, "CONTROL_ID2", "20081231"))
	f.AddMessage(newControlMessage(t, "CONTROL_ID3", "20081231"))
	f.End()

	want := strings.Join([]string{
		`FHS|^~\&|||||20081231`,
		`BHS|^~\&|||||20081231`,
		`MSH|^~\&|||||20081231||ADT^A01^ADT_A01|CONTROL_ID1|D|2.7`,
		`MSH|^~\&|||||20081231||ADT^A01^ADT_A01|CONTROL_ID2|D|2.7`,
		`MSH|^~\&|||||20081231||ADT^A01^ADT_A01|CONTROL_ID3|D|2.7`,
		"BTS|3",
		"FTS|1",
	}, "\r")
	if got := f.String(); got != want {
		t.Fatalf("toString = %q want %q", got, want)
	}
}

func TestFileBatchTwoConsecutiveBatches(t *testing.T) {
	f := newFileBatchStarted(t)

	b1, _ := builder.NewBatch(builder.BatchOptions{})
	b1.Start("")
	b1.Set("BHS.7", "20081231")
	b1.End()
	if err := f.AddBatch(b1); err != nil {
		t.Fatalf("AddBatch: %v", err)
	}

	b2, _ := builder.NewBatch(builder.BatchOptions{})
	b2.Start("")
	b2.Set("BHS.7", "20081231")
	b2.End()
	if err := f.AddBatch(b2); err != nil {
		t.Fatalf("AddBatch: %v", err)
	}
	f.End()

	want := strings.Join([]string{
		`FHS|^~\&|||||20081231`,
		`BHS|^~\&|||||20081231`,
		"BTS|0",
		`BHS|^~\&|||||20081231`,
		"BTS|0",
		"FTS|2",
	}, "\r")
	if got := f.String(); got != want {
		t.Fatalf("toString = %q want %q", got, want)
	}
}

// --- complex file generation (disk) ---

func TestFileBatchCreateFile(t *testing.T) {
	dir := t.TempDir()
	f, _ := builder.NewFileBatch(builder.FileOptions{Location: dir})
	f.Start()
	f.Set("FHS.7", "20081231")
	f.End()
	if err := f.CreateFile("HELLO"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "hl7.HELLO.20081231.hl7")); err != nil {
		t.Fatalf("file not written: %v", err)
	}
}

func TestMessageToFile(t *testing.T) {
	dir := t.TempDir()
	m := newControlMessage(t, "CONTROL_ID", "20081231")
	name, err := m.ToFile("ADT", true, dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
		t.Fatalf("file not written: %v", err)
	}
}

func TestBatchToFile(t *testing.T) {
	dir := t.TempDir()
	b, _ := builder.NewBatch(builder.BatchOptions{})
	b.Start("")
	b.Set("BHS.7", "20081231")
	b.End()
	name, err := b.ToFile("ADTs", true, dir, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
		t.Fatalf("file not written: %v", err)
	}
}

// --- file tests (read back) ---

func TestFileBatchReadsAndYieldsMessages(t *testing.T) {
	dir := t.TempDir()
	hl7String := "MSH|^~\\&|||||20081231||ADT^A01^ADT_A01|12345||2.7\rEVN||20081231"
	hl7Batch := "BHS|^~\\&|||||20231208\rMSH|^~\\&|||||20231208||ADT^A01^ADT_A01|CONTROL_ID||2.7\rEVN||20081231\rEVN||20081231\rBTS|1"

	msg, err := builder.NewMessage(builder.MessageOptions{Text: hl7String})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := msg.ToFile("readTestMSH", true, dir, ""); err != nil {
		t.Fatal(err)
	}

	batch, err := builder.NewBatch(builder.BatchOptions{Text: hl7Batch})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := batch.ToFile("readTestBHS", true, dir, ""); err != nil {
		t.Fatal(err)
	}

	t.Run("reads from a fullFilePath", func(t *testing.T) {
		one, err := builder.NewFileBatch(builder.FileOptions{FullFilePath: filepath.Join(dir, "hl7.readTestMSH.20081231.hl7")})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(one.Text(), hl7String) {
			t.Fatalf("text = %q", one.Text())
		}
		two, err := builder.NewFileBatch(builder.FileOptions{FullFilePath: filepath.Join(dir, "hl7.readTestBHS.20231208.hl7")})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(two.Text(), hl7Batch) {
			t.Fatalf("text = %q", two.Text())
		}
	})

	t.Run("reads from a fileBuffer", func(t *testing.T) {
		buf, _ := os.ReadFile(filepath.Join(dir, "hl7.readTestMSH.20081231.hl7"))
		one, err := builder.NewFileBatch(builder.FileOptions{FileBuffer: buf})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(one.Text(), hl7String) {
			t.Fatalf("text = %q", one.Text())
		}
		buf2, _ := os.ReadFile(filepath.Join(dir, "hl7.readTestBHS.20231208.hl7"))
		two, err := builder.NewFileBatch(builder.FileOptions{FileBuffer: buf2})
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(two.Text(), hl7Batch) {
			t.Fatalf("text = %q", two.Text())
		}
	})

	t.Run("messages yields one Message with one EVN", func(t *testing.T) {
		fb, _ := builder.NewFileBatch(builder.FileOptions{FullFilePath: filepath.Join(dir, "hl7.readTestMSH.20081231.hl7")})
		messages := fb.Messages()
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
		if count != 1 {
			t.Fatalf("EVN count = %d", count)
		}
	})

	t.Run("BHS yields one Message with two EVN segments", func(t *testing.T) {
		fb, _ := builder.NewFileBatch(builder.FileOptions{FullFilePath: filepath.Join(dir, "hl7.readTestBHS.20231208.hl7")})
		messages := fb.Messages()
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
	})
}
