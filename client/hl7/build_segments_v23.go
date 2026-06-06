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

// These BuildXXX methods build the v2.3 segments (the
// scheduling/clinical-study/provider segments introduced in v2.3). Each is a
// validatorSetField sequence over the shared base; the version guard rejects
// the segment on earlier versions. The NK1/OBR/OBX/ORC/PID version extensions
// added in v2.3 live in build_segments_v21.go, gated by the usage catalog.

// aigStyle builds the AIG/AIL/AIP/AIS-shared field layout, repeated
// verbatim across those four scheduling segments. The date field index differs
// (AIS dates field 4, the others field 6/8), so the shared body is only the
// common prefix; each builder supplies its own field sequence.

// BuildAIG builds an AIG (Appointment Information - General Resource) segment.
// Chainable.
func (b *Builder) BuildAIG(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("AIG")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("AIG")
	b.setField(s, 1, jsString(pick(p, "aig_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "aig_2"), &ValidationRule{AllowedValues: []string{"A", "D", "U"}, Length: lenExact(1)})
	b.setField(s, 3, pick(p, "aig_3"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 4, pick(p, "aig_4"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 5, pick(p, "aig_5"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 6, pick(p, "aig_6"), &ValidationRule{Length: lenMinMax(1, 5)})
	b.setField(s, 7, pick(p, "aig_7"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 8, b.dv(pick(p, "aig_8"), ""), dateRule())
	b.setField(s, 9, pick(p, "aig_9"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 10, pick(p, "aig_10"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 11, pick(p, "aig_11"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 12, pick(p, "aig_12"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 13, pick(p, "aig_13"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 14, pick(p, "aig_14"), &ValidationRule{Length: lenMinMax(1, 200)})
	return b
}

// BuildAIL builds an AIL (Appointment Information - Location Resource) segment.
// Chainable.
func (b *Builder) BuildAIL(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("AIL")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("AIL")
	b.setField(s, 1, jsString(pick(p, "ail_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "ail_2"), &ValidationRule{AllowedValues: []string{"A", "D", "U"}, Length: lenExact(1)})
	b.setField(s, 3, pick(p, "ail_3"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 4, pick(p, "ail_4"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 5, pick(p, "ail_5"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 6, b.dv(pick(p, "ail_6"), ""), dateRule())
	b.setField(s, 7, pick(p, "ail_7"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 8, pick(p, "ail_8"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 9, pick(p, "ail_9"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 10, pick(p, "ail_10"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 11, pick(p, "ail_11"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 12, pick(p, "ail_12"), &ValidationRule{Length: lenMinMax(1, 200)})
	return b
}

// BuildAIP builds an AIP (Appointment Information - Personnel Resource) segment.
// Chainable.
func (b *Builder) BuildAIP(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("AIP")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("AIP")
	b.setField(s, 1, jsString(pick(p, "aip_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "aip_2"), &ValidationRule{AllowedValues: []string{"A", "D", "U"}, Length: lenExact(1)})
	b.setField(s, 3, pick(p, "aip_3"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 4, pick(p, "aip_4"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 5, pick(p, "aip_5"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 6, b.dv(pick(p, "aip_6"), ""), dateRule())
	b.setField(s, 7, pick(p, "aip_7"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 8, pick(p, "aip_8"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 9, pick(p, "aip_9"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 10, pick(p, "aip_10"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 11, pick(p, "aip_11"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 12, pick(p, "aip_12"), &ValidationRule{Length: lenMinMax(1, 200)})
	return b
}

// BuildAIS builds an AIS (Appointment Information - Service) segment. Chainable.
func (b *Builder) BuildAIS(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("AIS")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("AIS")
	b.setField(s, 1, jsString(pick(p, "ais_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "ais_2"), &ValidationRule{AllowedValues: []string{"A", "D", "U"}, Length: lenExact(1)})
	b.setField(s, 3, pick(p, "ais_3"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 4, b.dv(pick(p, "ais_4"), ""), dateRule())
	b.setField(s, 5, pick(p, "ais_5"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 6, pick(p, "ais_6"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 7, pick(p, "ais_7"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 8, pick(p, "ais_8"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 9, pick(p, "ais_9"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 10, pick(p, "ais_10"), &ValidationRule{Length: lenMinMax(1, 200)})
	return b
}

// BuildAPR builds an APR (Appointment Preferences) segment. Chainable.
func (b *Builder) BuildAPR(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("APR")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("APR")
	b.setField(s, 1, pick(p, "apr_1"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 2, pick(p, "apr_2"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 3, pick(p, "apr_3"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 4, pick(p, "apr_4"), &ValidationRule{Length: lenMinMax(1, 5)})
	b.setField(s, 5, pick(p, "apr_5"), &ValidationRule{Length: lenMinMax(1, 200)})
	return b
}

// BuildCSP builds a CSP (Clinical Study Phase) segment. Chainable.
func (b *Builder) BuildCSP(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("CSP")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("CSP")
	b.setField(s, 1, pick(p, "csp_1"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 2, b.dv(pick(p, "csp_2"), ""), dateRule())
	b.setField(s, 3, b.dv(pick(p, "csp_3"), ""), dateRule())
	b.setField(s, 4, pick(p, "csp_4"), &ValidationRule{Length: lenMinMax(1, 60)})
	return b
}

// BuildCSR builds a CSR (Clinical Study Registration) segment. Chainable.
func (b *Builder) BuildCSR(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("CSR")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("CSR")
	b.setField(s, 1, pick(p, "csr_1"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 2, pick(p, "csr_2"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 3, pick(p, "csr_3"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 4, pick(p, "csr_4"), &ValidationRule{Length: lenMinMax(1, 30)})
	b.setField(s, 5, pick(p, "csr_5"), &ValidationRule{Length: lenMinMax(1, 30)})
	b.setField(s, 6, b.dv(pick(p, "csr_6"), ""), dateRule())
	b.setField(s, 7, pick(p, "csr_7"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 8, pick(p, "csr_8"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 9, b.dv(pick(p, "csr_9"), ""), dateRule())
	b.setField(s, 10, pick(p, "csr_10"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 11, b.dv(pick(p, "csr_11"), ""), dateRule())
	b.setField(s, 12, pick(p, "csr_12"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 13, pick(p, "csr_13"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 14, pick(p, "csr_14"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 15, b.dv(pick(p, "csr_15"), ""), dateRule())
	b.setField(s, 16, pick(p, "csr_16"), &ValidationRule{Length: lenMinMax(1, 60)})
	return b
}

// BuildCSS builds a CSS (Clinical Study Data Schedule) segment. Chainable.
func (b *Builder) BuildCSS(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("CSS")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("CSS")
	b.setField(s, 1, pick(p, "css_1"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 2, b.dv(pick(p, "css_2"), ""), dateRule())
	b.setField(s, 3, pick(p, "css_3"), &ValidationRule{Length: lenMinMax(1, 60)})
	return b
}

// BuildCTD builds a CTD (Contact Data) segment.
// Chainable.
func (b *Builder) BuildCTD(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("CTD")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("CTD")
	b.setField(s, 1, pick(p, "ctd_1"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 2, pick(p, "ctd_2"), &ValidationRule{Length: lenMinMax(1, 106)})
	b.setField(s, 3, pick(p, "ctd_3"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 4, pick(p, "ctd_4"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 5, pick(p, "ctd_5"), &ValidationRule{Length: lenMinMax(1, 100)})
	b.setField(s, 6, pick(p, "ctd_6"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 7, pick(p, "ctd_7"), &ValidationRule{Length: lenMinMax(1, 100)})
	return b
}

// BuildPCR builds a PCR (Possible Causal Relationship) segment. Chainable.
func (b *Builder) BuildPCR(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("PCR")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("PCR")
	b.setField(s, 1, pick(p, "pcr_1"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 2, pick(p, "pcr_2"), &ValidationRule{AllowedValues: []string{"Y", "N", "NA"}, Length: lenMinMax(1, 2)})
	b.setField(s, 3, pick(p, "pcr_3"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 4, pick(p, "pcr_4"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 5, b.dv(pick(p, "pcr_5"), ""), dateRule())
	b.setField(s, 6, b.dv(pick(p, "pcr_6"), ""), dateRule())
	b.setField(s, 7, b.dv(pick(p, "pcr_7"), ""), dateRule())
	b.setField(s, 8, b.dv(pick(p, "pcr_8"), ""), dateRule())
	b.setField(s, 9, pick(p, "pcr_9"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 10, pick(p, "pcr_10"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 11, pick(p, "pcr_11"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 12, pick(p, "pcr_12"), &ValidationRule{Length: lenMinMax(1, 30)})
	b.setField(s, 13, pick(p, "pcr_13"), &ValidationRule{Length: lenMinMax(1, 1)})
	b.setField(s, 14, pick(p, "pcr_14"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 15, pick(p, "pcr_15"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 16, pick(p, "pcr_16"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 17, pick(p, "pcr_17"), &ValidationRule{Length: lenMinMax(1, 8)})
	b.setField(s, 18, b.dv(pick(p, "pcr_18"), ""), dateRule())
	b.setField(s, 19, pick(p, "pcr_19"), &ValidationRule{Length: lenMinMax(1, 1)})
	b.setField(s, 20, pick(p, "pcr_20"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 21, pick(p, "pcr_21"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 22, pick(p, "pcr_22"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 23, pick(p, "pcr_23"), &ValidationRule{Length: lenMinMax(1, 1)})
	return b
}

// BuildPD1 builds a PD1 (Patient Additional Demographic) segment. Chainable.
func (b *Builder) BuildPD1(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("PD1")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("PD1")
	b.setField(s, 1, pick(p, "pd1_1"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 2, pick(p, "pd1_2"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 3, pick(p, "pd1_3"), &ValidationRule{Length: lenMinMax(1, 90)})
	b.setField(s, 4, pick(p, "pd1_4"), &ValidationRule{Length: lenMinMax(1, 90)})
	b.setField(s, 5, pick(p, "pd1_5"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 6, pick(p, "pd1_6"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 7, pick(p, "pd1_7"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 8, pick(p, "pd1_8"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 9, pick(p, "pd1_9"), &ValidationRule{AllowedValues: []string{"Y", "N"}, Length: lenExact(1)})
	b.setField(s, 10, pick(p, "pd1_10"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 11, pick(p, "pd1_11"), &ValidationRule{Length: lenMinMax(1, 1)})
	b.setField(s, 12, pick(p, "pd1_12"), &ValidationRule{AllowedValues: []string{"Y", "N"}, Length: lenExact(1)})
	return b
}

// BuildPRA builds a PRA (Practitioner Detail) segment. Chainable.
func (b *Builder) BuildPRA(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("PRA")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("PRA")
	b.setField(s, 1, pick(p, "pra_1"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 2, pick(p, "pra_2"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 3, pick(p, "pra_3"), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 4, pick(p, "pra_4"), &ValidationRule{AllowedValues: []string{"I", "O"}, Length: lenExact(1)})
	b.setField(s, 5, pick(p, "pra_5"), &ValidationRule{Length: lenMinMax(1, 100)})
	b.setField(s, 6, pick(p, "pra_6"), &ValidationRule{Length: lenMinMax(1, 100)})
	b.setField(s, 7, pick(p, "pra_7"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 8, b.dv(pick(p, "pra_8"), ""), dateRule())
	return b
}

// BuildPRD builds a PRD (Provider Data) segment.
// Chainable.
func (b *Builder) BuildPRD(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("PRD")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("PRD")
	b.setField(s, 1, pick(p, "prd_1"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 2, pick(p, "prd_2"), &ValidationRule{Length: lenMinMax(1, 106)})
	b.setField(s, 3, pick(p, "prd_3"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 4, pick(p, "prd_4"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 5, pick(p, "prd_5"), &ValidationRule{Length: lenMinMax(1, 100)})
	b.setField(s, 6, pick(p, "prd_6"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 7, pick(p, "prd_7"), &ValidationRule{Length: lenMinMax(1, 100)})
	b.setField(s, 8, b.dv(pick(p, "prd_8"), ""), dateRule())
	b.setField(s, 9, b.dv(pick(p, "prd_9"), ""), dateRule())
	return b
}

// BuildPSH builds a PSH (Product Summary Header) segment. Chainable.
func (b *Builder) BuildPSH(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("PSH")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("PSH")
	b.setField(s, 1, pick(p, "psh_1"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 2, pick(p, "psh_2"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 3, b.dv(pick(p, "psh_3"), ""), dateRule())
	b.setField(s, 4, b.dv(pick(p, "psh_4"), ""), dateRule())
	b.setField(s, 5, b.dv(pick(p, "psh_5"), ""), dateRule())
	b.setField(s, 6, pick(p, "psh_6"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 7, pick(p, "psh_7"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 8, pick(p, "psh_8"), &ValidationRule{AllowedValues: []string{"A", "E"}, Length: lenExact(1)})
	b.setField(s, 9, pick(p, "psh_9"), &ValidationRule{Length: lenMinMax(1, 600)})
	b.setField(s, 10, pick(p, "psh_10"), &ValidationRule{Length: lenMinMax(1, 12)})
	b.setField(s, 11, pick(p, "psh_11"), &ValidationRule{AllowedValues: []string{"A", "E"}, Length: lenExact(1)})
	b.setField(s, 12, pick(p, "psh_12"), &ValidationRule{Length: lenMinMax(1, 600)})
	b.setField(s, 13, pick(p, "psh_13"), &ValidationRule{Length: lenMinMax(1, 2)})
	b.setField(s, 14, pick(p, "psh_14"), &ValidationRule{Length: lenMinMax(1, 2)})
	return b
}

// BuildRDF builds an RDF (Table Row Definition) segment. Chainable.
func (b *Builder) BuildRDF(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("RDF")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("RDF")
	b.setField(s, 1, pick(p, "rdf_1"), &ValidationRule{Length: lenMinMax(1, 3)})
	b.setField(s, 2, pick(p, "rdf_2"), &ValidationRule{Length: lenMinMax(1, 40)})
	return b
}

// BuildRDT builds an RDT (Table Row Data) segment.
// Chainable.
func (b *Builder) BuildRDT(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("RDT")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("RDT")
	b.setField(s, 1, pick(p, "rdt_1"), &ValidationRule{Length: lenMinMax(1, 99999)})
	return b
}

// BuildRGS builds an RGS (Resource Group) segment.
// Chainable.
func (b *Builder) BuildRGS(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("RGS")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("RGS")
	b.setField(s, 1, jsString(pick(p, "rgs_1")), &ValidationRule{Length: lenMinMax(1, 4)})
	b.setField(s, 2, pick(p, "rgs_2"), &ValidationRule{AllowedValues: []string{"A", "D", "U"}, Length: lenExact(1)})
	b.setField(s, 3, pick(p, "rgs_3"), &ValidationRule{Length: lenMinMax(1, 200)})
	return b
}

// BuildROL builds a ROL (Role) segment. Chainable.
func (b *Builder) BuildROL(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("ROL")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("ROL")
	b.setField(s, 1, pick(p, "rol_1"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 2, pick(p, "rol_2"), &ValidationRule{AllowedValues: []string{"A", "D", "U"}})
	b.setField(s, 3, pick(p, "rol_3"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 4, pick(p, "rol_4"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 5, b.dv(pick(p, "rol_5"), ""), dateRule())
	b.setField(s, 6, b.dv(pick(p, "rol_6"), ""), dateRule())
	b.setField(s, 7, pick(p, "rol_7"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 8, pick(p, "rol_8"), &ValidationRule{Length: lenMinMax(1, 200)})
	return b
}

// BuildSCH builds an SCH (Scheduling Activity Information) segment. Chainable.
func (b *Builder) BuildSCH(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("SCH")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("SCH")
	b.setField(s, 1, pick(p, "sch_1"), &ValidationRule{Length: lenMinMax(1, 75)})
	b.setField(s, 2, pick(p, "sch_2"), &ValidationRule{Length: lenMinMax(1, 75)})
	b.setField(s, 3, pick(p, "sch_3"), &ValidationRule{Length: lenMinMax(1, 5)})
	b.setField(s, 4, pick(p, "sch_4"), &ValidationRule{Length: lenMinMax(1, 22)})
	b.setField(s, 5, pick(p, "sch_5"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 6, pick(p, "sch_6"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 7, pick(p, "sch_7"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 8, pick(p, "sch_8"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 9, pick(p, "sch_9"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.setField(s, 10, pick(p, "sch_10"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 11, pick(p, "sch_11"), &ValidationRule{Length: lenMinMax(1, 200)})
	b.setField(s, 12, pick(p, "sch_12"), &ValidationRule{Length: lenMinMax(1, 48)})
	b.setField(s, 13, pick(p, "sch_13"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 14, pick(p, "sch_14"), &ValidationRule{Length: lenMinMax(1, 106)})
	b.setField(s, 15, pick(p, "sch_15"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 16, pick(p, "sch_16"), &ValidationRule{Length: lenMinMax(1, 48)})
	b.setField(s, 17, pick(p, "sch_17"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 18, pick(p, "sch_18"), &ValidationRule{Length: lenMinMax(1, 106)})
	b.setField(s, 19, pick(p, "sch_19"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 20, pick(p, "sch_20"), &ValidationRule{Length: lenMinMax(1, 48)})
	b.setField(s, 21, pick(p, "sch_21"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.setField(s, 22, pick(p, "sch_22"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 23, pick(p, "sch_23"), &ValidationRule{Length: lenMinMax(1, 75)})
	b.setField(s, 24, pick(p, "sch_24"), &ValidationRule{Length: lenMinMax(1, 75)})
	b.setField(s, 25, pick(p, "sch_25"), &ValidationRule{Length: lenMinMax(1, 200)})
	return b
}

// BuildVAR builds a VAR (Variance) segment.
// Chainable.
func (b *Builder) BuildVAR(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.3")
	s := b.spec("VAR")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("VAR")
	b.setField(s, 1, pick(p, "var_1"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 2, b.dv(pick(p, "var_2"), ""), dateRule())
	b.setField(s, 3, b.dv(pick(p, "var_3"), ""), dateRule())
	b.setField(s, 4, pick(p, "var_4"), &ValidationRule{Length: lenMinMax(1, 80)})
	b.setField(s, 5, pick(p, "var_5"), &ValidationRule{Length: lenMinMax(1, 60)})
	b.setField(s, 6, pick(p, "var_6"), &ValidationRule{Length: lenMinMax(1, 512)})
	return b
}
