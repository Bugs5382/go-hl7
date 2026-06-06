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

import "fmt"

// These BuildXXX methods build the v2.2 segments
// (AL1, MFE, MFI, ODS, ODT, RXA, RXD, RXE, RXG, RXO, RXR, STF, UB2). Each is a
// validatorSetField sequence over the shared base; the segment did not exist
// before v2.2, so the version guard rejects it on earlier versions. The
// OBR/OBX/ORC/PV1 version extensions added in v2.2 live in
// build_segments_v21.go, gated by the usage catalog.

// jsString stringifies v, coercing an absent value to the literal "undefined".
// It is used for the set-ID field of several segments, which is always
// populated in practice.
func jsString(v any) string {
	if v == nil {
		return "undefined"
	}
	return fmt.Sprint(v)
}

// jsStringOr stringifies v, coercing an absent value to "".
func jsStringOr(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

// BuildAL1 builds an AL1 (Allergy Information) segment.
// Introduced in v2.2. Chainable.
func (b *Builder) BuildAL1(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.2")
	s := b.spec("AL1")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("AL1")
	b.setField(s, 1, jsString(pick(p, "al1_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "al1_2"), &ValidationRule{AllowedValues: b.codeTable("0127")})
	b.setField(s, 3, pick(p, "al1_3"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 4, pick(p, "al1_4"), &ValidationRule{AllowedValues: b.codeTable("0128")})
	b.setField(s, 5, pick(p, "al1_5"), &ValidationRule{Length: lenMinMax(1, 15)})
	b.setField(s, 6, b.dv(pick(p, "al1_6"), ""), dateRule())
	return b
}

// BuildMFE builds an MFE (Master File Entry) segment.
// Chainable.
func (b *Builder) BuildMFE(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.2")
	s := b.spec("MFE")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("MFE")
	b.setField(s, 1, pick(p, "mfe_1"), &ValidationRule{AllowedValues: []string{"MAD", "MDC", "MDL", "MUP", "MAC"}})
	b.setField(s, 2, pick(p, "mfe_2"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 3, b.dv(pick(p, "mfe_3"), ""), dateRule())
	b.setField(s, 4, pick(p, "mfe_4"), &ValidationRule{Length: lenMinMax(1, 200)})
	return b
}

// BuildMFI builds an MFI (Master File Identification) segment. Chainable.
func (b *Builder) BuildMFI(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.2")
	s := b.spec("MFI")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("MFI")
	b.setField(s, 1, pick(p, "mfi_1"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 2, pick(p, "mfi_2"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 3, pick(p, "mfi_3"), &ValidationRule{AllowedValues: []string{"REP", "UPD"}})
	b.setField(s, 4, b.dv(pick(p, "mfi_4"), ""), dateRule())
	b.setField(s, 5, b.dv(pick(p, "mfi_5"), ""), dateRule())
	b.setField(s, 6, pick(p, "mfi_6"), &ValidationRule{AllowedValues: []string{"AL", "ER", "NE", "NR"}})
	return b
}

// BuildODS builds an ODS (Dietary Orders) segment.
// Chainable.
func (b *Builder) BuildODS(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.2")
	s := b.spec("ODS")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("ODS")
	b.setField(s, 1, pick(p, "ods_1"), &ValidationRule{AllowedValues: []string{"D", "S", "P"}, Length: lenExact(1)})
	b.setField(s, 2, pick(p, "ods_2"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 3, pick(p, "ods_3"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 4, pick(p, "ods_4"), &ValidationRule{Length: lenMinMax(1, 60)})
	return b
}

// BuildODT builds an ODT (Diet Tray Instructions) segment. Chainable.
func (b *Builder) BuildODT(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.2")
	s := b.spec("ODT")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("ODT")
	b.setField(s, 1, pick(p, "odt_1"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 2, pick(p, "odt_2"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 3, pick(p, "odt_3"), &ValidationRule{Length: lenMinMax(1, 60)})
	return b
}

// BuildRXA builds an RXA (Pharmacy/Treatment Administration) segment. Chainable.
func (b *Builder) BuildRXA(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.2")
	s := b.spec("RXA")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("RXA")
	b.setField(s, 1, strOrNil(p, "rxa_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, strOrNil(p, "rxa_2"), nil)
	b.setField(s, 3, b.dv(pick(p, "rxa_3"), ""), dateRule())
	b.setField(s, 4, b.dv(pick(p, "rxa_4"), ""), dateRule())
	b.setField(s, 5, pick(p, "rxa_5"), &ValidationRule{Length: lenMinMax(1, 100)})
	b.setField(s, 6, pick(p, "rxa_6"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 7, pick(p, "rxa_7"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 8, pick(p, "rxa_8"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 9, pick(p, "rxa_9"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 10, pick(p, "rxa_10"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 11, pick(p, "rxa_11"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 12, pick(p, "rxa_12"), &ValidationRule{Length: lenMinMax(1, 20)})
	return b
}

// BuildRXD builds an RXD (Pharmacy/Treatment Dispense) segment. Chainable.
func (b *Builder) BuildRXD(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.2")
	s := b.spec("RXD")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("RXD")
	b.setField(s, 1, pick(p, "rxd_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "rxd_2"), &ValidationRule{Length: lenMinMax(1, 100)})
	b.setField(s, 3, b.dv(pick(p, "rxd_3"), ""), dateRule())
	b.setField(s, 4, pick(p, "rxd_4"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 5, pick(p, "rxd_5"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 6, pick(p, "rxd_6"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 7, pick(p, "rxd_7"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 8, pick(p, "rxd_8"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 9, pick(p, "rxd_9"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 10, pick(p, "rxd_10"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 11, pick(p, "rxd_11"), &ValidationRule{Length: lenMinMax(1, 1)})
	b.setField(s, 12, pick(p, "rxd_12"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 13, pick(p, "rxd_13"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 14, pick(p, "rxd_14"), &ValidationRule{AllowedValues: []string{"Y", "N"}, Length: lenExact(1)})
	b.setField(s, 15, pick(p, "rxd_15"), &ValidationRule{Length: lenMinMax(1, 200)})
	return b
}

// BuildRXE builds an RXE (Pharmacy/Treatment Encoded Order) segment. Chainable.
func (b *Builder) BuildRXE(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.2")
	s := b.spec("RXE")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("RXE")
	b.setField(s, 1, pick(p, "rxe_1"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 2, pick(p, "rxe_2"), &ValidationRule{Length: lenMinMax(1, 100)})
	b.setField(s, 3, pick(p, "rxe_3"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 4, pick(p, "rxe_4"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 5, pick(p, "rxe_5"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 6, pick(p, "rxe_6"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 7, pick(p, "rxe_7"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 8, pick(p, "rxe_8"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 9, pick(p, "rxe_9"), &ValidationRule{Length: lenMinMax(1, 1)})
	b.setField(s, 10, pick(p, "rxe_10"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 11, pick(p, "rxe_11"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 12, pick(p, "rxe_12"), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 13, pick(p, "rxe_13"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 14, pick(p, "rxe_14"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 15, pick(p, "rxe_15"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 16, pick(p, "rxe_16"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 17, pick(p, "rxe_17"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 18, b.dv(pick(p, "rxe_18"), ""), dateRule())
	b.setField(s, 19, pick(p, "rxe_19"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 20, pick(p, "rxe_20"), &ValidationRule{AllowedValues: []string{"Y", "N"}, Length: lenExact(1)})
	b.setField(s, 21, pick(p, "rxe_21"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 22, pick(p, "rxe_22"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 23, pick(p, "rxe_23"), &ValidationRule{Length: lenMinMax(1, 6)})
	b.setField(s, 24, pick(p, "rxe_24"), &ValidationRule{Length: lenMinMax(1, 60)})
	return b
}

// BuildRXG builds an RXG (Pharmacy/Treatment Give) segment. Chainable.
func (b *Builder) BuildRXG(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.2")
	s := b.spec("RXG")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("RXG")
	b.setField(s, 1, pick(p, "rxg_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "rxg_2"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 3, pick(p, "rxg_3"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 4, pick(p, "rxg_4"), &ValidationRule{Length: lenMinMax(1, 100)})
	b.setField(s, 5, pick(p, "rxg_5"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 6, pick(p, "rxg_6"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 7, pick(p, "rxg_7"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 8, pick(p, "rxg_8"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 9, pick(p, "rxg_9"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 10, pick(p, "rxg_10"), &ValidationRule{Length: lenMinMax(1, 1)})
	b.setField(s, 11, pick(p, "rxg_11"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 12, pick(p, "rxg_12"), &ValidationRule{AllowedValues: []string{"Y", "N"}, Length: lenExact(1)})
	b.setField(s, 13, pick(p, "rxg_13"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 14, pick(p, "rxg_14"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 15, pick(p, "rxg_15"), &ValidationRule{Length: lenMinMax(1, 6)})
	b.setField(s, 16, pick(p, "rxg_16"), &ValidationRule{Length: lenMinMax(1, 60)})
	return b
}

// BuildRXO builds an RXO (Pharmacy/Treatment Order) segment. Chainable.
func (b *Builder) BuildRXO(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.2")
	s := b.spec("RXO")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("RXO")
	b.setField(s, 1, pick(p, "rxo_1"), &ValidationRule{Length: lenMinMax(1, 100)})
	b.setField(s, 2, pick(p, "rxo_2"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 3, pick(p, "rxo_3"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 4, pick(p, "rxo_4"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 5, pick(p, "rxo_5"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 6, pick(p, "rxo_6"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 7, pick(p, "rxo_7"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 8, pick(p, "rxo_8"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 9, pick(p, "rxo_9"), &ValidationRule{AllowedValues: []string{"G", "N"}, Length: lenExact(1)})
	b.setField(s, 10, pick(p, "rxo_10"), &ValidationRule{Length: lenMinMax(1, 100)})
	b.setField(s, 11, pick(p, "rxo_11"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 12, pick(p, "rxo_12"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 13, pick(p, "rxo_13"), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 14, pick(p, "rxo_14"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 15, pick(p, "rxo_15"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 16, pick(p, "rxo_16"), &ValidationRule{AllowedValues: []string{"Y", "N"}, Length: lenExact(1)})
	b.setField(s, 17, pick(p, "rxo_17"), &ValidationRule{Length: lenMinMax(1, 20)})
	return b
}

// BuildRXR builds an RXR (Pharmacy/Treatment Route) segment. Chainable.
func (b *Builder) BuildRXR(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.2")
	s := b.spec("RXR")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("RXR")
	b.setField(s, 1, pick(p, "rxr_1"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 2, pick(p, "rxr_2"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 3, pick(p, "rxr_3"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 4, pick(p, "rxr_4"), &ValidationRule{Length: lenMinMax(1, 60)})
	return b
}

// BuildSTF builds an STF (Staff Identification) segment. Chainable.
func (b *Builder) BuildSTF(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.2")
	s := b.spec("STF")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("STF")
	b.setField(s, 1, pick(p, "stf_1"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 2, pick(p, "stf_2", "staffIdCode"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 3, pick(p, "stf_3", "staffName"), &ValidationRule{Length: lenMinMax(1, 48)})
	b.setField(s, 4, pick(p, "stf_4"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 5, pick(p, "stf_5"), &ValidationRule{AllowedValues: b.codeTable("0001"), Length: lenExact(1)})
	b.setField(s, 6, b.dv(pick(p, "stf_6"), ""), dateRule())
	b.setField(s, 7, pick(p, "stf_7"), &ValidationRule{AllowedValues: []string{"A", "I"}, Length: lenExact(1)})
	b.setField(s, 8, pick(p, "stf_8"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 9, pick(p, "stf_9"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 10, pick(p, "stf_10"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 11, pick(p, "stf_11"), &ValidationRule{Length: lenMinMax(1, 106)})
	b.setField(s, 12, b.dv(pick(p, "stf_12"), ""), dateRule())
	b.setField(s, 13, b.dv(pick(p, "stf_13"), ""), dateRule())
	b.setField(s, 14, pick(p, "stf_14"), &ValidationRule{Length: lenMinMax(1, 60)})
	return b
}

// BuildUB2 builds a UB2 (UB92 Data) segment.
// Chainable.
func (b *Builder) BuildUB2(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.2")
	s := b.spec("UB2")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("UB2")
	b.setField(s, 1, strOrNil(p, "ub2_1"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "ub2_2"), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 3, pick(p, "ub2_3"), &ValidationRule{Length: lenMinMax(1, 14)})
	b.setField(s, 4, pick(p, "ub2_4"), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 5, pick(p, "ub2_5"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 6, pick(p, "ub2_6"), nil)
	b.setField(s, 7, pick(p, "ub2_7"), nil)
	b.setField(s, 8, pick(p, "ub2_8"), nil)
	b.setField(s, 9, pick(p, "ub2_9"), &ValidationRule{Length: lenMinMax(1, 29)})
	b.setField(s, 10, pick(p, "ub2_10"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 11, pick(p, "ub2_11"), &ValidationRule{Length: lenMinMax(1, 5)})
	b.setField(s, 12, pick(p, "ub2_12"), &ValidationRule{Length: lenMinMax(1, 23)})
	b.setField(s, 13, pick(p, "ub2_13"), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 14, pick(p, "ub2_14"), &ValidationRule{Length: lenMinMax(1, 14)})
	b.setField(s, 15, pick(p, "ub2_15"), &ValidationRule{Length: lenMinMax(1, 27)})
	b.setField(s, 16, pick(p, "ub2_16"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 17, pick(p, "ub2_17"), &ValidationRule{Length: lenMinMax(1, 3)})
	return b
}

// strOrNil returns the string form of a present prop value, or nil when absent.
func strOrNil(p Props, keys ...string) any {
	v := pick(p, keys...)
	if v == nil {
		return nil
	}
	return fmt.Sprint(v)
}
