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

// These BuildXXX methods port the HL7_2_5 typed segment builders (SFT,
// SPM) and HL7_2_6 typed segment builders (BPX, BTX, ITM, IVT, REL). Each is a
// validatorSetField sequence over the shared base; the version guard rejects
// the segment on earlier versions just as the Builder._buildXXX stub throws
// "Not Implemented".

// BuildSFT builds an SFT (Software Segment) (the HL7_2_5._buildSFT).
// Introduced in v2.5. Chainable.
func (b *Builder) BuildSFT(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.5")
	s := b.spec("SFT")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("SFT")
	b.setField(s, 1, pick(p, "sft_1"), &ValidationRule{Length: lenMinMax(1, 567)})
	b.setField(s, 2, pick(p, "sft_2"), &ValidationRule{Length: lenMinMax(1, 15)})
	b.setField(s, 3, pick(p, "sft_3"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 4, pick(p, "sft_4"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 5, pick(p, "sft_5"), &ValidationRule{Length: lenMinMax(1, 1024)})
	b.setField(s, 6, b.dv(pick(p, "sft_6"), ""), dateRule())
	return b
}

// BuildSPM builds an SPM (Specimen) segment (the HL7_2_5._buildSPM).
// Introduced in v2.5. Chainable.
func (b *Builder) BuildSPM(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.5")
	s := b.spec("SPM")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("SPM")
	b.setField(s, 1, jsStringOr(pick(p, "spm_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "spm_2"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 3, pick(p, "spm_3"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 4, pick(p, "spm_4"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 5, pick(p, "spm_5"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 6, pick(p, "spm_6"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 7, pick(p, "spm_7"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 8, pick(p, "spm_8"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 9, pick(p, "spm_9"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 10, pick(p, "spm_10"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 11, pick(p, "spm_11"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 12, pick(p, "spm_12"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 13, jsStringOr(pick(p, "spm_13")), &ValidationRule{Length: lenMinMax(1, 6)})
	b.setField(s, 14, pick(p, "spm_14"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 15, pick(p, "spm_15"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 16, pick(p, "spm_16"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 17, b.dv(pick(p, "spm_17"), ""), dateRule())
	b.setField(s, 18, b.dv(pick(p, "spm_18"), ""), dateRule())
	b.setField(s, 19, b.dv(pick(p, "spm_19"), ""), dateRule())
	b.setField(s, 20, pick(p, "spm_20"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 21, pick(p, "spm_21"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 22, pick(p, "spm_22"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 23, pick(p, "spm_23"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 24, pick(p, "spm_24"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 25, pick(p, "spm_25"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 26, jsStringOr(pick(p, "spm_26")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 27, pick(p, "spm_27"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}

// BuildBPX builds a BPX (Blood Product Dispense Status) segment (the
// HL7_2_6._buildBPX). Introduced in v2.6. Chainable.
func (b *Builder) BuildBPX(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.6")
	s := b.spec("BPX")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("BPX")
	b.setField(s, 1, jsStringOr(pick(p, "bpx_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "bpx_2"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 3, pick(p, "bpx_3"), &ValidationRule{Length: lenMinMax(1, 1)})
	b.setField(s, 4, b.dv(pick(p, "bpx_4"), ""), dateRule())
	b.setField(s, 5, pick(p, "bpx_5"), &ValidationRule{Length: lenMinMax(1, 22)})
	b.setField(s, 6, pick(p, "bpx_6"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 7, pick(p, "bpx_7"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 8, pick(p, "bpx_8"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 9, pick(p, "bpx_9"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 10, pick(p, "bpx_10"), &ValidationRule{Length: lenMinMax(1, 22)})
	b.setField(s, 11, pick(p, "bpx_11"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 12, pick(p, "bpx_12"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 13, b.dv(pick(p, "bpx_13"), ""), dateRule())
	b.setField(s, 14, pick(p, "bpx_14"), &ValidationRule{Length: lenMinMax(1, 5)})
	b.setField(s, 15, pick(p, "bpx_15"), &ValidationRule{Length: lenMinMax(1, 5)})
	b.setField(s, 16, pick(p, "bpx_16"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 17, pick(p, "bpx_17"), &ValidationRule{Length: lenMinMax(1, 22)})
	return b
}

// BuildBTX builds a BTX (Blood Product Transfusion/Disposition) segment (the
// HL7_2_6._buildBTX). Introduced in v2.6. Chainable.
func (b *Builder) BuildBTX(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.6")
	s := b.spec("BTX")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("BTX")
	b.setField(s, 1, jsStringOr(pick(p, "btx_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "btx_2"), &ValidationRule{Length: lenMinMax(1, 22)})
	b.setField(s, 3, pick(p, "btx_3"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 4, pick(p, "btx_4"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 5, pick(p, "btx_5"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 6, pick(p, "btx_6"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 7, pick(p, "btx_7"), &ValidationRule{Length: lenMinMax(1, 22)})
	b.setField(s, 8, pick(p, "btx_8"), &ValidationRule{Length: lenMinMax(1, 5)})
	b.setField(s, 9, pick(p, "btx_9"), &ValidationRule{Length: lenMinMax(1, 5)})
	b.setField(s, 10, pick(p, "btx_10"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 11, pick(p, "btx_11"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 12, b.dv(pick(p, "btx_12"), ""), dateRule())
	b.setField(s, 13, b.dv(pick(p, "btx_13"), ""), dateRule())
	b.setField(s, 14, pick(p, "btx_14"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 15, pick(p, "btx_15"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 16, b.dv(pick(p, "btx_16"), ""), dateRule())
	b.setField(s, 17, b.dv(pick(p, "btx_17"), ""), dateRule())
	return b
}

// BuildITM builds an ITM (Material Item) segment (the HL7_2_6._buildITM).
// Introduced in v2.6. Chainable.
func (b *Builder) BuildITM(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.6")
	s := b.spec("ITM")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("ITM")
	b.setField(s, 1, pick(p, "itm_1"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 2, pick(p, "itm_2"), &ValidationRule{Length: lenMinMax(1, 999)})
	b.setField(s, 3, pick(p, "itm_3"), &ValidationRule{AllowedValues: []string{"A", "I", "P"}})
	b.setField(s, 4, pick(p, "itm_4"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 5, pick(p, "itm_5"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 6, pick(p, "itm_6"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 7, pick(p, "itm_7"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 8, pick(p, "itm_8"), &ValidationRule{Length: lenMinMax(1, 999)})
	b.setField(s, 9, pick(p, "itm_9"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 10, pick(p, "itm_10"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 11, pick(p, "itm_11"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 12, pick(p, "itm_12"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 13, pick(p, "itm_13"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 14, pick(p, "itm_14"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 15, pick(p, "itm_15"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 16, pick(p, "itm_16"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 17, pick(p, "itm_17"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 18, pick(p, "itm_18"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 19, pick(p, "itm_19"), &ValidationRule{Length: lenMinMax(1, 30)})
	b.setField(s, 20, jsStringOr(pick(p, "itm_20")), &ValidationRule{Length: lenMinMax(1, 6)})
	b.setField(s, 21, pick(p, "itm_21"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 22, pick(p, "itm_22"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 23, pick(p, "itm_23"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 24, pick(p, "itm_24"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 25, pick(p, "itm_25"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 26, pick(p, "itm_26"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 27, pick(p, "itm_27"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 28, pick(p, "itm_28"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 29, pick(p, "itm_29"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 30, pick(p, "itm_30"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 31, pick(p, "itm_31"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 32, pick(p, "itm_32"), &ValidationRule{Length: lenMinMax(1, 16)})
	b.setField(s, 33, pick(p, "itm_33"), &ValidationRule{Length: lenMinMax(1, 30)})
	b.setField(s, 34, pick(p, "itm_34"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 35, pick(p, "itm_35"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 36, pick(p, "itm_36"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 37, pick(p, "itm_37"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 38, pick(p, "itm_38"), &ValidationRule{Length: lenMinMax(1, 6)})
	return b
}

// BuildIVT builds an IVT (Material Location) segment (the
// HL7_2_6._buildIVT). Introduced in v2.6. Chainable.
func (b *Builder) BuildIVT(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.6")
	s := b.spec("IVT")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("IVT")
	b.setField(s, 1, jsStringOr(pick(p, "ivt_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "ivt_2"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 3, pick(p, "ivt_3"), &ValidationRule{Length: lenMinMax(1, 999)})
	b.setField(s, 4, pick(p, "ivt_4"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 5, pick(p, "ivt_5"), &ValidationRule{Length: lenMinMax(1, 999)})
	b.setField(s, 6, pick(p, "ivt_6"), &ValidationRule{AllowedValues: []string{"A", "I", "P"}})
	b.setField(s, 7, pick(p, "ivt_7"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 8, pick(p, "ivt_8"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 9, pick(p, "ivt_9"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 10, pick(p, "ivt_10"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 11, pick(p, "ivt_11"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 12, pick(p, "ivt_12"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 13, pick(p, "ivt_13"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 14, pick(p, "ivt_14"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 15, pick(p, "ivt_15"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 16, pick(p, "ivt_16"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 17, pick(p, "ivt_17"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 18, pick(p, "ivt_18"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 19, pick(p, "ivt_19"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 20, pick(p, "ivt_20"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 21, pick(p, "ivt_21"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 22, pick(p, "ivt_22"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 23, pick(p, "ivt_23"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 24, pick(p, "ivt_24"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 25, pick(p, "ivt_25"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	return b
}

// BuildREL builds a REL (Clinical Relationship Segment) (the
// HL7_2_6._buildREL). Introduced in v2.6. Chainable.
func (b *Builder) BuildREL(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.6")
	s := b.spec("REL")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("REL")
	b.setField(s, 1, jsStringOr(pick(p, "rel_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "rel_2"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 3, pick(p, "rel_3"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 4, pick(p, "rel_4"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 5, pick(p, "rel_5"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 6, b.dv(pick(p, "rel_6"), ""), dateRule())
	b.setField(s, 7, b.dv(pick(p, "rel_7"), ""), dateRule())
	b.setField(s, 8, pick(p, "rel_8"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 9, pick(p, "rel_9"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 10, pick(p, "rel_10"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 11, pick(p, "rel_11"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 12, pick(p, "rel_12"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 13, b.dv(pick(p, "rel_13"), ""), dateRule())
	b.setField(s, 14, pick(p, "rel_14"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}
