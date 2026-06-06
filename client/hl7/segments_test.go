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

package hl7_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/hl7"
)

// These tests mirror the hl7.segments.test.ts: the v2.1 typed segment
// builders, the v2.4 PID extension, the per-version MSH branches, v2.7 IPC/ISD,
// v2.8 STZ, and the HL7_BASE common helpers.

var segDate = time.Date(2024, 1, 15, 10, 20, 30, 0, time.UTC)

func v21() *hl7.HL7_BASE {
	b := hl7.NewHL7_2_1()
	b.On("error", func(string) {})
	b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11": "T", "msh_7": segDate, "msh_9": "ACK"})
	return b
}

func matches(t *testing.T, s, pattern string) {
	t.Helper()
	if !regexp.MustCompile(pattern).MatchString(s) {
		t.Fatalf("expected %q to match %q", s, pattern)
	}
}

func TestHL721SegmentBuilders(t *testing.T) {
	t.Run("buildEVN required type and timestamp", func(t *testing.T) {
		b := v21()
		b.BuildEVN(hl7.Props{"evn_1": "A01"})
		matches(t, b.String(), `\rEVN\|A01\|\d{14}\|\|`)
	})
	t.Run("buildEVN explicit dates", func(t *testing.T) {
		b := v21()
		b.BuildEVN(hl7.Props{"evn_1": "A01", "evn_2": segDate, "evn_3": segDate})
		contains(t, b.String(), "\rEVN|A01|20240115")
	})
	t.Run("buildMSA", func(t *testing.T) {
		b := v21()
		b.BuildMSA(hl7.Props{"msa_1": "AA", "msa_2": "ORIG_ID"})
		contains(t, b.String(), "\rMSA|AA|ORIG_ID")
	})
	t.Run("buildPID required pid_3/pid_5", func(t *testing.T) {
		b := v21()
		b.BuildPID(hl7.Props{"pid_3": "MRN1", "pid_5": "DOE^JANE"})
		contains(t, b.String(), "\rPID|||MRN1||DOE^JANE")
	})
	t.Run("buildPID date of birth", func(t *testing.T) {
		b := v21()
		b.BuildPID(hl7.Props{"pid_3": "MRN1", "pid_5": "DOE^JANE", "pid_7": time.Date(1980, 6, 15, 12, 0, 0, 0, time.UTC)})
		matches(t, b.String(), `\rPID\|\|\|MRN1\|\|DOE\^JANE\|\|198006\d{2}`)
	})
	t.Run("buildPV1 required patient class", func(t *testing.T) {
		b := v21()
		b.BuildPV1(hl7.Props{"pv1_2": "I", "pv1_3": "WARD-A"})
		contains(t, b.String(), "\rPV1||I|WARD-A")
	})
	t.Run("buildNK1", func(t *testing.T) {
		b := v21()
		b.BuildNK1(hl7.Props{"nk1_1": "1"})
		contains(t, b.String(), "\rNK1|1|")
	})
	t.Run("buildDG1", func(t *testing.T) {
		b := v21()
		b.BuildDG1(hl7.Props{"dg1_1": "1", "dg1_2": "I9", "dg1_3": "401.9", "dg1_4": "HTN", "dg1_6": "A"})
		contains(t, b.String(), "\rDG1|1|I9|401.9|HTN")
	})
	t.Run("buildACC", func(t *testing.T) {
		b := v21()
		b.BuildACC(hl7.Props{"acc_1": segDate, "acc_2": "AA"})
		contains(t, b.String(), "\rACC|20240115")
	})
	t.Run("buildBLG", func(t *testing.T) {
		b := v21()
		b.BuildBLG(hl7.Props{"blg_1": "D", "blg_2": "CR", "blg_3": "ACCT123"})
		contains(t, b.String(), "\rBLG|D|CR|ACCT123")
	})
	t.Run("buildDSC", func(t *testing.T) {
		b := v21()
		b.BuildDSC(hl7.Props{"dsc_1": "PTR"})
		contains(t, b.String(), "\rDSC|PTR")
	})
	t.Run("buildERR", func(t *testing.T) {
		b := v21()
		b.BuildERR(hl7.Props{"err_1": "0^Required field missing^HL70357"})
		contains(t, b.String(), "\rERR|0^Required field missing^HL70357")
	})
	t.Run("buildMRG", func(t *testing.T) {
		b := v21()
		b.BuildMRG(hl7.Props{"mrg_1": "MRN_OLD"})
		contains(t, b.String(), "\rMRG|MRN_OLD")
	})
	t.Run("buildNTE", func(t *testing.T) {
		b := v21()
		b.BuildNTE(hl7.Props{"nte_1": "1", "nte_2": "L", "nte_3": "Some clinical note"})
		contains(t, b.String(), "\rNTE|1|L|Some clinical note")
	})
	t.Run("buildORC", func(t *testing.T) {
		b := v21()
		b.BuildORC(hl7.Props{"orc_1": "NW", "orc_2": "ORDER123"})
		contains(t, b.String(), "\rORC|NW|ORDER123")
	})
	t.Run("buildOBR", func(t *testing.T) {
		b := v21()
		b.BuildOBR(hl7.Props{"obr_1": "1", "obr_14": "20240115093000", "obr_22": "20240115110500", "obr_4": "GLU^Glucose^L", "obr_7": "20240115102030", "obr_8": "20240115110000", "obr_9": "10^mL"})
		contains(t, b.String(), "\rOBR|1")
	})
	t.Run("buildOBX", func(t *testing.T) {
		b := v21()
		// OBX.2 value type is HL7 table 0125, which is version-aware: the numeric
		// type "NM" is not part of the v2.1 value set, so this v2.1 fixture uses
		// "ST" (string), a value the v2.1 table does carry.
		b.BuildOBX(hl7.Props{"obx_1": "1", "obx_11": "F", "obx_2": "ST", "obx_3": "GLU^Glucose^L", "obx_5": "98"})
		contains(t, b.String(), "\rOBX|1|ST|GLU^Glucose^L||98")
	})
	t.Run("buildFT1", func(t *testing.T) {
		b := v21()
		b.BuildFT1(hl7.Props{"ft1_4": segDate, "ft1_6": "CG", "ft1_7": "BEDS^Bed charge^L"})
		contains(t, b.String(), "\rFT1|")
	})
	t.Run("buildQRF", func(t *testing.T) {
		b := v21()
		b.BuildQRF(hl7.Props{"qrf_1": "FAC1"})
		contains(t, b.String(), "\rQRF|")
	})
	t.Run("buildURD", func(t *testing.T) {
		b := v21()
		b.BuildURD(hl7.Props{"urd_1": segDate, "urd_3": "USER1"})
		contains(t, b.String(), "\rURD|")
	})
	t.Run("buildURS", func(t *testing.T) {
		b := v21()
		b.BuildURS(hl7.Props{"urs_1": "RES1"})
		contains(t, b.String(), "\rURS|RES1")
	})
	t.Run("buildIN1", func(t *testing.T) {
		b := v21()
		b.BuildIN1(hl7.Props{"in1_1": "1", "in1_2": "PLAN1", "in1_3": "COMP1", "in1_8": "GROUP1"})
		contains(t, b.String(), "\rIN1|1|PLAN1")
	})
	t.Run("buildUB1", func(t *testing.T) {
		b := v21()
		b.BuildUB1(hl7.Props{"ub1_1": "1"})
		contains(t, b.String(), "\rUB1|1")
	})
	t.Run("buildNPU", func(t *testing.T) {
		b := v21()
		b.BuildNPU(hl7.Props{"npu_1": "BED1"})
		contains(t, b.String(), "\rNPU|BED1")
	})
	t.Run("buildNSC", func(t *testing.T) {
		b := v21()
		b.BuildNSC(hl7.Props{"nsc_1": "M"})
		contains(t, b.String(), "\rNSC|")
	})
	t.Run("buildNCK", func(t *testing.T) {
		b := v21()
		b.BuildNCK()
		contains(t, b.String(), "\rNCK")
	})
	t.Run("buildSTZ on 2.1 throws HL7FatalError", func(t *testing.T) {
		b := v21()
		defer func() {
			r := recover()
			if r == nil {
				t.Fatal("expected panic")
			}
			if err, ok := r.(error); !ok || !errors.Is(err, helpers.ErrFatal) {
				t.Fatalf("expected HL7FatalError, got %v", r)
			}
		}()
		b.BuildSTZ(hl7.Props{"stz_1": "x"})
	})
	t.Run("headerExists guards before buildMSH", func(t *testing.T) {
		fresh := hl7.NewHL7_2_1()
		expectThrows(t, "MSH Header", func() { fresh.BuildEVN(hl7.Props{"evn_1": "A01"}) })
	})
}

func TestHL724PIDExtension(t *testing.T) {
	b := hl7.NewHL7_2_4()
	b.On("error", func(string) {})
	b.BuildMSH(hl7.Props{"msh_10": "X", "msh_11_1": "P", "msh_7": segDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
	b.BuildPID(hl7.Props{"pid_3": "MRN1", "pid_31": "Y", "pid_32": "AL", "pid_33": segDate, "pid_37": "MRN2", "pid_5": "DOE^JANE"})
	contains(t, b.String(), "|Y|AL|")
}

func TestVersionMSHBranches(t *testing.T) {
	cases := []struct {
		newB func(...hl7.Options) *hl7.HL7_BASE
		want string
	}{
		{hl7.NewHL7_2_5, "|2.5"},
		{hl7.NewHL7_2_5_1, "|2.5.1"},
		{hl7.NewHL7_2_6, "|2.6"},
		{hl7.NewHL7_2_7, "|2.7"},
		{hl7.NewHL7_2_7_1, "|2.7.1"},
		{hl7.NewHL7_2_8, "|2.8"},
	}
	for _, c := range cases {
		b := c.newB()
		b.On("error", func(string) {})
		b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": segDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
		contains(t, b.String(), c.want)

		b2 := c.newB()
		b2.On("error", func(string) {})
		b2.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_11_2": "A", "msh_7": segDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
		contains(t, b2.String(), "|P^A|")
	}
}

func TestHL727IPCISD(t *testing.T) {
	mk := func() *hl7.HL7_BASE {
		b := hl7.NewHL7_2_7()
		b.On("error", func(string) {})
		b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": segDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
		return b
	}
	t.Run("buildIPC", func(t *testing.T) {
		b := mk()
		b.BuildIPC(hl7.Props{"ipc_1": "ACC123", "ipc_2": "REQ123", "ipc_3": "STUDY1", "ipc_4": "SCHED1"})
		contains(t, b.String(), "\rIPC|ACC123|REQ123|STUDY1|SCHED1")
	})
	t.Run("buildISD", func(t *testing.T) {
		b := mk()
		b.BuildISD(hl7.Props{"isd_1": "1", "isd_3": "OK"})
		contains(t, b.String(), "\rISD|1")
	})
}

func TestHL728STZ(t *testing.T) {
	b := hl7.NewHL7_2_8()
	b.On("error", func(string) {})
	b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": segDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
	b.BuildSTZ(hl7.Props{"stz_1": "STM", "stz_2": "DRT", "stz_3": "30MIN"})
	contains(t, b.String(), "\rSTZ|STM|DRT|30MIN")
}

func TestBaseCommonHelpers(t *testing.T) {
	t.Run("setDate uses now without args", func(t *testing.T) {
		b := hl7.NewHL7_2_7()
		matches(t, b.SetDate(time.Time{}, "14"), `^\d{14}$`)
	})
	t.Run("setDate honors length option", func(t *testing.T) {
		b := hl7.NewHL7_2_7()
		if got := b.SetDate(segDate, "8"); got != "20240115" {
			t.Fatalf("got %q", got)
		}
	})
	t.Run("toMessage returns the underlying message", func(t *testing.T) {
		b := v21()
		if b.ToMessage().String() != b.String() {
			t.Fatal("mismatch")
		}
	})
	t.Run("checkMSH on base 2.1 throws Not Implemented", func(t *testing.T) {
		b := hl7.NewHL7_2_1()
		expectThrows(t, "Not Implemented", func() { b.CheckMSH(hl7.Props{}) })
	})
	t.Run("checkMSH on 2.8 delegates to 2.7 checks", func(t *testing.T) {
		b := hl7.NewHL7_2_8()
		if !b.CheckMSH(hl7.Props{"msh_11_1": "P", "msh_9_1": "ADT", "msh_9_2": "A01"}) {
			t.Fatal("expected true")
		}
	})
}
