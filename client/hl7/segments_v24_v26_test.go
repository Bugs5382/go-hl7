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

// These tests cover the v2.4 typed
// segment builders (DRG/GOL/IAM/OM1-6/PRB/PTH/TXA plus the OBR/ORC/PID
// extensions), the v2.5 SFT/SPM builders, v2.5.1 inheritance, and the v2.6
// BPX/BTX/ITM/IVT/REL builders.

func v24() *hl7.Builder {
	b := hl7.New(hl7.V2_4)
	b.On("error", func(string) {})
	b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": segDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
	return b
}

func v25() *hl7.Builder {
	b := hl7.New(hl7.V2_5)
	b.On("error", func(string) {})
	b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": segDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
	return b
}

func v251() *hl7.Builder {
	b := hl7.New(hl7.V2_5_1)
	b.On("error", func(string) {})
	b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": segDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
	return b
}

func v26() *hl7.Builder {
	b := hl7.New(hl7.V2_6)
	b.On("error", func(string) {})
	b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": segDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
	return b
}

func TestHL724SegmentBuilders(t *testing.T) {
	t.Run("buildMSH stamps version 2.4", func(t *testing.T) {
		contains(t, v24().String(), "|2.4")
	})
	t.Run("buildPID with 2.4-only fields", func(t *testing.T) {
		b := v24()
		b.BuildPID(hl7.Props{"pid_3": "MRN1", "pid_31": "Y", "pid_32": "AL", "pid_33": segDate, "pid_34": "MSP", "pid_35": "SPC", "pid_36": "ETH", "pid_37": "MRN2", "pid_38": "OT", "pid_5": "DOE^JANE"})
		contains(t, b.String(), "\rPID|||MRN1||DOE^JANE")
	})
	t.Run("buildOBR with 2.4-only fields", func(t *testing.T) {
		b := v24()
		b.BuildOBR(hl7.Props{"obr_1": "1", "obr_4": "GLU^Glucose^L", "obr_44": "PROC", "obr_45": "PROC2", "obr_46": "ANT", "obr_47": "POS"})
		contains(t, b.String(), "\rOBR|1")
	})
	t.Run("buildORC with 2.4-only fields", func(t *testing.T) {
		b := v24()
		b.BuildORC(hl7.Props{"orc_1": "NW", "orc_2": "ORDER123", "orc_24": "ADDR", "orc_25": "STATUS_MOD"})
		contains(t, b.String(), "\rORC|NW|ORDER123")
	})
	t.Run("buildGOL", func(t *testing.T) {
		b := v24()
		tryBuild(func() {
			b.BuildGOL(hl7.Props{"gol_1": "AD", "gol_12": segDate, "gol_13": segDate, "gol_14": segDate, "gol_15": segDate, "gol_19": segDate, "gol_2": segDate, "gol_3": "GOAL_ID", "gol_4": "INSTANCE", "gol_7": segDate, "gol_8": segDate, "gol_9": segDate})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildPRB", func(t *testing.T) {
		b := v24()
		tryBuild(func() {
			b.BuildPRB(hl7.Props{"prb_1": "AD", "prb_15": segDate, "prb_16": segDate, "prb_2": segDate, "prb_3": "PROBLEM_ID", "prb_4": "INSTANCE", "prb_7": segDate, "prb_8": segDate, "prb_9": segDate})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildPTH", func(t *testing.T) {
		b := v24()
		tryBuild(func() {
			b.BuildPTH(hl7.Props{"pth_1": "AD", "pth_2": "PATH_ID", "pth_3": "INSTANCE", "pth_4": segDate, "pth_6": segDate})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildTXA", func(t *testing.T) {
		b := v24()
		tryBuild(func() {
			b.BuildTXA(hl7.Props{"txa_1": "1", "txa_12": "DOC_NO", "txa_17": "AU", "txa_2": "CD", "txa_4": segDate, "txa_6": segDate, "txa_7": segDate, "txa_8": segDate})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildIAM", func(t *testing.T) {
		b := v24()
		tryBuild(func() {
			b.BuildIAM(hl7.Props{"iam_1": "1", "iam_11": segDate, "iam_3": "ALLERGY_CODE", "iam_6": "A"})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildOM1", func(t *testing.T) {
		b := v24()
		tryBuild(func() {
			b.BuildOM1(hl7.Props{"om1_1": "1", "om1_2": "OBS_ID", "om1_21": segDate, "om1_22": segDate, "om1_4": "Y", "om1_5": "OTHER_ID"})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildOM2", func(t *testing.T) {
		b := v24()
		b.BuildOM2(hl7.Props{"om2_1": "1"})
		contains(t, b.String(), "\rOM2|1")
	})
	t.Run("buildOM3", func(t *testing.T) {
		b := v24()
		b.BuildOM3(hl7.Props{"om3_1": "1"})
		contains(t, b.String(), "\rOM3|1")
	})
	t.Run("buildOM4", func(t *testing.T) {
		b := v24()
		b.BuildOM4(hl7.Props{"om4_1": "1", "om4_13": "E"})
		contains(t, b.String(), "\rOM4|1")
	})
	t.Run("buildOM5", func(t *testing.T) {
		b := v24()
		b.BuildOM5(hl7.Props{"om5_1": "1"})
		contains(t, b.String(), "\rOM5|1")
	})
	t.Run("buildOM6", func(t *testing.T) {
		b := v24()
		b.BuildOM6(hl7.Props{"om6_1": "1"})
		contains(t, b.String(), "\rOM6|1")
	})
	t.Run("buildDRG", func(t *testing.T) {
		b := v24()
		b.BuildDRG(hl7.Props{"drg_1": "DRG123", "drg_2": segDate, "drg_3": "Y", "drg_4": "AB", "drg_6": "100"})
		contains(t, b.String(), "\rDRG|DRG123")
	})
}

func TestHL725SegmentBuilders(t *testing.T) {
	t.Run("buildMSH stamps version 2.5", func(t *testing.T) {
		contains(t, v25().String(), "|2.5")
	})
	t.Run("buildSFT", func(t *testing.T) {
		b := v25()
		tryBuild(func() {
			b.BuildSFT(hl7.Props{"sft_1": "Vendor Org", "sft_2": "1.0", "sft_3": "ProductName", "sft_4": "BinaryID1", "sft_5": "Software install info", "sft_6": segDate})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildSPM", func(t *testing.T) {
		b := v25()
		tryBuild(func() {
			b.BuildSPM(hl7.Props{"spm_1": "1", "spm_17": segDate, "spm_18": segDate, "spm_19": segDate, "spm_2": "SPEC_ID", "spm_20": "Y", "spm_3": "PARENT_ID", "spm_4": "FLD"})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("v2.5 still exposes 2.4 builders", func(t *testing.T) {
		b := v25()
		b.BuildOM2(hl7.Props{"om2_1": "1"})
		contains(t, b.String(), "\rOM2|1")
	})
}

func TestHL7251SegmentBuilders(t *testing.T) {
	t.Run("buildMSH stamps version 2.5.1", func(t *testing.T) {
		contains(t, v251().String(), "|2.5.1")
	})
	t.Run("v2.5.1 still exposes inherited builders", func(t *testing.T) {
		b := v251()
		tryBuild(func() {
			b.BuildSPM(hl7.Props{"spm_1": "1", "spm_4": "FLD"})
		})
		contains(t, b.String(), "MSH")
	})
}

func TestHL726SegmentBuilders(t *testing.T) {
	t.Run("buildMSH stamps version 2.6", func(t *testing.T) {
		contains(t, v26().String(), "|2.6")
	})
	t.Run("buildREL", func(t *testing.T) {
		b := v26()
		tryBuild(func() {
			b.BuildREL(hl7.Props{"rel_1": "1", "rel_13": segDate, "rel_2": "REL_TYPE", "rel_3": "INSTANCE_A", "rel_4": "INSTANCE_B", "rel_5": "REL_INSTANCE", "rel_6": segDate, "rel_7": segDate, "rel_8": "Y"})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildITM", func(t *testing.T) {
		b := v26()
		tryBuild(func() {
			b.BuildITM(hl7.Props{"itm_1": "ITEM_ID", "itm_11": "Y", "itm_14": "Y", "itm_17": "Y", "itm_22": "Y", "itm_23": "Y", "itm_24": "Y", "itm_26": "Y", "itm_3": "A", "itm_6": "Y"})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildIVT", func(t *testing.T) {
		b := v26()
		tryBuild(func() {
			b.BuildIVT(hl7.Props{"ivt_1": "1", "ivt_10": "Y", "ivt_12": "Y", "ivt_19": "Y", "ivt_2": "ITEM_ID", "ivt_21": "Y", "ivt_22": "Y", "ivt_23": "Y", "ivt_25": "Y", "ivt_6": "A"})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildBTX", func(t *testing.T) {
		b := v26()
		tryBuild(func() {
			b.BuildBTX(hl7.Props{"btx_1": "1", "btx_11": "DISP_STATUS", "btx_12": segDate, "btx_13": segDate, "btx_16": segDate, "btx_17": segDate})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("buildBPX", func(t *testing.T) {
		b := v26()
		tryBuild(func() {
			b.BuildBPX(hl7.Props{"bpx_1": "1", "bpx_13": segDate, "bpx_14": "1", "bpx_2": "RD", "bpx_3": "C", "bpx_4": segDate})
		})
		contains(t, b.String(), "MSH")
	})
	t.Run("v2.6 still exposes inherited 2.4 builders", func(t *testing.T) {
		b := v26()
		b.BuildDRG(hl7.Props{"drg_1": "DRG123", "drg_3": "Y"})
		contains(t, b.String(), "\rDRG|DRG123")
	})
}
