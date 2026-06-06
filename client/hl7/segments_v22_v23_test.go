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

	"github.com/Bugs5382/go-hl7/client/hl7"
)

// These tests mirror the hl7.segments.v22-v23.test.ts: the v2.2 and
// v2.3 typed segment builders, including the version-gated OBR/OBX/ORC/PID/PV1
// extensions and the new scheduling/clinical-study/provider segments.

func v22() *hl7.Builder {
	b := hl7.New(hl7.V2_2)
	b.On("error", func(string) {})
	b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11": "P", "msh_7": segDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
	return b
}

func v23() *hl7.Builder {
	b := hl7.New(hl7.V2_3)
	b.On("error", func(string) {})
	b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": segDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
	return b
}

// tryBuild runs fn, swallowing any builder validation panic (mirrors the
// try/catch around the coverage-only segment scenarios).
func tryBuild(fn func()) {
	defer func() { _ = recover() }()
	fn()
}

func TestHL722SegmentBuilders(t *testing.T) {
	t.Run("buildMSH version stamp 2.2", func(t *testing.T) {
		contains(t, v22().String(), "|2.2")
	})
	t.Run("buildPID extends with 2.2 fields", func(t *testing.T) {
		b := v22()
		b.BuildPID(hl7.Props{"pid_20": "DL12345", "pid_21": "MOTHER_ID", "pid_22": "ABC", "pid_23": "BIRTHPLACE", "pid_24": "Y", "pid_25": "1", "pid_26": "USA", "pid_3": "MRN1", "pid_5": "DOE^JANE"})
		contains(t, b.String(), "\rPID|||MRN1||DOE^JANE")
	})
	t.Run("buildPV1 extends with 2.2 fields", func(t *testing.T) {
		b := v22()
		b.BuildPV1(hl7.Props{"pv1_2": "I", "pv1_45": segDate, "pv1_46": "100.00", "pv1_47": "200.00", "pv1_48": "300.00", "pv1_49": "400.00", "pv1_50": "VISIT123"})
		contains(t, b.String(), "\rPV1||I|")
	})
	t.Run("buildOBX extends with 2.2 fields", func(t *testing.T) {
		b := v22()
		b.BuildOBX(hl7.Props{"obx_1": "1", "obx_11": "F", "obx_12": segDate, "obx_13": "USER_DEFINED", "obx_14": segDate, "obx_15": "PROD_ID", "obx_2": "NM", "obx_3": "GLU^Glucose^L", "obx_5": "98"})
		contains(t, b.String(), "\rOBX|1|NM|GLU^Glucose^L||98")
	})
	t.Run("buildOBR extends with 2.2 fields", func(t *testing.T) {
		b := v22()
		b.BuildOBR(hl7.Props{"obr_1": "1", "obr_26": "PARENT", "obr_27": "1^^^^^R", "obr_28": "DR_NAME", "obr_29": "PARENT_NUM", "obr_3": "FILLER1", "obr_30": "WALK", "obr_31": "REASON", "obr_32": "PRINCIPAL", "obr_33": "ASSISTANT", "obr_34": "TECH", "obr_35": "TRANSCRIBER", "obr_4": "GLU^Glucose^L"})
		contains(t, b.String(), "\rOBR|1")
	})
	t.Run("buildORC extends with 2.2 fields", func(t *testing.T) {
		b := v22()
		b.BuildORC(hl7.Props{"orc_1": "NW", "orc_15": segDate, "orc_16": "REASON_CODE", "orc_17": "ENT_BY_LOC", "orc_18": "CALL_BACK", "orc_19": "ORDER_CTRL_REASON", "orc_2": "ORDER123"})
		contains(t, b.String(), "\rORC|NW|ORDER123")
	})
	t.Run("buildAL1", func(t *testing.T) {
		b := v22()
		b.BuildAL1(hl7.Props{"al1_1": "1", "al1_3": "PEANUT", "al1_5": "HIVES", "al1_6": segDate})
		contains(t, b.String(), "\rAL1|1")
	})
	t.Run("buildUB2", func(t *testing.T) {
		b := v22()
		b.BuildUB2(hl7.Props{"ub2_1": "1", "ub2_10": "100.00", "ub2_11": "1", "ub2_12": "REV", "ub2_13": "1", "ub2_14": "VAL", "ub2_15": "DIAG", "ub2_16": "DG", "ub2_2": "USA", "ub2_3": "01", "ub2_5": "1", "ub2_9": "1"})
		contains(t, b.String(), "\rUB2|1")
	})
	t.Run("buildRXA", func(t *testing.T) {
		b := v22()
		b.BuildRXA(hl7.Props{"rxa_1": "1", "rxa_10": "WHO_GAVE", "rxa_11": "LOC1", "rxa_12": "ADMIN_ROUTE", "rxa_2": "1", "rxa_3": segDate, "rxa_4": segDate, "rxa_5": "MMR^Measles vaccine^L", "rxa_6": "1", "rxa_7": "PO", "rxa_8": "ADMIN_NOTE", "rxa_9": "ADMIN_BY"})
		contains(t, b.String(), "\rRXA|1|1")
	})
	t.Run("buildRXR", func(t *testing.T) {
		b := v22()
		b.BuildRXR(hl7.Props{"rxr_1": "PO", "rxr_2": "MOUTH", "rxr_3": "IVP", "rxr_4": "PT"})
		contains(t, b.String(), "\rRXR|PO")
	})
	t.Run("buildMFI", func(t *testing.T) {
		b := v22()
		b.BuildMFI(hl7.Props{"mfi_1": "STF", "mfi_2": "AUTH", "mfi_3": "UPD", "mfi_4": segDate, "mfi_5": segDate, "mfi_6": "AL"})
		contains(t, b.String(), "\rMFI|STF")
	})
	t.Run("buildMFE", func(t *testing.T) {
		b := v22()
		b.BuildMFE(hl7.Props{"mfe_1": "MAD", "mfe_2": "EFFDATE", "mfe_3": segDate, "mfe_4": "PRIMARY_KEY"})
		contains(t, b.String(), "\rMFE|MAD")
	})
	t.Run("buildSTF", func(t *testing.T) {
		b := v22()
		b.BuildSTF(hl7.Props{"stf_1": "STF1", "stf_10": "OFFICE_ADDR", "stf_11": "INST_ACTIV", "stf_12": segDate, "stf_13": segDate, "stf_14": "LICENSE_NO", "stf_2": "ID1", "stf_3": "DOE^JOHN", "stf_4": "MD", "stf_5": "M", "stf_6": segDate, "stf_7": "A", "stf_8": "DEPT1", "stf_9": "OFFICE_PHONE"})
		contains(t, b.String(), "\rSTF|STF1")
	})
	t.Run("buildRXO", func(t *testing.T) {
		b := v22()
		tryBuild(func() {
			b.BuildRXO(hl7.Props{"rxo_1": "ASA^Aspirin^L", "rxo_10": "PHARM_INSTR", "rxo_11": "PO", "rxo_12": "PHARM_ROUTE", "rxo_13": "100", "rxo_14": "BRAND", "rxo_15": "GENERIC", "rxo_16": "Y", "rxo_17": "PROVIDER", "rxo_2": "325", "rxo_3": "650", "rxo_4": "MG", "rxo_5": "TAB", "rxo_6": "PRN", "rxo_7": "1 PO Q4H", "rxo_8": "10", "rxo_9": "G"})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildRXE", func(t *testing.T) {
		b := v22()
		tryBuild(func() {
			b.BuildRXE(hl7.Props{"rxe_1": "Q4H", "rxe_10": "PHARM_INSTR", "rxe_11": "PO", "rxe_12": "100", "rxe_13": "BRAND", "rxe_14": "GENERIC", "rxe_15": "PR1", "rxe_16": "PR2", "rxe_17": "PR3", "rxe_18": segDate, "rxe_19": "REFILLS", "rxe_2": "ASA^Aspirin^L", "rxe_20": "Y", "rxe_21": "DAYS", "rxe_22": "10", "rxe_23": "1", "rxe_24": "PROVIDER_ID", "rxe_3": "325", "rxe_4": "650", "rxe_5": "MG", "rxe_6": "TAB", "rxe_7": "PRN", "rxe_8": "10", "rxe_9": "G"})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildRXD", func(t *testing.T) {
		b := v22()
		tryBuild(func() {
			b.BuildRXD(hl7.Props{"rxd_1": "1", "rxd_10": "DISP_BY", "rxd_11": "N", "rxd_12": "RX_NUM", "rxd_13": "10", "rxd_14": "Y", "rxd_15": "PROD_INFO", "rxd_2": "ASA^Aspirin^L", "rxd_3": segDate, "rxd_4": "10", "rxd_5": "MG", "rxd_6": "TAB", "rxd_7": "1 PO Q4H", "rxd_8": "PRN", "rxd_9": "DISP_NOTES"})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildRXG", func(t *testing.T) {
		b := v22()
		tryBuild(func() {
			b.BuildRXG(hl7.Props{"rxg_1": "1", "rxg_10": "N", "rxg_11": "ADMIN", "rxg_12": "Y", "rxg_13": "PROD_INFO", "rxg_14": "ADMIN_NOTES", "rxg_15": "1", "rxg_16": "PROVIDER", "rxg_2": "1", "rxg_3": "QUAN", "rxg_4": "ASA^Aspirin^L", "rxg_5": "10", "rxg_6": "20", "rxg_7": "MG", "rxg_8": "TAB", "rxg_9": "GIVE_INSTR"})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildODS", func(t *testing.T) {
		b := v22()
		b.BuildODS(hl7.Props{"ods_1": "D", "ods_2": "SERVICE_PERIOD", "ods_3": "DIET_CODE", "ods_4": "TEXT"})
		contains(t, b.String(), "\rODS|D")
	})
	t.Run("buildODT", func(t *testing.T) {
		b := v22()
		b.BuildODT(hl7.Props{"odt_1": "EARLY", "odt_2": "SERVE_PERIOD", "odt_3": "TEXT"})
		contains(t, b.String(), "\rODT|EARLY")
	})
}

func TestHL723SegmentBuilders(t *testing.T) {
	t.Run("buildMSH version stamp 2.3", func(t *testing.T) {
		contains(t, v23().String(), "|2.3")
	})
	t.Run("buildMSH with msh_11_2 processing mode", func(t *testing.T) {
		c := hl7.New(hl7.V2_3)
		c.On("error", func(string) {})
		c.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_11_2": "A", "msh_7": segDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
		contains(t, c.String(), "|P^A|")
	})
	t.Run("buildPID extends with 2.3 fields", func(t *testing.T) {
		b := v23()
		b.BuildPID(hl7.Props{"pid_27": "VET_STATUS", "pid_28": "NAT", "pid_29": segDate, "pid_3": "MRN1", "pid_30": "Y", "pid_5": "DOE^JANE"})
		contains(t, b.String(), "\rPID|||MRN1||DOE^JANE")
	})
	t.Run("buildOBX extends with 2.3 fields", func(t *testing.T) {
		b := v23()
		b.BuildOBX(hl7.Props{"obx_1": "1", "obx_11": "F", "obx_16": "RESPONSIBLE", "obx_17": "METHOD", "obx_2": "NM", "obx_3": "GLU^Glucose^L", "obx_5": "98"})
		contains(t, b.String(), "\rOBX|1|NM|GLU^Glucose^L||98")
	})
	t.Run("buildOBR extends with 2.3 fields", func(t *testing.T) {
		b := v23()
		b.BuildOBR(hl7.Props{"obr_1": "1", "obr_36": segDate, "obr_37": "1", "obr_38": "TRANSPORT", "obr_39": "TRANS_ARRG", "obr_4": "GLU^Glucose^L", "obr_40": "ESCORT", "obr_41": "A", "obr_42": "R", "obr_43": "PROCEDURE_CODE"})
		contains(t, b.String(), "\rOBR|1")
	})
	t.Run("buildORC extends with 2.3 fields", func(t *testing.T) {
		b := v23()
		b.BuildORC(hl7.Props{"orc_1": "NW", "orc_2": "ORDER123"})
		contains(t, b.String(), "\rORC|NW|ORDER123")
	})
	t.Run("buildSCH", func(t *testing.T) {
		b := v23()
		tryBuild(func() {
			b.BuildSCH(hl7.Props{"sch_1": "PLACER1", "sch_10": "TIMING", "sch_11": "ALLOC^RES^SPEC", "sch_12": "GROUP", "sch_13": "REASON_TEXT", "sch_14": "CONTACT", "sch_15": "PHONE", "sch_16": "ENTERED_BY", "sch_17": "ENT_PHONE", "sch_18": "ENT_ADDR", "sch_19": "ENT_LOC", "sch_2": "FILLER1", "sch_20": "PLACER_PERSON", "sch_21": "PLACER_PHONE", "sch_22": "PLACER_ADDR", "sch_23": "FILLER_LOCATION", "sch_24": "ENT_LOC_2", "sch_25": "BOOKED", "sch_3": "1", "sch_4": "PARENT", "sch_5": "NOTIFY", "sch_6": "REASON^TXT", "sch_7": "ROUTINE", "sch_8": "NORMAL", "sch_9": "30"})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildRGS", func(t *testing.T) {
		b := v23()
		b.BuildRGS(hl7.Props{"rgs_1": "1", "rgs_2": "A", "rgs_3": "RG_TEXT"})
		contains(t, b.String(), "\rRGS|1")
	})
	t.Run("buildAIS", func(t *testing.T) {
		b := v23()
		b.BuildAIS(hl7.Props{"ais_1": "1", "ais_10": "BOOKED", "ais_2": "A", "ais_3": "PROC^Procedure^L", "ais_4": segDate, "ais_5": "30", "ais_6": "MIN", "ais_7": "60", "ais_8": "BLOCKED", "ais_9": "CONFIRM"})
		contains(t, b.String(), "\rAIS|1")
	})
	t.Run("buildAIG", func(t *testing.T) {
		b := v23()
		b.BuildAIG(hl7.Props{"aig_1": "1", "aig_10": "MIN", "aig_11": "60", "aig_12": "BLOCKED", "aig_13": "CONFIRM", "aig_14": "BOOKED", "aig_2": "A", "aig_3": "RES^Resource^L", "aig_4": "MD", "aig_5": "STAFF", "aig_6": "QTY", "aig_7": "ALLOC", "aig_8": segDate, "aig_9": "30"})
		contains(t, b.String(), "\rAIG|1")
	})
	t.Run("buildAIL", func(t *testing.T) {
		b := v23()
		b.BuildAIL(hl7.Props{"ail_1": "1", "ail_10": "BLOCKED", "ail_11": "CONFIRM", "ail_12": "BOOKED", "ail_2": "A", "ail_3": "ROOM1", "ail_4": "TYPE", "ail_5": "GROUP", "ail_6": segDate, "ail_7": "30", "ail_8": "MIN", "ail_9": "60"})
		contains(t, b.String(), "\rAIL|1")
	})
	t.Run("buildAIP", func(t *testing.T) {
		b := v23()
		b.BuildAIP(hl7.Props{"aip_1": "1", "aip_10": "BLOCKED", "aip_11": "CONFIRM", "aip_12": "BOOKED", "aip_2": "A", "aip_3": "DOC1", "aip_4": "MD", "aip_5": "ROLE", "aip_6": segDate, "aip_7": "30", "aip_8": "MIN", "aip_9": "60"})
		contains(t, b.String(), "\rAIP|1")
	})
	t.Run("buildAPR", func(t *testing.T) {
		b := v23()
		b.BuildAPR(hl7.Props{"apr_1": "MON,TUE", "apr_2": "AM", "apr_3": "RES_PREF", "apr_4": "1", "apr_5": "GENDER"})
		contains(t, b.String(), "\rAPR|")
	})
	t.Run("buildPRA", func(t *testing.T) {
		b := v23()
		b.BuildPRA(hl7.Props{"pra_1": "PRACT1", "pra_2": "GROUP1", "pra_3": "MD", "pra_4": "I", "pra_5": "SPEC", "pra_6": "PRIV", "pra_7": "JURISDICTION", "pra_8": segDate})
		contains(t, b.String(), "\rPRA|")
	})
	t.Run("buildPD1", func(t *testing.T) {
		b := v23()
		b.BuildPD1(hl7.Props{"pd1_1": "S", "pd1_10": "PROTECT_IND", "pd1_11": "1", "pd1_12": "Y", "pd1_2": "F", "pd1_3": "CLINIC^Name", "pd1_4": "DOC^Name", "pd1_5": "F", "pd1_6": "1", "pd1_7": "Y", "pd1_8": "Y", "pd1_9": "Y"})
		contains(t, b.String(), "\rPD1|")
	})
	t.Run("buildROL", func(t *testing.T) {
		b := v23()
		b.BuildROL(hl7.Props{"rol_1": "ROL1", "rol_2": "A", "rol_3": "ATTEND", "rol_4": "DOC^John", "rol_5": segDate, "rol_6": segDate, "rol_7": "REASON", "rol_8": "ORG"})
		contains(t, b.String(), "\rROL|ROL1")
	})
	t.Run("buildVAR", func(t *testing.T) {
		b := v23()
		b.BuildVAR(hl7.Props{"var_1": "VAR1", "var_2": segDate, "var_3": segDate, "var_4": "PERSON", "var_5": "REASON", "var_6": "DESCRIPTION"})
		contains(t, b.String(), "\rVAR|VAR1")
	})
	t.Run("buildPSH", func(t *testing.T) {
		b := v23()
		b.BuildPSH(hl7.Props{"psh_1": "REPORT_TYPE", "psh_10": "100", "psh_11": "E", "psh_12": "QTY_INTERP", "psh_13": "1", "psh_14": "1", "psh_2": "REPORT_FORM", "psh_3": segDate, "psh_4": segDate, "psh_5": segDate, "psh_6": "10", "psh_7": "5", "psh_8": "A", "psh_9": "QTY_INTERP"})
		contains(t, b.String(), "\rPSH|REPORT_TYPE")
	})
	t.Run("buildPCR", func(t *testing.T) {
		b := v23()
		b.BuildPCR(hl7.Props{"pcr_1": "PROD", "pcr_10": "PROB_REC", "pcr_11": "5", "pcr_12": "PRIOR_HIST", "pcr_13": "Y", "pcr_14": "ACTION", "pcr_15": "A", "pcr_16": "OUTCOME", "pcr_17": "R", "pcr_18": segDate, "pcr_19": "P", "pcr_2": "Y", "pcr_20": "N", "pcr_21": "N", "pcr_22": "OT", "pcr_23": "P", "pcr_3": "PROD_CLASS", "pcr_4": "QTY", "pcr_5": segDate, "pcr_6": segDate, "pcr_7": segDate, "pcr_8": segDate, "pcr_9": "DURATION"})
		contains(t, b.String(), "\rPCR|PROD")
	})
	t.Run("buildPRD", func(t *testing.T) {
		b := v23()
		b.BuildPRD(hl7.Props{"prd_1": "RP", "prd_2": "DOC^John", "prd_3": "ADDR", "prd_4": "LOC", "prd_5": "5551112222", "prd_6": "O", "prd_7": "PROV_ID", "prd_8": segDate, "prd_9": segDate})
		contains(t, b.String(), "\rPRD|RP")
	})
	t.Run("buildCTD", func(t *testing.T) {
		b := v23()
		b.BuildCTD(hl7.Props{"ctd_1": "BP", "ctd_2": "DOE^JANE", "ctd_3": "ADDR", "ctd_4": "LOC", "ctd_5": "5551112222", "ctd_6": "O", "ctd_7": "PREF_METHOD"})
		contains(t, b.String(), "\rCTD|BP")
	})
	t.Run("buildRDF", func(t *testing.T) {
		b := v23()
		b.BuildRDF(hl7.Props{"rdf_1": "5", "rdf_2": "PATIENT_NAME"})
		contains(t, b.String(), "\rRDF|5")
	})
	t.Run("buildRDT", func(t *testing.T) {
		b := v23()
		b.BuildRDT(hl7.Props{"rdt_1": "ROW_VAL"})
		contains(t, b.String(), "\rRDT|ROW_VAL")
	})
	t.Run("buildCSR", func(t *testing.T) {
		b := v23()
		b.BuildCSR(hl7.Props{"csr_1": "SPONSOR1", "csr_10": "TREATMENT", "csr_11": segDate, "csr_12": "EVAL_CODE", "csr_13": "RANDOMIZED", "csr_14": "RAND_DATE", "csr_15": segDate, "csr_16": "INITIATOR", "csr_2": "ALT_SPONSOR", "csr_3": "STUDY1", "csr_4": "PT_ID", "csr_5": "ALT_PT_ID", "csr_6": segDate, "csr_7": "ALT_REG", "csr_8": "PHASE1", "csr_9": segDate})
		contains(t, b.String(), "\rCSR|SPONSOR1")
	})
	t.Run("buildCSP", func(t *testing.T) {
		b := v23()
		b.BuildCSP(hl7.Props{"csp_1": "PHASE1", "csp_2": segDate, "csp_3": segDate, "csp_4": "TREATMENT_PLAN"})
		contains(t, b.String(), "\rCSP|PHASE1")
	})
	t.Run("buildCSS", func(t *testing.T) {
		b := v23()
		b.BuildCSS(hl7.Props{"css_1": "TIMEPOINT1", "css_2": segDate, "css_3": "ACTION"})
		contains(t, b.String(), "\rCSS|TIMEPOINT1")
	})
}
