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

package hl7

// These BuildXXX methods port the HL7_2_1 protected _buildXXX segment
// builders (the base segment set every later version inherits). Field-level
// per-version differences are gated by the usage catalog the validator
// consults, so these shared bodies stay version-correct: a field absent from
// the current version with no value is a no-op, and one with a value is
// rejected. The per-field length/date/allowedValues overrides are ported
// verbatim from the 2.1 source. The MSH builders live in build_msh.go.

// dv formats a Date-typed override field at the standard message date length,
// falling back to fallback. Shorthand for the `x instanceof Date ? setDate(x)
// : fallback` idiom the spec repeats in every date field.
func (b *HL7_BASE) dv(v any, fallback any) any { return b.dateField(v, b.opt.Date, fallback) }

// dv8 is dv at the fixed "8" (YYYYMMDD) length.
func (b *HL7_BASE) dv8(v any, fallback any) any { return b.dateField(v, "8", fallback) }

// BuildACC builds an ACC (Accident) segment (the _buildACC). Chainable.
func (b *HL7_BASE) BuildACC(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("ACC")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "ACC")
	b.setField(s, 1, b.dv(pick(p, "acc_1", "timeStamp"), ""), &ValidationRule{Length: lenMinMax(8, 19), Type: ruleDate, HasType: true})
	b.setField(s, 2, pick(p, "acc_2", "accidentCode"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 3, pick(p, "acc_3", "accidentLocation"), &ValidationRule{Length: lenMinMax(1, 25)})
	return b
}

// BuildBLG builds a BLG (Billing) segment (the _buildBLG). Chainable.
func (b *HL7_BASE) BuildBLG(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("BLG")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "BLG")
	b.setField(s, 1, pick(p, "blg_1", "billingWhenToCharge"), &ValidationRule{AllowedValues: b.codeTable("0100"), Length: lenMinMax(1, 15)})
	b.setField(s, 2, pick(p, "blg_2", "billingChargeType"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 3, pick(p, "blg_3", "billingAccountId"), &ValidationRule{Length: lenMinMax(1, 25)})
	return b
}

// BuildDG1 builds a DG1 (Diagnosis) segment (the _buildDG1). Chainable.
func (b *HL7_BASE) BuildDG1(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("DG1")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "DG1")
	b.setField(s, 1, pick(p, "dg1_1", "diagnosisId"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "dg1_2", "diagnosisCodingMethod"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 3, pick(p, "dg1_3", "diagnosisCode"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 4, pick(p, "dg1_4", "diagnosisDescription"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 5, b.dv(pick(p, "dg1_5", "timeStamp"), ""), &ValidationRule{Length: lenMinMax(1, 19)})
	b.setField(s, 6, pick(p, "dg1_6", "diagnosisType"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 7, pick(p, "dg1_7", "diagnosisMajorCategory"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 8, pick(p, "dg1_8", "diagnosisRelatedGroup"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 9, pick(p, "dg1_9", "diagnosisApprovalIndicator"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 10, pick(p, "dg1_10", "diagnosisGrouperReviewCode"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 11, pick(p, "dg1_11", "diagnosisOutlierType"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 12, pick(p, "dg1_12", "diagnosisOutlierDays"), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 13, pick(p, "dg1_13", "diagnosisOutlierCost"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 14, pick(p, "dg1_14", "diagnosisGrouperVersionAndType"), &ValidationRule{Length: lenMinMax(1, 4)})
	return b
}

// BuildDSC builds a DSC (Continuation Pointer) segment (the _buildDSC).
func (b *HL7_BASE) BuildDSC(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("DSC")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "DSC")
	b.setField(s, 1, pick(p, "dsc_1", "continuationPointer"), &ValidationRule{Length: lenMinMax(1, 60)})
	return b
}

// BuildERR builds an ERR (Error) segment (the _buildERR). Chainable.
func (b *HL7_BASE) BuildERR(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("ERR")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "ERR")
	b.setField(s, 1, pick(p, "err_1", "errorIdAndLocation"), &ValidationRule{Length: lenMinMax(1, 80)})
	return b
}

// BuildEVN builds an EVN (Event Type) segment (the _buildEVN). Chainable.
func (b *HL7_BASE) BuildEVN(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("EVN")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "EVN")
	b.setField(s, 1, pick(p, "evn_1"), &ValidationRule{AllowedValues: b.codeTable("0003")})
	b.setField(s, 2, b.dv(pick(p, "evn_2"), b.SetDate(timeNow(), b.opt.Date)), dateRule())
	b.setField(s, 3, b.dv(pick(p, "evn_3"), ""), dateRule())
	b.setField(s, 4, pick(p, "evn_4"), &ValidationRule{AllowedValues: b.codeTable("0062")})
	return b
}

// BuildFT1 builds an FT1 (Financial Transaction) segment (the _buildFT1).
func (b *HL7_BASE) BuildFT1(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("FT1")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "FT1")
	b.setField(s, 1, pick(p, "ft1_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "ft1_2"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 3, pick(p, "ft1_3"), &ValidationRule{Length: lenMinMax(1, 5)})
	b.setField(s, 4, b.dv8(pick(p, "ft1_4"), b.SetDate(timeNow(), "8")), dateRule())
	b.setField(s, 5, b.dv8(pick(p, "ft1_5"), ""), dateRule())
	b.setField(s, 6, pick(p, "ft1_6"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 7, pick(p, "ft1_7"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 8, pick(p, "ft1_8"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 9, pick(p, "ft1_9"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 10, pick(p, "ft1_10"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 11, pick(p, "ft1_11"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 12, pick(p, "ft1_12"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 13, pick(p, "ft1_13"), &ValidationRule{Length: lenMinMax(1, 16)})
	b.setField(s, 14, pick(p, "ft1_14"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 15, pick(p, "ft1_15"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 16, pick(p, "ft1_16"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 17, pick(p, "ft1_17"), &ValidationRule{Length: lenExact(1)})
	b.setField(s, 18, pick(p, "ft1_18"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 19, pick(p, "ft1_19"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 20, pick(p, "ft1_20"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 21, pick(p, "ft1_21"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 22, pick(p, "ft1_22"), &ValidationRule{Length: lenMinMax(1, 12)})
	return b
}

// BuildGT1 builds a GT1 (Guarantor) segment (the _buildGT1). Chainable.
func (b *HL7_BASE) BuildGT1(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("GT1")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "GT1")
	b.setField(s, 1, pick(p, "gt1_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "gt1_2"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 3, pick(p, "gt1_3"), &ValidationRule{Length: lenMinMax(1, 5)})
	b.setField(s, 4, pick(p, "gt1_4"), nil)
	b.setField(s, 5, pick(p, "gt1_5"), nil)
	b.setField(s, 6, pick(p, "gt1_6"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 7, pick(p, "gt1_7"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 8, pick(p, "gt1_8"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 9, pick(p, "gt1_9"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 10, pick(p, "gt1_10"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 11, pick(p, "gt1_11"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 12, pick(p, "gt1_12"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 13, pick(p, "gt1_13"), &ValidationRule{Length: lenMinMax(1, 16)})
	b.setField(s, 14, pick(p, "gt1_14"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 15, pick(p, "gt1_15"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 16, pick(p, "gt1_16"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 17, pick(p, "gt1_17"), &ValidationRule{Length: lenExact(1)})
	b.setField(s, 18, pick(p, "gt1_18"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 19, pick(p, "gt1_19"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 20, pick(p, "gt1_20"), &ValidationRule{Length: lenMinMax(1, 60)})
	return b
}

// BuildIN1 builds an IN1 (Insurance) segment (the _buildIN1). Chainable.
func (b *HL7_BASE) BuildIN1(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("IN1")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "IN1")
	for i := 1; i <= 44; i++ {
		key := "in1_" + itoa(i)
		switch i {
		case 43:
			b.setField(s, 43, pick(p, key), &ValidationRule{AllowedValues: b.codeTable("0001"), Length: lenExact(1)})
		default:
			b.setField(s, i, pick(p, key), nil)
		}
	}
	return b
}

// BuildMRG builds an MRG (Merge Patient Info) segment (the _buildMRG).
func (b *HL7_BASE) BuildMRG(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("MRG")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "MRG")
	b.setField(s, 1, pick(p, "mrg_1"), &ValidationRule{Length: lenMinMax(1, 16)})
	b.setField(s, 2, pick(p, "mrg_2"), &ValidationRule{Length: lenMinMax(1, 16)})
	b.setField(s, 3, pick(p, "mrg_3"), &ValidationRule{Length: lenMinMax(1, 20)})
	return b
}

// BuildMSA builds an MSA (Acknowledgment) segment (the _buildMSA). Chainable.
func (b *HL7_BASE) BuildMSA(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("MSA")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "MSA")
	b.setField(s, 1, pick(p, "msa_1"), &ValidationRule{AllowedValues: b.codeTable("0008")})
	for i := 2; i <= 5; i++ {
		b.setField(s, i, pick(p, "msa_"+itoa(i)), nil)
	}
	return b
}

// BuildNK1 builds an NK1 (Next of Kin) segment (the _buildNK1). Chainable.
func (b *HL7_BASE) BuildNK1(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("NK1")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "NK1")
	for i := 1; i <= 5; i++ {
		b.setField(s, i, pick(p, "nk1_"+itoa(i)), nil)
	}

	// Version extensions (the HL7_2_3._buildNK1, gated by the usage catalog).
	if compareVersions(b.version, "2.3") >= 0 {
		b.setField(s, 6, pick(p, "nk1_6"), &ValidationRule{Length: lenMinMax(1, 40)})
		b.setField(s, 7, pick(p, "nk1_7"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 8, b.dv(pick(p, "nk1_8"), ""), dateRule())
		b.setField(s, 9, b.dv(pick(p, "nk1_9"), ""), dateRule())
		b.setField(s, 10, pick(p, "nk1_10"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 11, pick(p, "nk1_11"), &ValidationRule{Length: lenMinMax(1, 20)})
		b.setField(s, 12, pick(p, "nk1_12"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 13, pick(p, "nk1_13"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 14, pick(p, "nk1_14"), &ValidationRule{Length: lenMinMax(1, 3)})
		b.setField(s, 15, pick(p, "nk1_15"), &ValidationRule{Length: lenExact(1)})
		b.setField(s, 16, b.dv(pick(p, "nk1_16"), ""), dateRule())
		b.setField(s, 17, pick(p, "nk1_17"), &ValidationRule{Length: lenMinMax(1, 2)})
		b.setField(s, 18, pick(p, "nk1_18"), &ValidationRule{Length: lenMinMax(1, 2)})
		b.setField(s, 19, pick(p, "nk1_19"), &ValidationRule{Length: lenMinMax(1, 4)})
		b.setField(s, 20, pick(p, "nk1_20"), &ValidationRule{Length: lenMinMax(1, 60)})
	}
	return b
}

// BuildNPU builds an NPU (Bed Status Update) segment (the _buildNPU).
func (b *HL7_BASE) BuildNPU(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("NPU")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "NPU")
	b.setField(s, 1, pick(p, "npu_1"), nil)
	b.setField(s, 2, pick(p, "npu_2"), &ValidationRule{AllowedValues: b.codeTable("0116")})
	return b
}

// BuildNSC builds an NSC (Network Change) segment (the _buildNSC). Chainable.
func (b *HL7_BASE) BuildNSC(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("NSC")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "NSC")
	for i := 1; i <= 9; i++ {
		b.setField(s, i, pick(p, "nsc_"+itoa(i)), nil)
	}
	return b
}

// BuildNTE builds an NTE (Notes and Comments) segment (the _buildNTE).
func (b *HL7_BASE) BuildNTE(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("NTE")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "NTE")
	b.setField(s, 1, pick(p, "nte_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "nte_2", "sourceOfComment"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 3, pick(p, "nte_3", "comment"), &ValidationRule{Length: lenMinMax(1, 65536)})
	return b
}

// BuildOBR builds an OBR (Observation Request) segment (the _buildOBR).
func (b *HL7_BASE) BuildOBR(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("OBR")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "OBR")
	b.setField(s, 1, pick(p, "obr_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "obr_2", "placerOrderNumber"), &ValidationRule{Length: lenMinMax(1, 75)})
	b.setField(s, 3, pick(p, "obr_3", "fillerOrderNumber"), &ValidationRule{Length: lenMinMax(1, 75)})
	b.setField(s, 4, pick(p, "obr_4", "universalServiceId"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 5, pick(p, "obr_5"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 6, b.dv(pick(p, "obr_6"), ""), dateRule())
	b.setField(s, 7, b.dv(pick(p, "obr_7", "observationDateTime"), ""), dateRule())
	b.setField(s, 8, b.dv(pick(p, "obr_8", "observationEndDateTime"), ""), dateRule())
	b.setField(s, 9, pick(p, "obr_9", "collectionVolume"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 10, pick(p, "obr_10"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 11, pick(p, "obr_11"), &ValidationRule{Length: lenExact(1)})
	b.setField(s, 12, pick(p, "obr_12"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 13, pick(p, "obr_13"), &ValidationRule{Length: lenMinMax(1, 300)})
	b.setField(s, 14, b.dv(pick(p, "obr_14"), ""), dateRule())
	b.setField(s, 15, pick(p, "obr_15", "specimenSource"), &ValidationRule{Length: lenMinMax(1, 300)})
	b.setField(s, 16, pick(p, "obr_16", "orderingProvider"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 17, pick(p, "obr_17"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 18, pick(p, "obr_18"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 19, pick(p, "obr_19"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 20, pick(p, "obr_20"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 21, pick(p, "obr_21"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 22, b.dv(pick(p, "obr_22"), ""), dateRule())
	b.setField(s, 23, pick(p, "obr_23"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 24, pick(p, "obr_24", "diagnosticServiceSectionId"), &ValidationRule{AllowedValues: b.codeTable("0074"), Length: lenMinMax(1, 10)})
	b.setField(s, 25, pick(p, "obr_25", "resultStatus"), &ValidationRule{AllowedValues: b.codeTable("0123"), Length: lenExact(1)})

	// Version extensions: the spec overrides _buildOBR in later versions, calling
	// super first then appending the new fields. The usage catalog gates each
	// field, so these run only from the version that introduced them onward.
	if compareVersions(b.version, "2.2") >= 0 {
		b.setField(s, 26, pick(p, "obr_26"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 27, pick(p, "obr_27"), &ValidationRule{Length: lenMinMax(1, 200)})
		b.setField(s, 28, pick(p, "obr_28"), &ValidationRule{Length: lenMinMax(1, 150)})
		b.setField(s, 29, pick(p, "obr_29"), &ValidationRule{Length: lenMinMax(1, 150)})
		b.setField(s, 30, pick(p, "obr_30"), &ValidationRule{Length: lenMinMax(1, 20)})
		b.setField(s, 31, pick(p, "obr_31"), &ValidationRule{Length: lenMinMax(1, 300)})
		b.setField(s, 32, pick(p, "obr_32"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 33, pick(p, "obr_33"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 34, pick(p, "obr_34"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 35, pick(p, "obr_35"), &ValidationRule{Length: lenMinMax(1, 60)})
	}
	if compareVersions(b.version, "2.3") >= 0 {
		b.setField(s, 36, b.dv(pick(p, "obr_36"), ""), dateRule())
		b.setField(s, 37, pick(p, "obr_37"), &ValidationRule{Length: lenMinMax(1, 4)})
		b.setField(s, 38, pick(p, "obr_38"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 39, pick(p, "obr_39"), &ValidationRule{Length: lenMinMax(1, 200)})
		b.setField(s, 40, pick(p, "obr_40"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 41, pick(p, "obr_41"), &ValidationRule{AllowedValues: []string{"A", "W", "N"}, Length: lenExact(1)})
		b.setField(s, 42, pick(p, "obr_42"), &ValidationRule{AllowedValues: []string{"R", "O", "N"}, Length: lenExact(1)})
		b.setField(s, 43, pick(p, "obr_43"), &ValidationRule{Length: lenMinMax(1, 200)})
	}
	if compareVersions(b.version, "2.4") >= 0 {
		b.setField(s, 44, pick(p, "obr_44"), &ValidationRule{Length: lenMinMax(1, 250)})
		b.setField(s, 45, pick(p, "obr_45"), &ValidationRule{Length: lenMinMax(1, 250)})
		b.setField(s, 46, pick(p, "obr_46"), &ValidationRule{Length: lenMinMax(1, 250)})
		b.setField(s, 47, pick(p, "obr_47"), &ValidationRule{Length: lenMinMax(1, 250)})
	}
	return b
}

// BuildOBX builds an OBX (Observation/Result) segment (the _buildOBX).
func (b *HL7_BASE) BuildOBX(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("OBX")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "OBX")
	b.setField(s, 1, pick(p, "obx_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "obx_2", "valueType"), &ValidationRule{AllowedValues: b.codeTable("0125"), Length: lenMinMax(1, 3)})
	b.setField(s, 3, pick(p, "obx_3", "observationIdentifier"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 4, pick(p, "obx_4", "observationSubId"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 5, pick(p, "obx_5", "observationValue"), &ValidationRule{Length: lenMinMax(1, 65536)})
	b.setField(s, 6, pick(p, "obx_6", "units"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 7, pick(p, "obx_7", "referencesRange"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 8, pick(p, "obx_8", "abnormalFlags"), &ValidationRule{AllowedValues: b.codeTable("0078"), Length: lenMinMax(1, 5)})
	b.setField(s, 9, pick(p, "obx_9"), &ValidationRule{Length: lenMinMax(1, 5)})
	b.setField(s, 10, pick(p, "obx_10"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 11, pick(p, "obx_11", "observationResultStatus"), &ValidationRule{AllowedValues: b.codeTable("0085"), Length: lenExact(1)})

	// Version extensions (the HL7_2_x._buildOBX, gated by the usage catalog).
	if compareVersions(b.version, "2.2") >= 0 {
		b.setField(s, 12, b.dv(pick(p, "obx_12"), ""), dateRule())
		b.setField(s, 13, pick(p, "obx_13"), &ValidationRule{Length: lenMinMax(1, 20)})
		b.setField(s, 14, b.dv(pick(p, "obx_14"), ""), dateRule())
		b.setField(s, 15, pick(p, "obx_15"), &ValidationRule{Length: lenMinMax(1, 60)})
	}
	if compareVersions(b.version, "2.3") >= 0 {
		b.setField(s, 16, pick(p, "obx_16"), &ValidationRule{Length: lenMinMax(1, 80)})
		b.setField(s, 17, pick(p, "obx_17"), &ValidationRule{Length: lenMinMax(1, 60)})
	}
	return b
}

// BuildORC builds an ORC (Common Order) segment (the _buildORC). Chainable.
func (b *HL7_BASE) BuildORC(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("ORC")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "ORC")
	b.setField(s, 1, pick(p, "orc_1", "orderControl"), &ValidationRule{AllowedValues: b.codeTable("0119"), Length: lenMinMax(1, 2)})
	b.setField(s, 2, pick(p, "orc_2", "placerOrderNumber"), &ValidationRule{Length: lenMinMax(1, 75)})
	b.setField(s, 3, pick(p, "orc_3", "fillerOrderNumber"), &ValidationRule{Length: lenMinMax(1, 75)})
	b.setField(s, 4, pick(p, "orc_4"), &ValidationRule{Length: lenMinMax(1, 75)})
	b.setField(s, 5, pick(p, "orc_5", "orderStatus"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 6, pick(p, "orc_6", "responseFlag"), &ValidationRule{AllowedValues: b.codeTable("0121"), Length: lenExact(1)})
	b.setField(s, 7, pick(p, "orc_7"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 8, pick(p, "orc_8"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 9, b.dv(pick(p, "orc_9", "transactionDateTime"), ""), dateRule())
	b.setField(s, 10, pick(p, "orc_10", "enteredBy"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 11, pick(p, "orc_11", "verifiedBy"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 12, pick(p, "orc_12", "orderingProvider"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 13, pick(p, "orc_13"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 14, pick(p, "orc_14", "callBackPhoneNumber"), &ValidationRule{Length: lenMinMax(1, 40)})

	// Version extensions (the HL7_2_x._buildORC, gated by the usage catalog).
	if compareVersions(b.version, "2.2") >= 0 {
		b.setField(s, 15, b.dv(pick(p, "orc_15"), ""), dateRule())
		b.setField(s, 16, pick(p, "orc_16"), &ValidationRule{Length: lenMinMax(1, 200)})
		b.setField(s, 17, pick(p, "orc_17"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 18, pick(p, "orc_18"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 19, pick(p, "orc_19"), &ValidationRule{Length: lenMinMax(1, 80)})
		b.setField(s, 20, pick(p, "orc_20"), &ValidationRule{Length: lenMinMax(1, 40)})
	}
	if compareVersions(b.version, "2.3") >= 0 {
		b.setField(s, 21, pick(p, "orc_21"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 22, pick(p, "orc_22"), &ValidationRule{Length: lenMinMax(1, 106)})
		b.setField(s, 23, pick(p, "orc_23"), &ValidationRule{Length: lenMinMax(1, 40)})
	}
	if compareVersions(b.version, "2.4") >= 0 {
		b.setField(s, 24, pick(p, "orc_24"), &ValidationRule{Length: lenMinMax(1, 250)})
		b.setField(s, 25, pick(p, "orc_25"), &ValidationRule{Length: lenMinMax(1, 40)})
		b.setField(s, 26, b.dv(pick(p, "orc_26"), ""), dateRule())
		b.setField(s, 27, pick(p, "orc_27"), &ValidationRule{Length: lenMinMax(1, 250)})
		b.setField(s, 28, pick(p, "orc_28"), &ValidationRule{AllowedValues: []string{"I", "O"}, Length: lenExact(1)})
		b.setField(s, 29, pick(p, "orc_29"), &ValidationRule{Length: lenMinMax(1, 250)})
		b.setField(s, 30, pick(p, "orc_30"), &ValidationRule{Length: lenMinMax(1, 250)})
	}
	return b
}

// BuildPID builds a PID (Patient Identification) segment (the _buildPID).
// PID.11 (XAD) and PID.5 (XPN) accept composite-object inputs via the validator.
func (b *HL7_BASE) BuildPID(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("PID")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "PID")
	b.setField(s, 1, pick(p, "pid_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "pid_2", "patientIdExternal"), &ValidationRule{Length: lenMinMax(1, 16)})
	b.setField(s, 3, pick(p, "pid_3", "patientIdInternal"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 4, pick(p, "pid_4", "alternatePatientId"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 5, pick(p, "pid_5", "patientName"), &ValidationRule{Length: lenMinMax(1, 48)})
	b.setField(s, 6, pick(p, "pid_6", "mothersMaidenName"), &ValidationRule{Length: lenMinMax(1, 30)})
	b.setField(s, 7, b.dv(pick(p, "pid_7", "dateOfBirth"), ""), dateRule())
	b.setField(s, 8, pick(p, "pid_8", "sex"), &ValidationRule{AllowedValues: b.codeTable("0001"), Length: lenExact(1)})
	b.setField(s, 9, pick(p, "pid_9", "patientAlias"), &ValidationRule{Length: lenMinMax(1, 48)})
	b.setField(s, 10, pick(p, "pid_10", "race"), &ValidationRule{Length: lenMinMax(1, 1)})
	b.setField(s, 11, pick(p, "pid_11", "patientAddress"), &ValidationRule{Length: lenMinMax(1, 106)})
	b.setField(s, 12, pick(p, "pid_12", "countyCode"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 13, pick(p, "pid_13", "phoneNumberHome"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 14, pick(p, "pid_14", "phoneNumberBusiness"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 15, pick(p, "pid_15", "language"), &ValidationRule{Length: lenMinMax(1, 25)})
	b.setField(s, 16, pick(p, "pid_16", "maritalStatus"), &ValidationRule{AllowedValues: b.codeTable("0002"), Length: lenExact(1)})
	b.setField(s, 17, pick(p, "pid_17", "religion"), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 18, pick(p, "pid_18", "patientAccountNumber"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 19, pick(p, "pid_19", "ssn"), &ValidationRule{Length: lenMinMax(1, 16)})

	// Version extensions: the spec adds PID fields in later versions by overriding
	// _buildPID and calling super first. The usage catalog gates each field, so
	// these run only from the version that introduced them onward.
	if compareVersions(b.version, "2.2") >= 0 {
		b.setField(s, 20, pick(p, "pid_20"), &ValidationRule{Length: lenMinMax(1, 25)})
		b.setField(s, 21, pick(p, "pid_21"), &ValidationRule{Length: lenMinMax(1, 20)})
		b.setField(s, 22, pick(p, "pid_22"), &ValidationRule{Length: lenMinMax(1, 3)})
		b.setField(s, 23, pick(p, "pid_23"), &ValidationRule{Length: lenMinMax(1, 25)})
		b.setField(s, 24, pick(p, "pid_24"), &ValidationRule{AllowedValues: b.codeTable("0136"), Length: lenExact(1)})
		b.setField(s, 25, pickStrOrNil(p, "pid_25"), &ValidationRule{Length: lenExact(2)})
		b.setField(s, 26, pick(p, "pid_26"), &ValidationRule{Length: lenMinMax(1, 4)})
	}
	if compareVersions(b.version, "2.3") >= 0 {
		b.setField(s, 27, pick(p, "pid_27"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 28, pick(p, "pid_28"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.setField(s, 29, b.dv(pick(p, "pid_29"), ""), dateRule())
		b.setField(s, 30, pick(p, "pid_30"), &ValidationRule{AllowedValues: []string{"Y", "N"}, Length: lenExact(1)})
	}
	if compareVersions(b.version, "2.4") >= 0 {
		b.setField(s, 31, pick(p, "pid_31"), &ValidationRule{AllowedValues: []string{"Y", "N"}, Length: lenExact(1)})
		b.setField(s, 32, pick(p, "pid_32"), &ValidationRule{Length: lenMinMax(1, 20)})
		b.setField(s, 33, b.dv(pick(p, "pid_33"), ""), dateRule())
		b.setField(s, 34, pick(p, "pid_34"), &ValidationRule{Length: lenMinMax(1, 21)})
		b.setField(s, 35, pick(p, "pid_35"), &ValidationRule{Length: lenMinMax(1, 250)})
		b.setField(s, 36, pick(p, "pid_36"), &ValidationRule{Length: lenMinMax(1, 250)})
		b.setField(s, 37, pick(p, "pid_37"), &ValidationRule{Length: lenMinMax(1, 80)})
		b.setField(s, 38, pick(p, "pid_38"), &ValidationRule{Length: lenMinMax(1, 250)})
		b.setField(s, 39, pick(p, "pid_39"), &ValidationRule{Length: lenMinMax(1, 250)})
	}
	return b
}

// pickStrOrNil returns the string form of a prop value, or nil when absent. It
// mirrors the `pid_25 === undefined ? undefined : String(pid_25)`.
func pickStrOrNil(p Props, keys ...string) any {
	v := pick(p, keys...)
	if v == nil {
		return nil
	}
	return toStr(v)
}

// BuildPR1 builds a PR1 (Procedures) segment (the _buildPR1). Chainable.
func (b *HL7_BASE) BuildPR1(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("PR1")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "PR1")
	b.setField(s, 1, pick(p, "pr1_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "pr1_2", "procedureCodingMethod"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 3, pick(p, "pr1_3", "procedureCode"), &ValidationRule{Length: lenMinMax(1, 10)})
	b.setField(s, 4, pick(p, "pr1_4", "procedureDescription"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 5, b.dv(pick(p, "pr1_5", "procedureDateTime"), ""), dateRule())
	b.setField(s, 6, pick(p, "pr1_6", "procedureType"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 7, pick(p, "pr1_7", "procedureMinutes"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 8, pick(p, "pr1_8", "anesthesiologist"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 9, pick(p, "pr1_9", "anesthesiaCode"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 10, pick(p, "pr1_10", "anesthesiaMinutes"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 11, pick(p, "pr1_11", "surgeon"), &ValidationRule{Length: lenMinMax(1, 60)})
	return b
}

// BuildPV1 builds a PV1 (Patient Visit) segment (the _buildPV1). Chainable.
func (b *HL7_BASE) BuildPV1(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("PV1")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "PV1")
	b.setField(s, 1, pick(p, "pv1_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "pv1_2", "patientClass"), &ValidationRule{AllowedValues: b.codeTable("0004"), Length: lenExact(1)})
	b.setField(s, 3, pick(p, "pv1_3", "assignedPatientLocation"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 4, pick(p, "pv1_4", "admissionType"), &ValidationRule{AllowedValues: b.codeTable("0007"), Length: lenExact(1)})
	b.setField(s, 5, pick(p, "pv1_5", "preadmitNumber"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 6, pick(p, "pv1_6", "priorPatientLocation"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 7, pick(p, "pv1_7", "attendingDoctor"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 8, pick(p, "pv1_8", "referringDoctor"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 9, pick(p, "pv1_9", "consultingDoctor"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 10, pick(p, "pv1_10", "hospitalService"), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 11, pick(p, "pv1_11", "temporaryLocation"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 12, pick(p, "pv1_12"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 13, pick(p, "pv1_13"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 14, pick(p, "pv1_14", "admitSource"), &ValidationRule{Length: lenExact(1)})
	b.setField(s, 15, pick(p, "pv1_15", "ambulatoryStatus"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 16, pick(p, "pv1_16", "vipIndicator"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 17, pick(p, "pv1_17", "admittingDoctor"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 18, pick(p, "pv1_18", "patientType"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 19, pick(p, "pv1_19", "visitNumber"), &ValidationRule{Length: lenMinMax(1, 15)})
	b.setField(s, 20, pick(p, "pv1_20", "financialClass"), &ValidationRule{Length: lenMinMax(1, 50)})
	b.setField(s, 21, pick(p, "pv1_21"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 22, pick(p, "pv1_22"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 23, pick(p, "pv1_23"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 24, pick(p, "pv1_24"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 25, b.dv8(pick(p, "pv1_25"), ""), dateRule())
	b.setField(s, 26, pick(p, "pv1_26"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 27, pick(p, "pv1_27"), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 28, pick(p, "pv1_28"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 29, pick(p, "pv1_29"), &ValidationRule{Length: lenExact(4)})
	b.setField(s, 30, b.dv8(pick(p, "pv1_30"), ""), dateRule())
	b.setField(s, 31, pick(p, "pv1_31"), &ValidationRule{Length: lenMinMax(1, 10)})
	b.setField(s, 32, pick(p, "pv1_32"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 33, pick(p, "pv1_33"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 34, pick(p, "pv1_34"), &ValidationRule{Length: lenExact(3)})
	b.setField(s, 35, b.dv8(pick(p, "pv1_35"), ""), dateRule())
	b.setField(s, 36, pick(p, "pv1_36", "dischargeDisposition"), &ValidationRule{Length: lenExact(3)})
	b.setField(s, 37, pick(p, "pv1_37"), &ValidationRule{Length: lenMinMax(1, 25)})
	b.setField(s, 38, pick(p, "pv1_38"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 39, pick(p, "pv1_39"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 40, pick(p, "pv1_40", "bedStatus"), &ValidationRule{AllowedValues: b.codeTable("0116"), Length: lenExact(1)})
	b.setField(s, 41, pick(p, "pv1_41"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 42, pick(p, "pv1_42"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 43, pick(p, "pv1_43"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 44, b.dv(pick(p, "pv1_44", "admitDateTime"), ""), dateRule())

	// Version extensions (the HL7_2_2._buildPV1, gated by the usage catalog).
	if compareVersions(b.version, "2.2") >= 0 {
		b.setField(s, 45, b.dv(pick(p, "pv1_45"), ""), dateRule())
		b.setField(s, 46, pick(p, "pv1_46"), &ValidationRule{Length: lenMinMax(1, 12)})
		b.setField(s, 47, pick(p, "pv1_47"), &ValidationRule{Length: lenMinMax(1, 12)})
		b.setField(s, 48, pick(p, "pv1_48"), &ValidationRule{Length: lenMinMax(1, 12)})
		b.setField(s, 49, pick(p, "pv1_49"), &ValidationRule{Length: lenMinMax(1, 12)})
		b.setField(s, 50, pick(p, "pv1_50"), &ValidationRule{Length: lenMinMax(1, 15)})
		b.setField(s, 51, pick(p, "pv1_51"), &ValidationRule{AllowedValues: b.codeTable("0326"), Length: lenExact(1)})
		b.setField(s, 52, pick(p, "pv1_52"), &ValidationRule{Length: lenMinMax(1, 60)})
	}
	return b
}

// BuildQRD builds a QRD (Query Definition) segment (the _buildQRD).
func (b *HL7_BASE) BuildQRD(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("QRD")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "QRD")
	b.setField(s, 1, b.dv(pick(p, "qrd_1", "queryDateTime"), ""), dateRule())
	b.setField(s, 2, pick(p, "qrd_2", "queryFormatCode"), &ValidationRule{Length: lenExact(1)})
	b.setField(s, 3, pick(p, "qrd_3", "queryPriority"), &ValidationRule{Length: lenExact(1)})
	b.setField(s, 4, pick(p, "qrd_4", "queryId"), &ValidationRule{Length: lenMinMax(1, 10)})
	b.setField(s, 5, pick(p, "qrd_5", "deferredResponseType"), &ValidationRule{Length: lenExact(1)})
	b.setField(s, 6, b.dv(pick(p, "qrd_6"), ""), dateRule())
	b.setField(s, 7, pick(p, "qrd_7", "quantityLimitedRequest"), &ValidationRule{Length: lenMinMax(1, 10)})
	b.setField(s, 8, pick(p, "qrd_8", "whoSubjectFilter"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 9, pick(p, "qrd_9", "whatSubjectFilter"), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 10, pick(p, "qrd_10", "whatDepartmentDataCode"), &ValidationRule{Length: lenMinMax(1, 20)})
	return b
}

// BuildQRF builds a QRF (Query Filter) segment (the _buildQRF). Chainable.
func (b *HL7_BASE) BuildQRF(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("QRF")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "QRF")
	b.setField(s, 1, pick(p, "qrf_1", "whereSubjectFilter"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 2, b.dv(pick(p, "qrf_2", "whenDataStartDateTime"), ""), dateRule())
	b.setField(s, 3, b.dv(pick(p, "qrf_3", "whenDataEndDateTime"), ""), dateRule())
	b.setField(s, 4, pick(p, "qrf_4", "whatUserQualifier"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 5, pick(p, "qrf_5", "otherQrySubjectFilter"), &ValidationRule{Length: lenMinMax(1, 20)})
	return b
}

// BuildRX1 builds an RX1 (Pharmacy/Treatment Order) segment (the _buildRX1).
func (b *HL7_BASE) BuildRX1(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("RX1")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "RX1")
	b.setField(s, 1, pick(p, "rx1_1"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 2, pick(p, "rx1_2", "giveCode"), &ValidationRule{Length: lenMinMax(1, 100)})
	b.setField(s, 3, pick(p, "rx1_3", "giveAmountMinimum"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 4, pick(p, "rx1_4", "giveAmountMaximum"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 5, pick(p, "rx1_5", "giveUnits"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 6, pick(p, "rx1_6", "giveDosageForm"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 7, pick(p, "rx1_7", "providersPharmacyInstructions"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 8, pick(p, "rx1_8", "providersAdministrationInstructions"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 9, pick(p, "rx1_9", "deliverToLocation"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 10, pick(p, "rx1_10", "allowSubstitutions"), &ValidationRule{Length: lenMinMax(1, 1)})
	b.setField(s, 11, pick(p, "rx1_11", "requestedGiveCode"), &ValidationRule{Length: lenMinMax(1, 100)})
	b.setField(s, 12, pick(p, "rx1_12", "requestedGiveAmountMinimum"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 13, pick(p, "rx1_13", "requestedGiveAmountMaximum"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 14, pick(p, "rx1_14", "requestedGiveUnits"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 15, pick(p, "rx1_15", "requestedDosageForm"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 16, pick(p, "rx1_16", "pharmacistSpecialDispensingInstructions"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 17, pick(p, "rx1_17", "givePer"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 18, pick(p, "rx1_18", "giveRateAmount"), &ValidationRule{Length: lenMinMax(1, 6)})
	b.setField(s, 19, pick(p, "rx1_19", "giveRateUnits"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 20, pick(p, "rx1_20"), &ValidationRule{Length: lenMinMax(1, 6)})
	b.setField(s, 21, pick(p, "rx1_21"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 22, pick(p, "rx1_22", "totalDailyDose"), &ValidationRule{Length: lenMinMax(1, 10)})
	return b
}

// BuildUB1 builds a UB1 (UB82 Data) segment (the _buildUB1). Chainable.
func (b *HL7_BASE) BuildUB1(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("UB1")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "UB1")
	b.setField(s, 1, pick(p, "ub1_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "ub1_2", "bloodDeductible"), &ValidationRule{Length: lenMinMax(1, 1)})
	b.setField(s, 3, pick(p, "ub1_3", "bloodFurnishedPints"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 4, pick(p, "ub1_4", "bloodReplacedPints"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 5, pick(p, "ub1_5", "bloodNotReplacedPints"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 6, pick(p, "ub1_6", "coInsuranceDays"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 7, pick(p, "ub1_7", "conditionCode"), &ValidationRule{Length: lenMinMax(1, 14)})
	b.setField(s, 8, pick(p, "ub1_8", "coveredDays"), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 9, pick(p, "ub1_9", "nonCoveredDays"), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 10, pick(p, "ub1_10", "valueAmountCode"), &ValidationRule{Length: lenMinMax(1, 55)})
	b.setField(s, 11, pick(p, "ub1_11", "numberOfGraceDays"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 12, pick(p, "ub1_12", "specialProgramIndicator"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 13, pick(p, "ub1_13"), &ValidationRule{Length: lenExact(2)})
	b.setField(s, 14, b.dv8(pick(p, "ub1_14"), ""), dateRule())
	b.setField(s, 15, b.dv8(pick(p, "ub1_15"), ""), dateRule())
	b.setField(s, 16, pick(p, "ub1_16", "occurrence"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 17, pick(p, "ub1_17", "occurrenceSpan"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 18, b.dv8(pick(p, "ub1_18"), ""), dateRule())
	b.setField(s, 19, b.dv8(pick(p, "ub1_19"), ""), dateRule())
	b.setField(s, 20, pick(p, "ub1_20"), &ValidationRule{Length: lenMinMax(1, 30)})
	b.setField(s, 21, pick(p, "ub1_21"), &ValidationRule{Length: lenMinMax(1, 7)})
	b.setField(s, 22, pick(p, "ub1_22"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 23, pick(p, "ub1_23"), &ValidationRule{Length: lenMinMax(1, 17)})
	return b
}

// BuildURD builds a URD (Results/Update Definition) segment (the _buildURD).
func (b *HL7_BASE) BuildURD(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("URD")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "URD")
	b.setField(s, 1, b.dv(pick(p, "urd_1", "ruDateTime"), ""), dateRule())
	b.setField(s, 2, pick(p, "urd_2", "reportPriority"), &ValidationRule{Length: lenExact(1)})
	b.setField(s, 3, pick(p, "urd_3", "ruWhoSubjectDefinition"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 4, pick(p, "urd_4", "ruWhatSubjectDefinition"), &ValidationRule{Length: lenMinMax(1, 3)})
	return b
}

// BuildURS builds a URS (Unsolicited Selection) segment (the _buildURS).
func (b *HL7_BASE) BuildURS(p Props) *HL7_BASE {
	b.headerExists()
	s := spec("URS")
	b.assertSegmentInVersion(s)
	b.segment = mustAddSegment(b.message, "URS")
	b.setField(s, 1, pick(p, "urs_1", "ruWhereSubjectDefinition"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 2, b.dv(pick(p, "urs_2"), ""), dateRule())
	b.setField(s, 3, b.dv(pick(p, "urs_3"), ""), dateRule())
	b.setField(s, 4, pick(p, "urs_4", "ruWhatUserQualifier"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 5, pick(p, "urs_5"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 6, pick(p, "urs_6"), &ValidationRule{Length: lenMinMax(1, 12)})
	return b
}
