package hl7_test

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

	"github.com/Bugs5382/go-hl7/client/hl7"
)

// These tests mirror the hl7.ecd.test.ts: ECD per-version availability.
// ECD did not exist before v2.4; ECD.4 is O in 2.4-2.5.1, B (deprecated, warns)
// in 2.6-2.7.1, and W (withdrawn, hard reject) in 2.8 — already withdrawn in 2.7
// per the catalog. The compile-time "no buildECD before v2.4" check is a
// TypeScript-only construct; its Go mirror is the runtime version rejection
// (covered by the segment-specs tests).

func ecdHeader(b *hl7.HL7_BASE) {
	b.On("error", func(string) {})
	b.BuildMSH(hl7.Props{
		"msh_10": "X", "msh_11_1": "P",
		"msh_7":   time.Date(2024, 1, 15, 10, 20, 30, 0, time.UTC),
		"msh_9_1": "ADT", "msh_9_2": "A01",
	})
}

func TestECDPerVersion(t *testing.T) {
	t.Run("v2.4 accepts ecd_4", func(t *testing.T) {
		b := hl7.NewHL7_2_4()
		ecdHeader(b)
		b.BuildECD(hl7.Props{"ecd_1": "1", "ecd_2": "RC", "ecd_4": "20260101"})
		contains(t, b.String(), "ECD|1|RC||20260101")
	})

	t.Run("v2.6 accepts ecd_4 but warns deprecated", func(t *testing.T) {
		b := hl7.NewHL7_2_6()
		var warned string
		b.On("warning", func(m string) { warned += m })
		ecdHeader(b)
		b.BuildECD(hl7.Props{"ecd_1": "1", "ecd_2": "RC", "ecd_4": "20260101"})
		if !strings.Contains(warned, "deprecated") {
			t.Fatalf("expected deprecated warning, got %q", warned)
		}
		contains(t, b.String(), "ECD|1|RC||20260101")
	})

	t.Run("v2.8 rejects ecd_4 with a hard error", func(t *testing.T) {
		b := hl7.NewHL7_2_8()
		ecdHeader(b)
		expectThrows(t, "", func() {
			b.BuildECD(hl7.Props{"ecd_1": "1", "ecd_2": "RC", "ecd_4": "20260101"})
		})
	})

	t.Run("v2.8 succeeds when ecd_4 is omitted", func(t *testing.T) {
		b := hl7.NewHL7_2_8()
		ecdHeader(b)
		b.BuildECD(hl7.Props{"ecd_1": "1", "ecd_2": "RC"})
		contains(t, b.String(), "ECD|1|RC")
		if strings.Contains(b.String(), "20260101") {
			t.Fatalf("did not expect 20260101 in %q", b.String())
		}
	})

	t.Run("required ECD.1 unset throws", func(t *testing.T) {
		b := hl7.NewHL7_2_4()
		ecdHeader(b)
		expectThrows(t, "", func() { b.BuildECD(hl7.Props{"ecd_2": "RC"}) })
	})
}
