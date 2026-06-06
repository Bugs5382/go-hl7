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
	"testing"
	"time"

	"github.com/Bugs5382/go-hl7/client/hl7"
)

// These tests cover the version-aware HL7 value-table enforcement: every
// table-bound field (and composite component) is validated against the real
// HL7 value set for the active version, out-of-table values being a hard
// validation error. This goes beyond the original hand-wired tables: the full
// HL7-defined set generated from Caristix is enforced.

func tblMSH(b *hl7.Builder, props hl7.Props) *hl7.Builder {
	b.On("error", func(string) {})
	b.BuildMSH(props)
	return b
}

func v21WithMSH() *hl7.Builder {
	return tblMSH(hl7.New(hl7.V2_1), hl7.Props{"msh_10": "X", "msh_11": "P", "msh_7": time.Now(), "msh_9": "ACK"})
}

func v28WithMSH() *hl7.Builder {
	return tblMSH(hl7.New(hl7.V2_8), hl7.Props{"msh_10": "X", "msh_11_1": "P", "msh_7": time.Now(), "msh_9_1": "ADT", "msh_9_2": "A01"})
}

func TestTableEnforcementVersionAware(t *testing.T) {
	// OBX.2 value type is HL7 table 0125. "NM" (numeric) is part of the v2.8
	// value set but not the v2.1 one: the same code is accepted in one version
	// and rejected in the other.
	t.Run("code valid in 2.8 is accepted there", func(t *testing.T) {
		b := v28WithMSH()
		b.BuildOBX(hl7.Props{"obx_1": "1", "obx_2": "NM", "obx_3": "GLU^Glucose^L", "obx_5": "98", "obx_11": "F"})
		contains(t, b.String(), "\rOBX|1|NM|")
	})

	t.Run("same code absent in 2.1 is rejected", func(t *testing.T) {
		b := v21WithMSH()
		expectThrows(t, "must be one of", func() {
			b.BuildOBX(hl7.Props{"obx_1": "1", "obx_2": "NM", "obx_3": "GLU^Glucose^L", "obx_5": "98", "obx_11": "F"})
		})
	})

	t.Run("known-good code passes (OBX.2 ST in 2.1)", func(t *testing.T) {
		b := v21WithMSH()
		b.BuildOBX(hl7.Props{"obx_1": "1", "obx_2": "ST", "obx_3": "GLU^Glucose^L", "obx_5": "x", "obx_11": "F"})
		contains(t, b.String(), "\rOBX|1|ST|")
	})
}

func TestTableEnforcementSpotChecks(t *testing.T) {
	// 0001 Sex on PID.8.
	t.Run("0001 sex valid", func(t *testing.T) {
		b := v21WithMSH()
		b.BuildPID(hl7.Props{"pid_3": "MRN1", "pid_5": "DOE^JANE", "pid_8": "F"})
		contains(t, b.String(), "DOE^JANE|||F")
	})
	t.Run("0001 sex invalid", func(t *testing.T) {
		b := v21WithMSH()
		expectThrows(t, "must be one of", func() {
			b.BuildPID(hl7.Props{"pid_3": "MRN1", "pid_5": "DOE^JANE", "pid_8": "Z"})
		})
	})

	// 0003 Event Type on EVN.1.
	t.Run("0003 event type valid", func(t *testing.T) {
		b := v21WithMSH()
		b.BuildEVN(hl7.Props{"evn_1": "A01"})
		contains(t, b.String(), "\rEVN|A01")
	})
	t.Run("0003 event type invalid", func(t *testing.T) {
		b := v21WithMSH()
		expectThrows(t, "must be one of", func() {
			b.BuildEVN(hl7.Props{"evn_1": "ZZZ"})
		})
	})

	// 0125 Value Type on OBX.2.
	t.Run("0125 value type invalid", func(t *testing.T) {
		b := v28WithMSH()
		expectThrows(t, "must be one of", func() {
			b.BuildOBX(hl7.Props{"obx_1": "1", "obx_2": "NOTATYPE", "obx_5": "x", "obx_11": "F"})
		})
	})
}

func TestTableEnforcementComponentLevel(t *testing.T) {
	// PID.11 (XAD) component 7 "Address Type" is bound to HL7 table 0190, which
	// has a value set in v2.8. The composite-object assembly enforces it.
	t.Run("component table value valid", func(t *testing.T) {
		b := v28WithMSH()
		b.BuildPID(hl7.Props{
			"pid_3": "MRN1", "pid_5": "DOE^JANE",
			"pid_11": map[string]any{"streetAddress": "123 Elm St", "city": "Springfield", "addressType": "H"},
		})
		contains(t, b.String(), "123 Elm St^^Springfield")
	})

	t.Run("component table value invalid is rejected", func(t *testing.T) {
		b := v28WithMSH()
		expectThrows(t, "must be one of", func() {
			b.BuildPID(hl7.Props{
				"pid_3": "MRN1", "pid_5": "DOE^JANE",
				"pid_11": map[string]any{"streetAddress": "123 Elm St", "city": "Springfield", "addressType": "ZZ"},
			})
		})
	})
}

func TestTableEnforcementAbsentTableNotEnforced(t *testing.T) {
	// PID.15 (Primary Language) is bound to HL7 table 0296, which carries no
	// fixed value set in v2.1. With no values to check against, any value is
	// accepted (no enforcement).
	b := v21WithMSH()
	b.BuildPID(hl7.Props{"pid_3": "MRN1", "pid_5": "DOE^JANE", "pid_15": "anything-goes"})
	contains(t, b.String(), "anything-goes")
}
