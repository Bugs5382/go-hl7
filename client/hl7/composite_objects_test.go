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
	"regexp"
	"testing"
	"time"

	"github.com/Bugs5382/go-hl7/client/hl7"
	"github.com/Bugs5382/go-hl7/client/hl7/metadata"
)

// These tests mirror the hl7.composite-objects.test.ts: composite HL7
// fields (XAD, XPN) accept typed component objects in addition to pre-formatted
// ^-delimited strings, the composer validates each component (R/W/X, length),
// and object/string forms produce byte-identical wire output. The typed
// component objects are modeled as the same Props maps the builder accepts.

func compositeBuilder() *hl7.Builder {
	b := hl7.New(hl7.V2_8)
	b.On("error", func(string) {})
	b.BuildMSH(hl7.Props{
		"msh_10": "X", "msh_11_1": "P",
		"msh_7":   time.Date(2024, 1, 15, 10, 20, 30, 0, time.UTC),
		"msh_9_1": "ADT", "msh_9_2": "A01",
	})
	return b
}

func TestCompositeObjectInputs(t *testing.T) {
	t.Run("PID.11 (XAD) accepts a typed object with camelCase keys", func(t *testing.T) {
		b := compositeBuilder()
		addr := map[string]any{
			"city": "Springfield", "stateOrProvince": "IL",
			"streetAddress": "123 Elm St", "zipOrPostalCode": "62701",
		}
		b.BuildPID(hl7.Props{"pid_11": addr, "pid_3": "MRN1", "pid_5": "DOE^JANE"})
		contains(t, b.String(), "123 Elm St^^Springfield^IL^62701")
	})

	t.Run("camelCase object and pre-formatted string produce identical output", func(t *testing.T) {
		obj := compositeBuilder()
		obj.BuildPID(hl7.Props{
			"pid_11": map[string]any{
				"city": "Springfield", "stateOrProvince": "IL",
				"streetAddress": "123 Elm St", "zipOrPostalCode": "62701",
			},
			"pid_3": "MRN1", "pid_5": "DOE^JANE",
		})
		str := compositeBuilder()
		str.BuildPID(hl7.Props{"pid_11": "123 Elm St^^Springfield^IL^62701", "pid_3": "MRN1", "pid_5": "DOE^JANE"})
		if obj.String() != str.String() {
			t.Fatalf("object %q != string %q", obj.String(), str.String())
		}
	})

	t.Run("PID.5 (XPN) accepts numeric and camelCase keys", func(t *testing.T) {
		b := compositeBuilder()
		name := map[string]any{"givenName": "JANE", "xpn_1": "DOE", "xpn_3": "M"}
		b.BuildPID(hl7.Props{"pid_3": "MRN1", "pid_5": name})
		contains(t, b.String(), "PID|||MRN1||DOE^JANE^M")
	})

	t.Run("trailing empty components are trimmed", func(t *testing.T) {
		b := compositeBuilder()
		b.BuildPID(hl7.Props{
			"pid_11": map[string]any{"city": "Springfield", "streetAddress": "123 Elm St"},
			"pid_3":  "MRN1", "pid_5": "DOE^JANE",
		})
		contains(t, b.String(), "123 Elm St^^Springfield")
		if regexp.MustCompile(`123 Elm St\^\^Springfield\^{15,}`).MatchString(b.String()) {
			t.Fatalf("trailing carets not trimmed: %q", b.String())
		}
	})

	t.Run("max-length violation on a component is rejected", func(t *testing.T) {
		b := compositeBuilder()
		expectThrows(t, "", func() {
			b.BuildPID(hl7.Props{
				"pid_11": map[string]any{"country": "UNITED_STATES_OF_AMERICA", "streetAddress": "123 Elm St"},
				"pid_3":  "MRN1", "pid_5": "DOE^JANE",
			})
		})
	})

	t.Run("withdrawn component (W) is rejected when set", func(t *testing.T) {
		b := compositeBuilder()
		expectThrows(t, "withdrawn", func() {
			b.BuildPID(hl7.Props{
				"pid_3": "MRN1",
				"pid_5": map[string]any{"xpn_1": "DOE", "xpn_10": "shouldNotSet", "xpn_2": "JANE"},
			})
		})
	})

	t.Run("primitive fields still accept plain strings", func(t *testing.T) {
		b := compositeBuilder()
		b.BuildPID(hl7.Props{"pid_3": "MRN1", "pid_5": "DOE^JANE", "pid_7": "19800101120000"})
		contains(t, b.String(), "19800101")
	})

	t.Run("DataTypes catalogue exposes composite layout", func(t *testing.T) {
		if len(metadata.DataTypes["XAD"]) < 20 {
			t.Fatalf("XAD has %d components, want >= 20", len(metadata.DataTypes["XAD"]))
		}
		if len(metadata.DataTypes["XPN"]) < 10 {
			t.Fatalf("XPN has %d components, want >= 10", len(metadata.DataTypes["XPN"]))
		}
		var xad1 *metadata.ComponentSpec
		for i := range metadata.DataTypes["XAD"] {
			if metadata.DataTypes["XAD"][i].Num == 1 {
				xad1 = &metadata.DataTypes["XAD"][i]
			}
		}
		if xad1 == nil || xad1.Name != "Street Address" {
			t.Fatalf("XAD.1 = %+v, want Street Address", xad1)
		}
	})
}
