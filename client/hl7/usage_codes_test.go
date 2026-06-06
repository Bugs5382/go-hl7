package hl7

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
	"errors"
	"strings"
	"testing"

	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/hl7/metadata"
)

// testBuilder mirrors the usage-codes test's TestBuilder: a v2.6 builder with a
// no-op "error" listener and a helper that lazily creates the synthetic segment
// before invoking the spec-driven validator.
func newTestBuilder() *HL7_BASE {
	b := newVersion("2.6", nil)
	b.On("error", func(string) {})
	return b
}

func (b *HL7_BASE) callSetField(spec metadata.SegmentSpec, num int, value any) []string {
	if b.segment == nil || b.segment.Name() != spec.Name {
		b.segment = mustAddSegment(b.message, spec.Name)
	}
	return b.validatorSetField(spec, num, value, nil)
}

// tstSpec is the synthetic TST segment spec from the spec usage-codes test.
func tstSpec() metadata.SegmentSpec {
	return metadata.SegmentSpec{
		Name:        "TST",
		Description: "Test",
		Versions:    []metadata.HL7Version{"2.4", "2.5", "2.5.1", "2.6", "2.7"},
		Fields: []metadata.FieldSpec{
			{Num: 1, Name: "Required Field", Usage: map[metadata.HL7Version]metadata.HL7UsageCode{"2.6": "R"}},
			{Num: 2, Name: "Optional Field", Usage: map[metadata.HL7Version]metadata.HL7UsageCode{"2.6": "O"}},
			{Num: 3, Name: "Backward Compat", Length: lenMinMax(1, 5), Usage: map[metadata.HL7Version]metadata.HL7UsageCode{"2.6": "B"}},
			{Num: 4, Name: "Withdrawn", Usage: map[metadata.HL7Version]metadata.HL7UsageCode{"2.6": "W"}},
			{Num: 5, Name: "Not Supported", Usage: map[metadata.HL7Version]metadata.HL7UsageCode{"2.6": "X"}},
			{Num: 6, Name: "Conditional", DependsOn: &metadata.Depends{MustBeSet: true, Path: "1"}, Usage: map[metadata.HL7Version]metadata.HL7UsageCode{"2.6": "D"}},
			{Num: 7, Name: "Future Field", Usage: map[metadata.HL7Version]metadata.HL7UsageCode{"2.7": "O"}},
		},
	}
}

// recoverValidation runs fn and reports whether it panicked with an
// HL7ValidationError whose message matches want (substring), mirroring the
// toThrow assertions.
func expectValidationPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic containing %q, got none", want)
		}
		err, ok := r.(error)
		if !ok {
			t.Fatalf("expected error panic, got %T", r)
		}
		if !errors.Is(err, helpers.ErrValidation) {
			t.Fatalf("expected HL7ValidationError, got %v", err)
		}
		if want != "" && !strings.Contains(err.Error(), want) {
			t.Fatalf("expected message containing %q, got %q", want, err.Error())
		}
	}()
	fn()
}

func expectNoPanic(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected no panic, got %v", r)
		}
	}()
	fn()
}

func TestUsageCodes(t *testing.T) {
	SPEC := tstSpec()

	t.Run("R required field unset throws HL7ValidationError", func(t *testing.T) {
		b := newTestBuilder()
		expectValidationPanic(t, "", func() { b.callSetField(SPEC, 1, nil) })
	})

	t.Run("R required field with value succeeds", func(t *testing.T) {
		b := newTestBuilder()
		expectNoPanic(t, func() { b.callSetField(SPEC, 1, "OK") })
	})

	t.Run("O optional field unset is fine", func(t *testing.T) {
		b := newTestBuilder()
		if got := b.callSetField(SPEC, 2, nil); len(got) != 0 {
			t.Fatalf("expected empty, got %v", got)
		}
	})

	t.Run("B backward-compat field warns but still serializes", func(t *testing.T) {
		b := newTestBuilder()
		var warning string
		b.On("warning", func(m string) { warning = m })
		errs := b.callSetField(SPEC, 3, "abc")
		if len(errs) == 0 {
			t.Fatal("expected at least one warning entry")
		}
		if !strings.Contains(warning, "deprecated") {
			t.Fatalf("expected deprecated warning, got %q", warning)
		}
		if !strings.Contains(b.ToMessage().String(), "abc") {
			t.Fatalf("expected serialized value, got %q", b.ToMessage().String())
		}
	})

	t.Run("W withdrawn field throws even with hardError false", func(t *testing.T) {
		b := newTestBuilder()
		expectValidationPanic(t, "withdrawn in HL7 v2.6", func() { b.callSetField(SPEC, 4, "boom") })
	})

	t.Run("X not-supported field throws", func(t *testing.T) {
		b := newTestBuilder()
		expectValidationPanic(t, "not supported in HL7 v2.6", func() { b.callSetField(SPEC, 5, "boom") })
	})

	t.Run("D conditional field dependsOn unmet throws", func(t *testing.T) {
		b := newTestBuilder()
		expectValidationPanic(t, "", func() { b.callSetField(SPEC, 6, "value") })
	})

	t.Run("D conditional field dependsOn satisfied succeeds", func(t *testing.T) {
		b := newTestBuilder()
		b.callSetField(SPEC, 1, "anchor")
		expectNoPanic(t, func() { b.callSetField(SPEC, 6, "value") })
	})

	t.Run("field not present in this version throws when set", func(t *testing.T) {
		b := newTestBuilder()
		expectValidationPanic(t, "not available in HL7 v2.6", func() { b.callSetField(SPEC, 7, "future") })
	})

	t.Run("field not present in this version plus no value is a no-op", func(t *testing.T) {
		b := newTestBuilder()
		if got := b.callSetField(SPEC, 7, nil); len(got) != 0 {
			t.Fatalf("expected empty, got %v", got)
		}
	})
}
