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
	"fmt"
	"time"

	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/hl7/metadata"
	"github.com/Bugs5382/go-hl7/client/utils"
)

// pick returns the first present, non-empty value among the given prop keys, or
// nil. It mirrors the `properties.msh_3 || properties.sendingApplication`
// fallback chains.
func pick(p Props, keys ...string) any {
	for _, k := range keys {
		if v, ok := p[k]; ok {
			if v == nil {
				continue
			}
			if s, isStr := v.(string); isStr && s == "" {
				continue
			}
			return v
		}
	}
	return nil
}

// pickStr is pick coerced to a string ("" when absent).
func pickStr(p Props, keys ...string) string {
	v := pick(p, keys...)
	if v == nil {
		return ""
	}
	return fmt.Sprint(v)
}

// dateField resolves a Date-typed prop: a time.Time formats at length, anything
// else passes through, and an absent value yields fallback. It mirrors the
// `x instanceof Date && !isNaN ? setDate(x, len) : fallback` idiom.
func (b *HL7_BASE) dateField(v any, length string, fallback any) any {
	if t, ok := v.(time.Time); ok && !t.IsZero() {
		return b.SetDate(t, length)
	}
	if v == nil {
		return fallback
	}
	return v
}

// lenMax is a max-only Length rule.
func lenMax(max int) metadata.Length { return metadata.Length{Max: max, Set: true} }

// lenMinMax is a min/max Length rule.
func lenMinMax(min, max int) metadata.Length { return metadata.Length{Min: min, Max: max, Set: true} }

// lenExact is an exact Length rule.
func lenExact(n int) metadata.Length { return metadata.Length{Exact: n, HasExact: true, Set: true} }

// dateRule is a date-type rule.
func dateRule() *ValidationRule { return &ValidationRule{Type: ruleDate, HasType: true} }

// BuildMSH builds the MSH header, enforcing the single-MSH rule and dispatching
// the per-version framing (the buildMSH -> version _buildMSH). Returns the
// receiver for chaining.
func (b *HL7_BASE) BuildMSH(properties Props) *HL7_BASE {
	b.buildMSHGuard()
	switch b.version {
	case "2.1":
		b.buildMSH21(properties)
	case "2.2":
		b.buildMSH22(properties)
	case "2.3":
		b.buildMSH23(properties)
	case "2.3.1":
		b.buildMSH23(properties)
		b.validatorSetValue("19", pick(properties, "msh_19"), &ValidationRule{Length: lenMinMax(1, 60)})
		b.validatorSetValue("20", pick(properties, "msh_20"), &ValidationRule{AllowedValues: b.codeTable("0356")})
	case "2.4", "2.5", "2.5.1", "2.6":
		b.buildMSH24(properties)
	default: // 2.7, 2.7.1, 2.8
		b.buildMSH27(properties)
	}
	return b
}

// mshSeparators returns the MSH.1/2 separator string from the options.
func (b *HL7_BASE) mshSeparators() string {
	return b.opt.SeparatorComponent + b.opt.SeparatorRepetition + b.opt.SeparatorEscape + b.opt.SeparatorSubComponent
}

func (b *HL7_BASE) mshCommonHead(properties Props, recvFacMax int) {
	b.segment = mustAddSegment(b.message, "MSH")
	if len(b.opt.SeparatorComponent) != 1 {
		panic(helpers.NewHL7ValidationError("Separator Component has to be a single character."))
	}
	b.validatorSetValue("1", b.mshSeparators(), &ValidationRule{Length: lenExact(4), Required: true})
	b.validatorSetValue("3", pick(properties, "msh_3", "sendingApplication"), &ValidationRule{Length: lenMinMax(1, 15)})
	b.validatorSetValue("4", pick(properties, "msh_4", "sendingFacility"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.validatorSetValue("5", pick(properties, "msh_5", "receivingApplication"), &ValidationRule{Length: lenMinMax(1, 15)})
	b.validatorSetValue("6", pick(properties, "msh_6", "receivingFacility"), &ValidationRule{Length: lenMinMax(1, recvFacMax)})
}

// buildMSH21 ports the HL7_2_1._buildMSH (MSH.9 single field, MSH.11 P/T).
func (b *HL7_BASE) buildMSH21(properties Props) {
	b.segment = mustAddSegment(b.message, "MSH")
	if len(b.opt.SeparatorComponent) != 1 {
		panic(helpers.NewHL7ValidationError("Separator Component has to be a single character."))
	}
	b.validatorSetValue("1", b.mshSeparators(), &ValidationRule{Length: lenExact(4), Required: true})
	b.validatorSetValue("3", pick(properties, "msh_3", "sendingApplication"), &ValidationRule{Length: lenMinMax(1, 15)})
	b.validatorSetValue("4", pick(properties, "msh_4", "sendingFacility"), &ValidationRule{Length: lenMinMax(1, 20)})
	b.validatorSetValue("5", pick(properties, "msh_5", "receivingApplication"), &ValidationRule{Length: lenMinMax(1, 15)})
	b.validatorSetValue("6", pick(properties, "msh_6", "receivingFacility"), &ValidationRule{Length: lenMinMax(1, 30)})
	b.validatorSetValue("7", b.dateField(pick(properties, "msh_7"), b.opt.Date, b.SetDate(time.Now(), b.opt.Date)), &ValidationRule{Length: lenMinMax(8, 19), Required: true, Type: ruleDate, HasType: true})
	b.validatorSetValue("8", pick(properties, "msh_8"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.validatorSetValue("9", pick(properties, "msh_9"), &ValidationRule{AllowedValues: b.codeTable("0076"), Required: true})
	b.validatorSetValue("10", mshControlID(properties), &ValidationRule{Length: lenMinMax(1, 20)})
	b.validatorSetValue("11", pick(properties, "msh_11"), &ValidationRule{AllowedValues: []string{"P", "T"}, Length: lenExact(1), Required: true})
	b.validatorSetValue("12", b.version, &ValidationRule{Required: true})
}

// buildMSH22 ports the HL7_2_2._buildMSH (MSH.9 splits to 9.1/9.2).
func (b *HL7_BASE) buildMSH22(properties Props) {
	b.mshCommonHead(properties, 20)
	b.validatorSetValue("7", b.dateField(pick(properties, "msh_7"), b.opt.Date, b.SetDate(time.Now(), b.opt.Date)), &ValidationRule{Required: true, Type: ruleDate, HasType: true})
	b.validatorSetValue("8", pick(properties, "msh_8"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.validatorSetValue("9.1", pick(properties, "msh_9_1"), &ValidationRule{Length: lenMinMax(1, 3), Required: true})
	b.validatorSetValue("9.2", pick(properties, "msh_9_2"), &ValidationRule{Length: lenMinMax(1, 3), Required: true})
	b.validatorSetValue("10", mshControlID(properties), &ValidationRule{Length: lenMinMax(1, 20)})
	b.validatorSetValue("11", pick(properties, "msh_11"), &ValidationRule{AllowedValues: []string{"P", "T"}, Length: lenExact(1), Required: true})
	b.validatorSetValue("12", b.version, &ValidationRule{Required: true})
}

// buildMSH23 ports the HL7_2_3._buildMSH (11 splits to 11.1/11.2, +13-18).
func (b *HL7_BASE) buildMSH23(properties Props) {
	b.mshCommonHead(properties, 20)
	b.validatorSetValue("7", b.dateField(pick(properties, "msh_7"), b.opt.Date, b.SetDate(time.Now(), b.opt.Date)), &ValidationRule{Required: true, Type: ruleDate, HasType: true})
	b.validatorSetValue("8", pick(properties, "msh_8"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.validatorSetValue("9.1", pick(properties, "msh_9_1"), &ValidationRule{Length: lenMinMax(1, 3), Required: true})
	b.validatorSetValue("9.2", pick(properties, "msh_9_2"), &ValidationRule{Length: lenMinMax(1, 3), Required: true})
	b.validatorSetValue("10", mshControlID(properties), &ValidationRule{Length: lenMinMax(1, 20)})
	b.validatorSetValue("11.1", pick(properties, "msh_11_1"), &ValidationRule{AllowedValues: []string{"D", "P", "T"}, Length: lenExact(1), Required: true})
	if v := pickStr(properties, "msh_11_2"); v != "" {
		b.validatorSetValue("11.2", v, &ValidationRule{AllowedValues: []string{"A", "I", "R"}, Length: lenExact(1)})
	}
	b.validatorSetValue("12", b.version, &ValidationRule{Required: true})
	b.validatorSetValue("13", pick(properties, "msh_13"), &ValidationRule{Length: lenMinMax(1, 15)})
	b.validatorSetValue("14", pick(properties, "msh_14"), &ValidationRule{Length: lenMinMax(1, 180)})
	b.validatorSetValue("15", pick(properties, "msh_15"), &ValidationRule{AllowedValues: b.codeTable("0155")})
	b.validatorSetValue("16", pick(properties, "msh_16"), &ValidationRule{AllowedValues: b.codeTable("0155")})
	b.validatorSetValue("17", pick(properties, "msh_17"), &ValidationRule{Length: lenExact(3)})
	b.validatorSetValue("18", pick(properties, "msh_18"), &ValidationRule{Length: lenMinMax(1, 16)})
}

// buildMSH24 ports the HL7_2_4._buildMSH (adds MSH.9.3, 11.2 incl T).
func (b *HL7_BASE) buildMSH24(properties Props) {
	b.mshCommonHead(properties, 20)
	b.validatorSetValue("7", b.dateField(pick(properties, "msh_7"), b.opt.Date, b.SetDate(time.Now(), b.opt.Date)), &ValidationRule{Required: true, Type: ruleDate, HasType: true})
	b.validatorSetValue("8", pick(properties, "msh_8"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.validatorSetValue("9.1", pick(properties, "msh_9_1"), &ValidationRule{Length: lenMinMax(1, 3), Required: true})
	b.validatorSetValue("9.2", pick(properties, "msh_9_2"), &ValidationRule{Length: lenMinMax(1, 3), Required: true})
	b.validatorSetValue("9.3", mshStructure(properties), &ValidationRule{Length: lenMinMax(3, 15)})
	b.validatorSetValue("10", mshControlID(properties), &ValidationRule{Length: lenMinMax(1, 20)})
	b.validatorSetValue("11.1", pick(properties, "msh_11_1"), &ValidationRule{AllowedValues: []string{"D", "P", "T"}, Length: lenExact(1), Required: true})
	if v := pickStr(properties, "msh_11_2"); v != "" {
		b.validatorSetValue("11.2", v, &ValidationRule{AllowedValues: []string{"A", "I", "R", "T"}, Length: lenExact(1)})
	}
	b.validatorSetValue("12", b.version, &ValidationRule{Required: true})
}

// buildMSH27 ports the HL7_2_7._buildMSH (227-char fields, MSH.10 max 199).
func (b *HL7_BASE) buildMSH27(properties Props) {
	b.segment = mustAddSegment(b.message, "MSH")
	if len(b.opt.SeparatorComponent) != 1 {
		panic(helpers.NewHL7ValidationError("Separator Component has to be a single character."))
	}
	b.validatorSetValue("1", b.mshSeparators(), &ValidationRule{Length: lenExact(4), Required: true})
	b.validatorSetValue("3", pick(properties, "msh_3", "sendingApplication"), &ValidationRule{Length: lenMinMax(1, 227)})
	b.validatorSetValue("4", pick(properties, "msh_4", "sendingFacility"), &ValidationRule{Length: lenMinMax(1, 227)})
	b.validatorSetValue("5", pick(properties, "msh_5", "receivingApplication"), &ValidationRule{Length: lenMinMax(1, 227)})
	b.validatorSetValue("6", pick(properties, "msh_6", "receivingFacility"), &ValidationRule{Length: lenMinMax(1, 227)})
	b.validatorSetValue("7", b.dateField(pick(properties, "msh_7"), b.opt.Date, b.SetDate(time.Now(), b.opt.Date)), &ValidationRule{Required: true, Type: ruleDate, HasType: true})
	b.validatorSetValue("8", pick(properties, "msh_8"), &ValidationRule{Length: lenMinMax(1, 40)})
	b.validatorSetValue("9.1", pick(properties, "msh_9_1"), &ValidationRule{Length: lenMinMax(1, 3), Required: true})
	b.validatorSetValue("9.2", pick(properties, "msh_9_2"), &ValidationRule{Length: lenMinMax(1, 3), Required: true})
	b.validatorSetValue("9.3", mshStructure(properties), &ValidationRule{Length: lenMinMax(3, 15)})
	b.validatorSetValue("10", mshControlID199(properties), &ValidationRule{Length: lenMinMax(1, 199)})
	b.validatorSetValue("11.1", pick(properties, "msh_11_1"), &ValidationRule{AllowedValues: []string{"D", "P", "T"}, Length: lenExact(1), Required: true})
	if v := pickStr(properties, "msh_11_2"); v != "" {
		b.validatorSetValue("11.2", v, &ValidationRule{AllowedValues: []string{"A", "I", "R", "T"}, Length: lenExact(1)})
	}
	b.validatorSetValue("12", b.version, &ValidationRule{Required: true})
}

// mshControlID resolves MSH.10, defaulting to a random string (the
// `msh_10 || randomString()`).
func mshControlID(p Props) any {
	if v := pickStr(p, "msh_10"); v != "" {
		return v
	}
	return utils.RandomString(20)
}

func mshControlID199(p Props) any { return mshControlID(p) }

// mshStructure resolves MSH.9.3, defaulting to "<9.1>_<9.2>" (the idiom).
func mshStructure(p Props) any {
	if v, ok := p["msh_9_3"]; ok && v != nil {
		return v
	}
	return pickStr(p, "msh_9_1") + "_" + pickStr(p, "msh_9_2")
}
