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
	"strconv"

	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/hl7/metadata"
)

// spec returns a registered SegmentSpec by name. The per-version builders
// reference compile-time-known specs, so an absent name is a fatal mismatch: it
// is recorded into the build chain and a zero spec returned (the downstream
// setField calls then short-circuit on b.err).
func (b *Builder) spec(name string) metadata.SegmentSpec {
	if b.err != nil {
		return metadata.SegmentSpec{}
	}
	s, ok := metadata.SegmentSpecs[name]
	if !ok {
		b.fail(helpers.NewHL7ValidationError("Unknown HL7 segment " + name))
		return metadata.SegmentSpec{}
	}
	return s
}

// notImplementedBefore records HL7FatalError("Not Implemented") when the
// current version predates the version that introduced a typed builder. It
// mirrors the Builder._buildXXX stubs: a segment-builder method is only
// defined from the version that introduced the segment onward, so on an earlier
// version the call falls through to the base stub that throws.
func (b *Builder) notImplementedBefore(introduced string) {
	if b.err != nil {
		return
	}
	if compareVersions(b.version, introduced) < 0 {
		b.fail(helpers.NewHL7FatalError("Not Implemented"))
	}
}

// setField is shorthand for validatorSetField with an optional override rule.
func (b *Builder) setField(s metadata.SegmentSpec, num int, value any, rule *ValidationRule) {
	b.validatorSetField(s, num, value, rule)
}

// BuildSegment builds any segment by name from its generated spec (the
// buildSegment). Chainable.
func (b *Builder) BuildSegment(name string, properties Props) *Builder {
	if b.err != nil {
		return b
	}
	b.buildSegmentGeneric(name, properties)
	return b
}

// BuildADD builds an ADD (Addendum) segment (the buildADD). It must not
// follow MSH/BHS/FHS. Chainable.
func (b *Builder) BuildADD(properties Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	if b.err != nil {
		return b
	}
	last := b.message.GetLastSegment()
	if last != nil && (last.Name() == "BHS" || last.Name() == "FHS" || last.Name() == "MSH") {
		b.fail(helpers.NewHL7ValidationError("This segment must not follow a MSH, BHS, or FHS"))
		return b
	}
	b.segment = b.mustAddSegment("ADD")
	rule := &ValidationRule{}
	if b.hasMaxAddSegment {
		rule.Length = lenMax(b.maxAddSegmentLength)
	}
	b.validatorSetValue("1", pick(properties, "add_1", "addendumContinuationPointer"), rule)
	return b
}

// BuildNCK builds an NCK (System Clock) segment with the version-appropriate
// timestamp (the buildNCK + _buildNCK). Chainable.
func (b *Builder) BuildNCK() *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	if b.err != nil {
		return b
	}
	if b.message.TotalSegment("NCK") > 0 {
		b.fail(helpers.NewHL7FatalError("You can only have one NCK segment per HL7 Message."))
		return b
	}
	b.startSegment("NCK")

	var setMaxLength string
	switch b.version {
	case "2.1":
		setMaxLength = "19"
	case "2.2", "2.3", "2.3.1", "2.4", "2.5", "2.5.1":
		setMaxLength = "26"
	case "2.6", "2.7", "2.7.1", "2.8":
		setMaxLength = "24"
	default:
		setMaxLength = "19"
	}
	max, _ := strconv.Atoi(setMaxLength)
	b.validatorSetValue("1", b.SetDate(timeNow(), setMaxLength), &ValidationRule{
		Length:   lenMinMax(8, max),
		Required: true,
		Type:     ruleDate,
		HasType:  true,
	})
	return b
}

// BuildNST builds an NST (Statistics) segment (the _buildNST). Chainable.
func (b *Builder) BuildNST(properties Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.startSegment("NST")
	b.validatorSetValue("1", pick(properties, "nst_1"), &ValidationRule{Required: true})
	for i := 2; i <= 15; i++ {
		b.validatorSetValue(strconv.Itoa(i), pick(properties, "nst_"+strconv.Itoa(i)), nil)
	}
	return b
}

// BuildDSP builds a DSP (Display Data) segment (the _buildDSP). Chainable.
func (b *Builder) BuildDSP(properties Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.startSegment("DSP")

	pre27 := b.version != "2.7" && b.version != "2.7.1" && b.version != "2.8"

	b.validatorSetValue("1", pick(properties, "dsp_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.validatorSetValue("2", pick(properties, "dsp_2"), &ValidationRule{Length: lenMinMax(1, 4)})
	r3 := &ValidationRule{Required: true}
	if pre27 {
		r3.Length = lenMinMax(1, 300)
	}
	b.validatorSetValue("3", pick(properties, "dsp_3"), r3)
	r4 := &ValidationRule{}
	if pre27 {
		r4.Length = lenMinMax(1, 2)
	}
	b.validatorSetValue("4", pick(properties, "dsp_4"), r4)
	r5 := &ValidationRule{}
	if pre27 {
		r5.Length = lenMinMax(1, 20)
	}
	b.validatorSetValue("5", pick(properties, "dsp_5"), r5)
	return b
}

// BuildECD builds an ECD (Equipment Command) segment (the HL7_2_4._buildECD,
// inherited through 2.8). ECD did not exist before v2.4; the version assertion
// and per-version usage codes (ECD.4: O in 2.4-2.5.1, B in 2.6-2.7.1, W in 2.8
// and withdrawn already in 2.7) are enforced by the validator. Chainable.
func (b *Builder) BuildECD(properties Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	// the spec exposes buildECD only from HL7_2_4 onward; earlier versions fall
	// through to the Builder._buildECD stub that throws "Not Implemented".
	b.notImplementedBefore("2.4")
	s := b.spec("ECD")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("ECD")
	b.setField(s, 1, pick(properties, "ecd_1", "referenceCommandNumber"), nil)
	b.setField(s, 2, pick(properties, "ecd_2", "remoteControlCommand"), nil)
	b.setField(s, 3, pick(properties, "ecd_3", "responseRequired"), nil)
	b.setField(s, 4, pick(properties, "ecd_4", "requestedCompletionTime"), nil)
	b.setField(s, 5, pick(properties, "ecd_5", "parameters"), nil)
	return b
}
