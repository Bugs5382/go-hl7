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

// These BuildXXX methods port the HL7_2_4 typed segment builders (DRG,
// GOL, IAM, OM1-OM6, PRB, PTH, TXA), introduced in v2.4. ECD (also v2.4) lives
// in build_segments.go. Each is a validatorSetField sequence over the shared
// base; the version guard rejects the segment on earlier versions just as
// the HL7_BASE._buildXXX stub throws "Not Implemented". The OBR/ORC/PID
// version extensions the spec adds in HL7_2_4 live in build_segments_v21.go.

// BuildDRG builds a DRG (Diagnosis Related Group) segment (the
// HL7_2_4._buildDRG). Chainable.
func (b *HL7_BASE) BuildDRG(p Props) *HL7_BASE {
	b.headerExists()
	b.notImplementedBefore("2.4")
	s := spec("DRG")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "DRG")
	b.setField(s, 1, pick(p, "drg_1"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 2, b.dv(pick(p, "drg_2"), ""), dateRule())
	b.setField(s, 3, pick(p, "drg_3"), &ValidationRule{AllowedValues: []string{"Y", "N"}, Length: lenExact(1)})
	b.setField(s, 4, pick(p, "drg_4"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 5, pick(p, "drg_5"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 6, jsStringOr(pick(p, "drg_6")), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 7, pick(p, "drg_7"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 8, pick(p, "drg_8"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}

// BuildGOL builds a GOL (Goal Detail) segment (the HL7_2_4._buildGOL).
// Chainable.
func (b *HL7_BASE) BuildGOL(p Props) *HL7_BASE {
	b.headerExists()
	b.notImplementedBefore("2.4")
	s := spec("GOL")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "GOL")
	b.setField(s, 1, pick(p, "gol_1"), &ValidationRule{AllowedValues: []string{"AD", "CO", "DE", "LI", "UC", "UN"}})
	b.setField(s, 2, b.dv(pick(p, "gol_2"), ""), dateRule())
	b.setField(s, 3, pick(p, "gol_3"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 4, pick(p, "gol_4"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 5, pick(p, "gol_5"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 6, jsStringOr(pick(p, "gol_6")), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 7, b.dv(pick(p, "gol_7"), ""), dateRule())
	b.setField(s, 8, b.dv(pick(p, "gol_8"), ""), dateRule())
	b.setField(s, 9, b.dv(pick(p, "gol_9"), ""), dateRule())
	b.setField(s, 10, pick(p, "gol_10"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 11, pick(p, "gol_11"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 12, b.dv(pick(p, "gol_12"), ""), dateRule())
	b.setField(s, 13, b.dv(pick(p, "gol_13"), ""), dateRule())
	b.setField(s, 14, b.dv(pick(p, "gol_14"), ""), dateRule())
	b.setField(s, 15, b.dv(pick(p, "gol_15"), ""), dateRule())
	b.setField(s, 16, pick(p, "gol_16"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 17, pick(p, "gol_17"), &ValidationRule{Length: lenMinMax(1, 300)})
	b.setField(s, 18, pick(p, "gol_18"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 19, b.dv(pick(p, "gol_19"), ""), dateRule())
	b.setField(s, 20, pick(p, "gol_20"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 21, pick(p, "gol_21"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}

// BuildIAM builds an IAM (Patient Adverse Reaction Information) segment (the
// HL7_2_4._buildIAM). Chainable.
func (b *HL7_BASE) BuildIAM(p Props) *HL7_BASE {
	b.headerExists()
	b.notImplementedBefore("2.4")
	s := spec("IAM")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "IAM")
	b.setField(s, 1, jsStringOr(pick(p, "iam_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "iam_2"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 3, pick(p, "iam_3"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 4, pick(p, "iam_4"), &ValidationRule{AllowedValues: []string{"MI", "MO", "SV", "U"}})
	b.setField(s, 5, pick(p, "iam_5"), &ValidationRule{Length: lenMinMax(1, 15)})
	b.setField(s, 6, pick(p, "iam_6"), &ValidationRule{AllowedValues: []string{"A", "D", "U"}})
	b.setField(s, 7, pick(p, "iam_7"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 8, pick(p, "iam_8"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 9, pick(p, "iam_9"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 10, pick(p, "iam_10"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 11, b.dv(pick(p, "iam_11"), ""), dateRule())
	return b
}

// BuildOM1 builds an OM1 (General Segment for Observation Definitions) (the
// HL7_2_4._buildOM1). Chainable.
func (b *HL7_BASE) BuildOM1(p Props) *HL7_BASE {
	b.headerExists()
	b.notImplementedBefore("2.4")
	s := spec("OM1")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "OM1")
	b.setField(s, 1, jsStringOr(pick(p, "om1_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "om1_2"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 3, pick(p, "om1_3"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 4, pick(p, "om1_4"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 5, pick(p, "om1_5"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 6, pick(p, "om1_6"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 7, pick(p, "om1_7"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 8, pick(p, "om1_8"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 9, pick(p, "om1_9"), &ValidationRule{Length: lenMinMax(1, 30)})
	b.setField(s, 10, pick(p, "om1_10"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 11, pick(p, "om1_11"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 12, pick(p, "om1_12"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 13, pick(p, "om1_13"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 14, pick(p, "om1_14"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 15, pick(p, "om1_15"), &ValidationRule{AllowedValues: []string{"Y", "N"}})
	b.setField(s, 16, pick(p, "om1_16"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 17, pick(p, "om1_17"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 18, pick(p, "om1_18"), &ValidationRule{AllowedValues: []string{"A", "C", "E", "F", "P", "S"}})
	b.setField(s, 19, pick(p, "om1_19"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 20, pick(p, "om1_20"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 21, b.dv(pick(p, "om1_21"), ""), dateRule())
	b.setField(s, 22, b.dv(pick(p, "om1_22"), ""), dateRule())
	b.setField(s, 23, pick(p, "om1_23"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 24, jsStringOr(pick(p, "om1_24")), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 25, pick(p, "om1_25"), &ValidationRule{AllowedValues: []string{"A", "B", "C", "P", "R", "S", "T"}})
	b.setField(s, 26, pick(p, "om1_26"), &ValidationRule{AllowedValues: []string{"C", "R"}})
	b.setField(s, 27, pick(p, "om1_27"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 28, pick(p, "om1_28"), &ValidationRule{Length: lenMinMax(1, 1000)})
	b.setField(s, 29, pick(p, "om1_29"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 30, pick(p, "om1_30"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 31, pick(p, "om1_31"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 32, pick(p, "om1_32"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 33, pick(p, "om1_33"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 34, pick(p, "om1_34"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 35, pick(p, "om1_35"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 36, pick(p, "om1_36"), &ValidationRule{Length: lenMinMax(1, 65536)})
	b.setField(s, 37, pick(p, "om1_37"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 38, pick(p, "om1_38"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 39, pick(p, "om1_39"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 40, pick(p, "om1_40"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 41, pick(p, "om1_41"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 42, pick(p, "om1_42"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 43, pick(p, "om1_43"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 44, pick(p, "om1_44"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 45, pick(p, "om1_45"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 46, pick(p, "om1_46"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 47, pick(p, "om1_47"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}

// BuildOM2 builds an OM2 (Numeric Observation) segment (the
// HL7_2_4._buildOM2). Chainable.
func (b *HL7_BASE) BuildOM2(p Props) *HL7_BASE {
	b.headerExists()
	b.notImplementedBefore("2.4")
	s := spec("OM2")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "OM2")
	b.setField(s, 1, jsStringOr(pick(p, "om2_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "om2_2"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 3, pick(p, "om2_3"), &ValidationRule{Length: lenMinMax(1, 10)})
	b.setField(s, 4, pick(p, "om2_4"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 5, pick(p, "om2_5"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 6, pick(p, "om2_6"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 7, pick(p, "om2_7"), &ValidationRule{Length: lenMinMax(1, 205)})
	b.setField(s, 8, pick(p, "om2_8"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 9, pick(p, "om2_9"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 10, pick(p, "om2_10"), &ValidationRule{Length: lenMinMax(1, 20)})
	return b
}

// BuildOM3 builds an OM3 (Categorical Test/Observation) segment (the
// HL7_2_4._buildOM3). Chainable.
func (b *HL7_BASE) BuildOM3(p Props) *HL7_BASE {
	b.headerExists()
	b.notImplementedBefore("2.4")
	s := spec("OM3")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "OM3")
	b.setField(s, 1, jsStringOr(pick(p, "om3_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "om3_2"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 3, pick(p, "om3_3"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 4, pick(p, "om3_4"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 5, pick(p, "om3_5"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 6, pick(p, "om3_6"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 7, pick(p, "om3_7"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 8, pick(p, "om3_8"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}

// BuildOM4 builds an OM4 (Observations Requiring Specimens) segment (the
// HL7_2_4._buildOM4). Chainable.
func (b *HL7_BASE) BuildOM4(p Props) *HL7_BASE {
	b.headerExists()
	b.notImplementedBefore("2.4")
	s := spec("OM4")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "OM4")
	b.setField(s, 1, jsStringOr(pick(p, "om4_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "om4_2"), &ValidationRule{Length: lenMinMax(1, 1)})
	b.setField(s, 3, pick(p, "om4_3"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 4, jsStringOr(pick(p, "om4_4")), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 5, pick(p, "om4_5"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 6, pick(p, "om4_6"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 7, pick(p, "om4_7"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 8, pick(p, "om4_8"), &ValidationRule{Length: lenMinMax(1, 10240)})
	b.setField(s, 9, pick(p, "om4_9"), &ValidationRule{Length: lenMinMax(1, 10240)})
	b.setField(s, 10, pick(p, "om4_10"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 11, pick(p, "om4_11"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 12, pick(p, "om4_12"), &ValidationRule{Length: lenMinMax(1, 10240)})
	b.setField(s, 13, pick(p, "om4_13"), &ValidationRule{AllowedValues: []string{"E", "R", "S", "T"}})
	b.setField(s, 14, pick(p, "om4_14"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 15, pick(p, "om4_15"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 16, pick(p, "om4_16"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 17, pick(p, "om4_17"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 18, pick(p, "om4_18"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}

// BuildOM5 builds an OM5 (Observation Batteries) segment (the
// HL7_2_4._buildOM5). Chainable.
func (b *HL7_BASE) BuildOM5(p Props) *HL7_BASE {
	b.headerExists()
	b.notImplementedBefore("2.4")
	s := spec("OM5")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "OM5")
	b.setField(s, 1, jsStringOr(pick(p, "om5_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "om5_2"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 3, pick(p, "om5_3"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}

// BuildOM6 builds an OM6 (Observations Calculated from Other Observations)
// segment (the HL7_2_4._buildOM6). Chainable.
func (b *HL7_BASE) BuildOM6(p Props) *HL7_BASE {
	b.headerExists()
	b.notImplementedBefore("2.4")
	s := spec("OM6")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "OM6")
	b.setField(s, 1, jsStringOr(pick(p, "om6_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "om6_2"), &ValidationRule{Length: lenMinMax(1, 10240)})
	return b
}

// BuildPRB builds a PRB (Problem Detail) segment (the HL7_2_4._buildPRB).
// Chainable.
func (b *HL7_BASE) BuildPRB(p Props) *HL7_BASE {
	b.headerExists()
	b.notImplementedBefore("2.4")
	s := spec("PRB")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "PRB")
	b.setField(s, 1, pick(p, "prb_1"), &ValidationRule{AllowedValues: []string{"AD", "CO", "DE", "LI", "UC", "UN"}})
	b.setField(s, 2, b.dv(pick(p, "prb_2"), ""), dateRule())
	b.setField(s, 3, pick(p, "prb_3"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 4, pick(p, "prb_4"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 5, pick(p, "prb_5"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 6, jsStringOr(pick(p, "prb_6")), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 7, b.dv(pick(p, "prb_7"), ""), dateRule())
	b.setField(s, 8, b.dv(pick(p, "prb_8"), ""), dateRule())
	b.setField(s, 9, b.dv(pick(p, "prb_9"), ""), dateRule())
	b.setField(s, 10, pick(p, "prb_10"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 11, pick(p, "prb_11"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 12, pick(p, "prb_12"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 13, pick(p, "prb_13"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 14, pick(p, "prb_14"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 15, b.dv(pick(p, "prb_15"), ""), dateRule())
	b.setField(s, 16, b.dv(pick(p, "prb_16"), ""), dateRule())
	b.setField(s, 17, pick(p, "prb_17"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 18, pick(p, "prb_18"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 19, pick(p, "prb_19"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 20, jsStringOr(pick(p, "prb_20")), &ValidationRule{Length: lenMinMax(1, 5)})
	b.setField(s, 21, pick(p, "prb_21"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 22, pick(p, "prb_22"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 23, pick(p, "prb_23"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 24, pick(p, "prb_24"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 25, pick(p, "prb_25"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 26, pick(p, "prb_26"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}

// BuildPTH builds a PTH (Pathway) segment (the HL7_2_4._buildPTH).
// Chainable.
func (b *HL7_BASE) BuildPTH(p Props) *HL7_BASE {
	b.headerExists()
	b.notImplementedBefore("2.4")
	s := spec("PTH")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "PTH")
	b.setField(s, 1, pick(p, "pth_1"), &ValidationRule{AllowedValues: []string{"AD", "DE", "LI", "UN"}})
	b.setField(s, 2, pick(p, "pth_2"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 3, pick(p, "pth_3"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 4, b.dv(pick(p, "pth_4"), ""), dateRule())
	b.setField(s, 5, pick(p, "pth_5"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 6, b.dv(pick(p, "pth_6"), ""), dateRule())
	b.setField(s, 7, pick(p, "pth_7"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}

// BuildTXA builds a TXA (Transcription Document Header) segment (the
// HL7_2_4._buildTXA). Chainable.
func (b *HL7_BASE) BuildTXA(p Props) *HL7_BASE {
	b.headerExists()
	b.notImplementedBefore("2.4")
	s := spec("TXA")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "TXA")
	b.setField(s, 1, jsStringOr(pick(p, "txa_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "txa_2"), &ValidationRule{Length: lenMinMax(1, 30)})
	b.setField(s, 3, pick(p, "txa_3"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 4, b.dv(pick(p, "txa_4"), ""), dateRule())
	b.setField(s, 5, pick(p, "txa_5"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 6, b.dv(pick(p, "txa_6"), ""), dateRule())
	b.setField(s, 7, b.dv(pick(p, "txa_7"), ""), dateRule())
	b.setField(s, 8, b.dv(pick(p, "txa_8"), ""), dateRule())
	b.setField(s, 9, pick(p, "txa_9"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 10, pick(p, "txa_10"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 11, pick(p, "txa_11"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 12, pick(p, "txa_12"), &ValidationRule{Length: lenMinMax(1, 30)})
	b.setField(s, 13, pick(p, "txa_13"), &ValidationRule{Length: lenMinMax(1, 30)})
	b.setField(s, 14, pick(p, "txa_14"), &ValidationRule{Length: lenMinMax(1, 22)})
	b.setField(s, 15, pick(p, "txa_15"), &ValidationRule{Length: lenMinMax(1, 22)})
	b.setField(s, 16, pick(p, "txa_16"), &ValidationRule{Length: lenMinMax(1, 30)})
	b.setField(s, 17, pick(p, "txa_17"), &ValidationRule{AllowedValues: []string{"AU", "CA", "DO", "DT", "IN", "IP", "LA", "OB", "PA", "PR", "PY", "RD", "RV", "UN"}})
	b.setField(s, 18, pick(p, "txa_18"), &ValidationRule{AllowedValues: []string{"ET", "EMP", "UWL", "V", "R"}})
	b.setField(s, 19, pick(p, "txa_19"), &ValidationRule{AllowedValues: []string{"AV", "CA", "OB", "UN"}})
	b.setField(s, 20, pick(p, "txa_20"), &ValidationRule{AllowedValues: []string{"AC", "AA", "AH", "AL", "AR", "PU"}})
	b.setField(s, 21, pick(p, "txa_21"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 22, pick(p, "txa_22"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 23, pick(p, "txa_23"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}
