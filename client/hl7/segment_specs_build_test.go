// MIT License
//
// Copyright (c) 2026 Shane
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package hl7_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Bugs5382/go-hl7/client/hl7"
)

// These tests mirror the buildSegment half of the reference's
// hl7.segment-specs.test.ts: the generic spec-driven BuildSegment across
// versions. (The catalogue-coverage half lives in the metadata package test.)

func specHeader() hl7.Props {
	return hl7.Props{
		"msh_10": "X", "msh_11_1": "P",
		"msh_7":   time.Date(2024, 1, 15, 10, 20, 30, 0, time.UTC),
		"msh_9_1": "ADT", "msh_9_2": "A01",
	}
}

func TestBuildSegmentGeneric(t *testing.T) {
	t.Run("v2.4 builds ECD with required fields", func(t *testing.T) {
		b := hl7.NewHL7_2_4()
		b.On("error", func(string) {})
		b.BuildMSH(specHeader())
		b.BuildSegment("ECD", hl7.Props{"ecd_1": "1", "ecd_2": "RC"})
		contains(t, b.String(), "ECD|1|RC")
	})

	t.Run("v2.4 to v2.8 chained build with multiple segments", func(t *testing.T) {
		b := hl7.NewHL7_2_8()
		b.On("error", func(string) {})
		out := b.BuildMSH(specHeader()).BuildSegment("ECD", hl7.Props{"ecd_1": "1", "ecd_2": "RC"}).String()
		contains(t, out, "MSH|")
		contains(t, out, "ECD|")
	})

	t.Run("v2.8 rejects ECD.4 (W) via buildSegment", func(t *testing.T) {
		b := hl7.NewHL7_2_8()
		b.On("error", func(string) {})
		b.BuildMSH(specHeader())
		expectThrows(t, "", func() {
			b.BuildSegment("ECD", hl7.Props{"ecd_1": "1", "ecd_2": "RC", "ecd_4": "20260101"})
		})
	})

	t.Run("v2.3.1 rejects ECD entirely (segment not in version)", func(t *testing.T) {
		b := hl7.NewHL7_2_3_1()
		b.On("error", func(string) {})
		b.BuildMSH(specHeader())
		expectThrows(t, "not part of HL7 v2.3.1", func() {
			b.BuildSegment("ECD", hl7.Props{"ecd_1": "1"})
		})
	})

	t.Run("v2.1 rejects ECD entirely (introduced in v2.4)", func(t *testing.T) {
		b := hl7.NewHL7_2_1()
		b.On("error", func(string) {})
		b.BuildMSH(hl7.Props{
			"msh_10": "X", "msh_11": "T", "msh_3": "APP", "msh_5": "RECV_APP",
			"msh_7": time.Date(2024, 1, 15, 10, 20, 30, 0, time.UTC), "msh_9": "ACK",
		})
		expectThrows(t, "not part of HL7 v2.1", func() {
			b.BuildSegment("ECD", hl7.Props{"ecd_1": "1"})
		})
	})

	t.Run("buildSegment refuses unknown segment names", func(t *testing.T) {
		b := hl7.NewHL7_2_8()
		b.On("error", func(string) {})
		b.BuildMSH(specHeader())
		expectThrows(t, "Unknown HL7 segment", func() { b.BuildSegment("XYZ", hl7.Props{}) })
	})

	t.Run("buildSegment refuses MSH", func(t *testing.T) {
		b := hl7.NewHL7_2_8()
		b.On("error", func(string) {})
		b.BuildMSH(specHeader())
		expectThrows(t, "Use buildMSH", func() { b.BuildSegment("MSH", hl7.Props{}) })
	})

	t.Run("buildSegment is chainable (returns this)", func(t *testing.T) {
		b := hl7.NewHL7_2_4()
		b.On("error", func(string) {})
		b.BuildMSH(specHeader())
		rv := b.BuildSegment("ECD", hl7.Props{"ecd_1": "1", "ecd_2": "RC"})
		if rv != b {
			t.Fatal("expected BuildSegment to return the receiver")
		}
	})

	t.Run("v2.6 emits B warning for ECD.4", func(t *testing.T) {
		b := hl7.NewHL7_2_6()
		var warned string
		b.On("warning", func(m string) { warned += m })
		b.On("error", func(string) {})
		b.BuildMSH(specHeader())
		b.BuildSegment("ECD", hl7.Props{"ecd_1": "1", "ecd_2": "RC", "ecd_4": "20260101"})
		if !strings.Contains(warned, "deprecated") {
			t.Fatalf("expected deprecated warning, got %q", warned)
		}
	})

	t.Run("v2.7 rejects ECD.4 (W) — withdrawn earlier than v2.8", func(t *testing.T) {
		b := hl7.NewHL7_2_7()
		b.On("error", func(string) {})
		b.BuildMSH(specHeader())
		expectThrows(t, "withdrawn in HL7 v2.7", func() {
			b.BuildSegment("ECD", hl7.Props{"ecd_1": "1", "ecd_2": "RC", "ecd_4": "20260101"})
		})
	})

	t.Run("buildSegment accepts numeric-string keys", func(t *testing.T) {
		b := hl7.NewHL7_2_4()
		b.On("error", func(string) {})
		b.BuildMSH(specHeader())
		b.BuildSegment("ECD", hl7.Props{"1": "1", "2": "RC"})
		contains(t, b.String(), "ECD|1|RC")
	})
}
