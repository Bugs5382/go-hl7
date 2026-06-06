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
	"strings"
	"testing"
	"time"

	"github.com/Bugs5382/go-hl7/client/hl7"
	"github.com/Bugs5382/go-hl7/client/utils"
)

// These tests mirror the hl7.build.test.ts "builder message - all
// versions" block: per-version BuildMSH wire format and CheckMSH behavior, plus
// the v2.1 BuildEVN/BuildFT1/BuildNCK cases. Batch/FileBatch coverage from that
// file is out of the spec-builder scope.

func contains(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected %q to contain %q", s, sub)
	}
}

func expectThrows(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic containing %q, got none", want)
		}
		err, ok := r.(error)
		if !ok {
			t.Fatalf("expected error panic, got %T (%v)", r, r)
		}
		if want != "" && !strings.Contains(err.Error(), want) {
			t.Fatalf("expected message containing %q, got %q", want, err.Error())
		}
	}()
	fn()
}

func TestBuildAllVersions(t *testing.T) {
	useThisDate := time.Now()

	t.Run("2.1", func(t *testing.T) {
		base := func() *hl7.Builder {
			b := hl7.New(hl7.V2_1)
			b.BuildMSH(hl7.Props{"msh_10": "12345", "msh_11": "T", "msh_7": useThisDate, "msh_9": "ACK"})
			return b
		}
		baseResult := `MSH|^~\&|||||` + utils.CreateHL7Date(useThisDate, "14") + `||ACK|12345|T|2.1`

		t.Run("buildMSH produces the 2.1 base wire format", func(t *testing.T) {
			if got := base().String(); got != baseResult {
				t.Fatalf("got %q want %q", got, baseResult)
			}
		})

		t.Run("buildEVN appends an EVN segment", func(t *testing.T) {
			b := base()
			b.BuildEVN(hl7.Props{"evn_1": "A01", "evn_2": useThisDate})
			want := baseResult + "\rEVN|A01|" + utils.CreateHL7Date(useThisDate, "14") + "||"
			if got := b.String(); got != want {
				t.Fatalf("got %q want %q", got, want)
			}
		})

		t.Run("buildFT1 appends a Financial Transaction segment", func(t *testing.T) {
			b := base()
			b.BuildFT1(hl7.Props{"ft1_4": useThisDate, "ft1_6": "ADD", "ft1_7": "HELLO"})
			want := baseResult + "\rFT1||||" + utils.CreateHL7Date(useThisDate, "8") + "||ADD|HELLO|||||||||||||||"
			if got := b.String(); got != want {
				t.Fatalf("got %q want %q", got, want)
			}
		})

		t.Run("buildNCK appends a System Clock segment", func(t *testing.T) {
			b := base()
			b.BuildNCK()
			contains(t, b.String(), baseResult)
			contains(t, b.String(), "\rNCK|")
		})
	})

	t.Run("2.2", func(t *testing.T) {
		newB := func() *hl7.Builder { return hl7.New(hl7.V2_2) }

		t.Run("buildMSH produces a 2.2 base header", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11": "T", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
			contains(t, b.String(), `MSH|^~\&`)
			contains(t, b.String(), `|ADT^A01|CONTROL_ID|T|2.2`)
		})
		t.Run("buildMSH carries optional sending/receiving fields", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "MSG001", "msh_11": "P", "msh_3": "SENDAPP", "msh_4": "SENDFAC", "msh_5": "RECVAPP", "msh_6": "RECVFAC", "msh_7": useThisDate, "msh_9_1": "ORM", "msh_9_2": "O01"})
			contains(t, b.String(), "|SENDAPP|SENDFAC|RECVAPP|RECVFAC|")
			contains(t, b.String(), "|ORM^O01|MSG001|P|2.2")
		})
		t.Run("checkMSH accepts a valid 2.2 header", func(t *testing.T) {
			if !newB().CheckMSH(hl7.Props{"msh_10": "MSG001", "msh_11": "P", "msh_9_1": "ADT", "msh_9_2": "A01"}) {
				t.Fatal("expected true")
			}
		})
		t.Run("checkMSH rejects msh_9_1 wrong length", func(t *testing.T) {
			expectThrows(t, "MSH.9.1 must be 3 characters in length.", func() {
				newB().CheckMSH(hl7.Props{"msh_9_1": "ADTY", "msh_9_2": "A01"})
			})
		})
		t.Run("checkMSH rejects msh_9_2 wrong length", func(t *testing.T) {
			expectThrows(t, "MSH.9.2 must be 3 characters in length.", func() {
				newB().CheckMSH(hl7.Props{"msh_9_1": "ADT", "msh_9_2": "A01Y"})
			})
		})
		t.Run("checkMSH rejects msh_10 longer than 20", func(t *testing.T) {
			expectThrows(t, "MSH.10 must be greater than 0 and less than 20 characters.", func() {
				newB().CheckMSH(hl7.Props{"msh_10": strings.Repeat("A", 21), "msh_9_1": "ADT", "msh_9_2": "A01"})
			})
		})
		t.Run("buildMSH rejects missing msh_11", func(t *testing.T) {
			expectThrows(t, "", func() {
				newB().BuildMSH(hl7.Props{"msh_10": "MSG001", "msh_9_1": "ADT", "msh_9_2": "A01"})
			})
		})
	})

	t.Run("2.3", func(t *testing.T) {
		newB := func() *hl7.Builder { return hl7.New(hl7.V2_3) }
		t.Run("buildMSH produces a 2.3 base header", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "T", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
			contains(t, b.String(), `|ADT^A01|CONTROL_ID|T|2.3`)
		})
		t.Run("buildMSH carries msh_11_2 processing mode", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_11_2": "I", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
			contains(t, b.String(), "|P^I|2.3")
		})
		t.Run("buildMSH carries msh_15 and msh_16 ack types", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_15": "AL", "msh_16": "NE", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
			contains(t, b.String(), "|AL|NE|")
			contains(t, b.String(), "|2.3")
		})
		t.Run("checkMSH accepts a valid 2.3 header", func(t *testing.T) {
			if !newB().CheckMSH(hl7.Props{"msh_11_1": "T", "msh_9_1": "ADT", "msh_9_2": "A01"}) {
				t.Fatal("expected true")
			}
		})
		t.Run("checkMSH rejects msh_11_1 longer than 1", func(t *testing.T) {
			expectThrows(t, "MSH.11.1 has to be 1 character long.", func() {
				newB().CheckMSH(hl7.Props{"msh_11_1": "PT", "msh_9_1": "ADT", "msh_9_2": "A01"})
			})
		})
		t.Run("checkMSH rejects empty-string msh_11_2", func(t *testing.T) {
			expectThrows(t, "MSH.11.2 can either be undefined/blank and 1 character long.", func() {
				newB().CheckMSH(hl7.Props{"msh_11_1": "T", "msh_11_2": "", "msh_9_1": "ADT", "msh_9_2": "A01"})
			})
		})
		t.Run("checkMSH accepts debug processing id", func(t *testing.T) {
			if !newB().CheckMSH(hl7.Props{"msh_11_1": "D", "msh_9_1": "ADT", "msh_9_2": "A01"}) {
				t.Fatal("expected true")
			}
		})
	})

	t.Run("2.3.1", func(t *testing.T) {
		newB := func() *hl7.Builder { return hl7.New(hl7.V2_3_1) }
		t.Run("buildMSH produces a 2.3.1 base header", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
			contains(t, b.String(), `|ADT^A01|CONTROL_ID|P|2.3.1`)
		})
		t.Run("buildMSH carries msh_19 principal language", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_19": "ENG", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
			contains(t, b.String(), "|ENG")
			contains(t, b.String(), "|2.3.1")
		})
		t.Run("checkMSH accepts a valid 2.3.1 header", func(t *testing.T) {
			if !newB().CheckMSH(hl7.Props{"msh_11_1": "P", "msh_9_1": "ADT", "msh_9_2": "A01"}) {
				t.Fatal("expected true")
			}
		})
	})

	t.Run("2.4", func(t *testing.T) {
		newB := func() *hl7.Builder { return hl7.New(hl7.V2_4) }
		t.Run("buildMSH auto-generates msh_9_3", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "T", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
			contains(t, b.String(), `|ADT^A01^ADT_A01|CONTROL_ID|T|2.4`)
		})
		t.Run("buildMSH carries explicit msh_9_3", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01", "msh_9_3": "ADT_A01"})
			contains(t, b.String(), "|ADT^A01^ADT_A01|")
		})
		t.Run("buildMSH carries msh_11_2 T", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_11_2": "T", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
			contains(t, b.String(), "|P^T|2.4")
		})
		t.Run("checkMSH accepts a 2.4 header with msh_9_3", func(t *testing.T) {
			if !newB().CheckMSH(hl7.Props{"msh_11_1": "T", "msh_9_1": "ADT", "msh_9_2": "A01", "msh_9_3": "ADT_A01"}) {
				t.Fatal("expected true")
			}
		})
		t.Run("checkMSH rejects msh_9_3 shorter than 3", func(t *testing.T) {
			expectThrows(t, "MSH.9.3 must be 3 to 10 characters in length if specified.", func() {
				newB().CheckMSH(hl7.Props{"msh_11_1": "T", "msh_9_1": "ADT", "msh_9_2": "A01", "msh_9_3": "AD"})
			})
		})
		t.Run("checkMSH rejects msh_9_3 longer than 10", func(t *testing.T) {
			expectThrows(t, "MSH.9.3 must be 3 to 10 characters in length if specified.", func() {
				newB().CheckMSH(hl7.Props{"msh_11_1": "T", "msh_9_1": "ADT", "msh_9_2": "A01", "msh_9_3": "ADT_A01_ABCDE"})
			})
		})
	})

	t.Run("2.5", func(t *testing.T) {
		b := hl7.New(hl7.V2_5)
		b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "T", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
		contains(t, b.String(), `|ADT^A01^ADT_A01|CONTROL_ID|T|2.5`)
		if !hl7.New(hl7.V2_5).CheckMSH(hl7.Props{"msh_11_1": "P", "msh_9_1": "ORU", "msh_9_2": "R01"}) {
			t.Fatal("expected true")
		}
	})

	t.Run("2.5.1", func(t *testing.T) {
		b := hl7.New(hl7.V2_5_1)
		b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A04"})
		contains(t, b.String(), `|ADT^A04^ADT_A04|CONTROL_ID|P|2.5.1`)
	})

	t.Run("2.6", func(t *testing.T) {
		b := hl7.New(hl7.V2_6)
		b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "D", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
		contains(t, b.String(), `|ADT^A01^ADT_A01|CONTROL_ID|D|2.6`)
		b2 := hl7.New(hl7.V2_6)
		b2.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_3": "SRCSYS", "msh_5": "TGTSYS", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
		contains(t, b2.String(), "|SRCSYS||TGTSYS|")
	})

	t.Run("2.7", func(t *testing.T) {
		newB := func() *hl7.Builder { return hl7.New(hl7.V2_7) }
		t.Run("buildMSH produces a 2.7 base header", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
			contains(t, b.String(), `|ADT^A01^ADT_A01|CONTROL_ID|P|2.7`)
		})
		t.Run("buildMSH carries 2.7 sending application", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_3": "LABSYSTEM", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
			contains(t, b.String(), "|LABSYSTEM||||")
		})
		t.Run("buildMSH carries msh_11_2 archive mode", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_11_2": "A", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
			contains(t, b.String(), "|P^A|2.7")
		})
		t.Run("checkMSH rejects msh_10 longer than 199", func(t *testing.T) {
			expectThrows(t, "MSH.10 must be greater than 0 and less than 199 characters.", func() {
				newB().CheckMSH(hl7.Props{"msh_10": strings.Repeat("A", 200), "msh_11_1": "P", "msh_9_1": "ADT", "msh_9_2": "A01"})
			})
		})
	})

	t.Run("2.7.1", func(t *testing.T) {
		b := hl7.New(hl7.V2_7_1)
		b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": useThisDate, "msh_9_1": "ORM", "msh_9_2": "O01"})
		contains(t, b.String(), `|ORM^O01^ORM_O01|CONTROL_ID|P|2.7.1`)
		b2 := hl7.New(hl7.V2_7_1)
		b2.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_11_2": "R", "msh_7": useThisDate, "msh_9_1": "ORM", "msh_9_2": "O01"})
		contains(t, b2.String(), "|P^R|2.7.1")
	})

	t.Run("2.8", func(t *testing.T) {
		newB := func() *hl7.Builder { return hl7.New(hl7.V2_8) }
		t.Run("buildMSH produces a 2.8 base header", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": useThisDate, "msh_9_1": "ADT", "msh_9_2": "A01"})
			contains(t, b.String(), `|ADT^A01^ADT_A01|CONTROL_ID|P|2.8`)
		})
		t.Run("buildMSH carries explicit msh_9_3 message structure", func(t *testing.T) {
			b := newB()
			b.BuildMSH(hl7.Props{"msh_10": "CONTROL_ID", "msh_11_1": "P", "msh_7": useThisDate, "msh_9_1": "ORU", "msh_9_2": "R01", "msh_9_3": "ORU_R01"})
			contains(t, b.String(), "|ORU^R01^ORU_R01|")
		})
		t.Run("checkMSH accepts a valid 2.8 header", func(t *testing.T) {
			if !newB().CheckMSH(hl7.Props{"msh_11_1": "P", "msh_9_1": "ADT", "msh_9_2": "A01"}) {
				t.Fatal("expected true")
			}
		})
		t.Run("checkMSH rejects msh_9_1 wrong length", func(t *testing.T) {
			expectThrows(t, "MSH.9.1 must be 3 characters in length.", func() {
				newB().CheckMSH(hl7.Props{"msh_11_1": "P", "msh_9_1": "ADTY", "msh_9_2": "A01"})
			})
		})
		t.Run("checkMSH inherits 2.7 msh_10 199-char limit", func(t *testing.T) {
			expectThrows(t, "MSH.10 must be greater than 0 and less than 199 characters.", func() {
				newB().CheckMSH(hl7.Props{"msh_10": strings.Repeat("A", 200), "msh_11_1": "P", "msh_9_1": "ADT", "msh_9_2": "A01"})
			})
		})
	})
}

func TestBuildMSHRejectsSecond(t *testing.T) {
	b := hl7.New(hl7.V2_1)
	b.BuildMSH(hl7.Props{"msh_10": "12345", "msh_11": "T", "msh_9": "ACK"})
	expectThrows(t, "You can only have one MSH Header per HL7 Message.", func() {
		b.BuildMSH(hl7.Props{"msh_10": "12345", "msh_11": "T", "msh_9": "ACK"})
	})
}

func TestBuildADDCannotFollowMSH(t *testing.T) {
	b := hl7.New(hl7.V2_1)
	b.BuildMSH(hl7.Props{"msh_10": "12345", "msh_11": "T", "msh_9": "ACK"})
	expectThrows(t, "This segment must not follow a MSH, BHS, or FHS", func() {
		b.BuildADD(hl7.Props{"add_1": "Fail cause you can't have this after MSH"})
	})
}
