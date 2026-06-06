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
	"time"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/utils"
)

// mshHeader mirrors the MSH_HEADER fixture (msh_9_1: ADT, msh_9_2: A01,
// msh_11_1: D).
func mshHeader(controlID string) *builder.MessageHeader {
	return (&builder.MessageHeader{
		MSH9_1: "ADT",
		MSH9_2: "A01",
	}).WithMSH10(controlID).WithMSH11_1("D")
}

func newControlMessage(t *testing.T, controlID, mshDate string) *builder.Message {
	t.Helper()
	h := mshHeader(controlID)
	m, err := builder.NewMessage(builder.MessageOptions{MessageHeader: h})
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	m.Set("MSH.7", mshDate)
	return m
}

// --- basic batch basics ---

func TestBatchStartEndProducesBHSAndBTS(t *testing.T) {
	b, err := builder.NewBatch(builder.BatchOptions{})
	if err != nil {
		t.Fatal(err)
	}
	b.Start("")
	b.End()
	if got := b.String(); !strings.Contains(got, `BHS|^~\&`) {
		t.Fatalf("missing BHS framing: %q", got)
	}
	if got := b.String(); !strings.Contains(got, "BTS|0") {
		t.Fatalf("missing BTS|0: %q", got)
	}
}

func TestBatchBHS7Reflected(t *testing.T) {
	b, _ := builder.NewBatch(builder.BatchOptions{})
	b.Start("")
	b.Set("BHS.7", "20081231")
	b.End()
	if got := b.Get("BHS.7").String(); got != "20081231" {
		t.Fatalf("BHS.7 = %q", got)
	}
	if got := b.String(); got != "BHS|^~\\&|||||20081231\rBTS|0" {
		t.Fatalf("toString = %q", got)
	}
}

func TestBatchBHS3Through6(t *testing.T) {
	b, _ := builder.NewBatch(builder.BatchOptions{})
	b.Start("")
	b.Set("BHS.7", "20081231")
	b.Set("BHS.3", "SendingApp")
	b.Set("BHS.4", "SendingFacility")
	b.Set("BHS.5", "ReceivingApp")
	b.Set("BHS.6", "ReceivingFacility")
	b.End()

	checks := map[string]string{
		"BHS.3": "SendingApp",
		"BHS.4": "SendingFacility",
		"BHS.5": "ReceivingApp",
		"BHS.6": "ReceivingFacility",
	}
	for path, want := range checks {
		if got := b.Get(path).String(); got != want {
			t.Fatalf("%s = %q want %q", path, got, want)
		}
	}
	want := "BHS|^~\\&|SendingApp|SendingFacility|ReceivingApp|ReceivingFacility|20081231\rBTS|0"
	if got := b.String(); got != want {
		t.Fatalf("toString = %q want %q", got, want)
	}
}

// --- complex builder batch ---

func TestBatchAddSingleMessage(t *testing.T) {
	date := utils.CreateHL7Date(time.Now(), "8")
	b, _ := builder.NewBatch(builder.BatchOptions{})
	b.Start("")
	b.Set("BHS.7", date)

	m := newControlMessage(t, "CONTROL_ID", date)
	b.Add(m, -1)
	b.End()

	want := strings.Join([]string{
		`BHS|^~\&|||||` + date,
		`MSH|^~\&|||||` + date + `||ADT^A01^ADT_A01|CONTROL_ID|D|2.7`,
		"BTS|1",
	}, "\r")
	if got := b.String(); got != want {
		t.Fatalf("toString = %q want %q", got, want)
	}
}

func TestBatchBTSCounterTen(t *testing.T) {
	date := utils.CreateHL7Date(time.Now(), "8")
	b, _ := builder.NewBatch(builder.BatchOptions{})
	b.Start("")
	b.Set("BHS.7", date)

	m := newControlMessage(t, "CONTROL_ID", date)
	for i := 0; i < 10; i++ {
		b.Add(m, -1)
	}
	b.End()
	if got := b.String(); !strings.Contains(got, "BTS|10") {
		t.Fatalf("missing BTS|10: %q", got)
	}
}

func TestBatchPreservesAdditionalSegments(t *testing.T) {
	date := utils.CreateHL7Date(time.Now(), "8")
	b, _ := builder.NewBatch(builder.BatchOptions{})
	b.Start("")
	b.Set("BHS.7", date)

	m := newControlMessage(t, "CONTROL_ID", "20231208")
	seg, err := m.AddSegment("EVN")
	if err != nil {
		t.Fatal(err)
	}
	seg.Set("2", "20081231")

	b.Add(m, -1)
	b.End()
	want := "BHS|^~\\&|||||" + date + "\rMSH|^~\\&|||||20231208||ADT^A01^ADT_A01|CONTROL_ID|D|2.7\rEVN||20081231\rBTS|1"
	if got := b.String(); got != want {
		t.Fatalf("toString = %q want %q", got, want)
	}
}

func TestBatchPreservesRepeatingEVN(t *testing.T) {
	date := utils.CreateHL7Date(time.Now(), "8")
	b, _ := builder.NewBatch(builder.BatchOptions{})
	b.Start("")
	b.Set("BHS.7", date)

	m := newControlMessage(t, "CONTROL_ID", "20231208")
	seg, _ := m.AddSegment("EVN")
	seg.Set("2", "20081231")
	seg, _ = m.AddSegment("EVN")
	seg.Set("2", "20081231")

	b.Add(m, -1)
	b.End()
	want := "BHS|^~\\&|||||" + date + "\rMSH|^~\\&|||||20231208||ADT^A01^ADT_A01|CONTROL_ID|D|2.7\rEVN||20081231\rEVN||20081231\rBTS|1"
	if got := b.String(); got != want {
		t.Fatalf("toString = %q want %q", got, want)
	}
}
