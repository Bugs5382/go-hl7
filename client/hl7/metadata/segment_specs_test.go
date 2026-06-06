package metadata

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

// These tests mirror the catalogue-coverage half of the reference's
// hl7.segment-specs.test.ts ("SegmentSpec catalogue coverage"). The generic
// buildSegment runtime cases live with the spec-driven builder (phase 2).

func findField(spec SegmentSpec, num int) (FieldSpec, bool) {
	for _, f := range spec.Fields {
		if f.Num == num {
			return f, true
		}
	}
	return FieldSpec{}, false
}

func TestSegmentSpecCatalogueCoverage(t *testing.T) {
	t.Run("catalogue covers a substantial number of segments", func(t *testing.T) {
		if got := len(SegmentSpecs); got <= 150 {
			t.Fatalf("expected >150 segments, got %d", got)
		}
	})

	t.Run("every spec lists at least one supported version", func(t *testing.T) {
		for name, spec := range SegmentSpecs {
			if len(spec.Versions) == 0 {
				t.Fatalf("segment %s: no versions", name)
			}
		}
	})

	t.Run("every field's usage map only references declared segment versions", func(t *testing.T) {
		for name, spec := range SegmentSpecs {
			declared := map[HL7Version]bool{}
			for _, v := range spec.Versions {
				declared[v] = true
			}
			for _, f := range spec.Fields {
				for v := range f.Usage {
					if !declared[v] {
						t.Fatalf("%s.%d v%s: usage references undeclared version", name, f.Num, v)
					}
				}
			}
		}
	})

	t.Run("ECD spec — ECD.4 is W in v2.8 (the user-flagged canonical case)", func(t *testing.T) {
		ecd := SegmentSpecs["ECD"]
		ecd4, ok := findField(ecd, 4)
		if !ok {
			t.Fatal("ECD.4 not found")
		}
		if ecd4.Usage[V28] != UsageWithdrawn {
			t.Fatalf("ECD.4 v2.8 = %q, want W", ecd4.Usage[V28])
		}
		for _, v := range ecd.Versions {
			if v == V231 || v == V21 {
				t.Fatalf("ECD should not exist in %s", v)
			}
		}
	})

	t.Run("composite fields carry sub-component metadata (PID.11 / XAD)", func(t *testing.T) {
		pid11, ok := findField(SegmentSpecs["PID"], 11)
		if !ok {
			t.Fatal("PID.11 not found")
		}
		if pid11.HL7Type != "XAD" {
			t.Fatalf("PID.11 type = %q, want XAD", pid11.HL7Type)
		}
		if len(pid11.Components) <= 5 {
			t.Fatalf("PID.11 components = %d, want >5", len(pid11.Components))
		}
		byName := func(n string) bool {
			for _, c := range pid11.Components {
				if strings.EqualFold(c.Name, n) {
					return true
				}
			}
			return false
		}
		for _, n := range []string{"Street Address", "City", "State Or Province", "Zip Or Postal Code"} {
			if !byName(n) {
				t.Fatalf("PID.11 missing component %q", n)
			}
		}
	})

	t.Run("primitive fields have no sub-components", func(t *testing.T) {
		pid1, ok := findField(SegmentSpecs["PID"], 1)
		if !ok {
			t.Fatal("PID.1 not found")
		}
		if pid1.HL7Type != "SI" && pid1.HL7Type != "NM" {
			t.Fatalf("PID.1 type = %q, want SI or NM", pid1.HL7Type)
		}
		if len(pid1.Components) != 0 {
			t.Fatalf("PID.1 should have no components, got %d", len(pid1.Components))
		}
	})

	t.Run("composite components track HL7 tables when applicable", func(t *testing.T) {
		pid11, _ := findField(SegmentSpecs["PID"], 11)
		var country *ComponentSpec
		for i := range pid11.Components {
			if pid11.Components[i].Name == "Country" {
				country = &pid11.Components[i]
				break
			}
		}
		if country == nil {
			t.Fatal("PID.11 Country component not found")
		}
		if country.Table != 399 {
			t.Fatalf("Country table = %d, want 399", country.Table)
		}
	})
}

func TestDataTypesCatalogue(t *testing.T) {
	if _, ok := DataTypes["XAD"]; !ok {
		t.Fatal("DataTypes missing XAD")
	}
	if _, ok := DataTypes["XPN"]; !ok {
		t.Fatal("DataTypes missing XPN")
	}
	xad := DataTypes["XAD"]
	if len(xad) < 5 {
		t.Fatalf("XAD components = %d, want >=5", len(xad))
	}
}
