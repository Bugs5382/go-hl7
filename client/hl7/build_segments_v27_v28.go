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

import "strconv"

// itoa is a local strconv.Itoa shorthand for building "<seg>_<num>" prop keys.
func itoa(n int) string { return strconv.Itoa(n) }

// BuildIPC builds an IPC (Imaging Procedure Control) segment (the
// HL7_2_7._buildIPC). IPC is introduced in v2.7. Chainable.
func (b *Builder) BuildIPC(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.7")
	s := b.spec("IPC")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("IPC")
	b.setField(s, 1, pick(p, "ipc_1"), &ValidationRule{Length: lenMinMax(1, 427)})
	b.setField(s, 2, pick(p, "ipc_2"), &ValidationRule{Length: lenMinMax(1, 22)})
	b.setField(s, 3, pick(p, "ipc_3"), &ValidationRule{Length: lenMinMax(1, 70)})
	b.setField(s, 4, pick(p, "ipc_4"), &ValidationRule{Length: lenMinMax(1, 22)})
	b.setField(s, 5, pick(p, "ipc_5"), &ValidationRule{Length: lenMinMax(1, 16)})
	b.setField(s, 6, pick(p, "ipc_6"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 7, pick(p, "ipc_7"), &ValidationRule{Length: lenMinMax(1, 22)})
	b.setField(s, 8, pick(p, "ipc_8"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}

// BuildISD builds an ISD (Interaction Status Detail) segment (the
// HL7_2_7._buildISD). ISD.1 is coerced to a string. Chainable.
func (b *Builder) BuildISD(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.7")
	s := b.spec("ISD")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("ISD")
	isd1 := ""
	if v := pick(p, "isd_1"); v != nil {
		isd1 = toStr(v)
	}
	b.setField(s, 1, isd1, &ValidationRule{Length: lenMinMax(1, 10)})
	b.setField(s, 2, pick(p, "isd_2"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 3, pick(p, "isd_3"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}

// BuildSTZ builds an STZ (Sterilization Parameter) segment (the
// HL7_2_8._buildSTZ). STZ is introduced in v2.8. Chainable.
func (b *Builder) BuildSTZ(p Props) *Builder {
	if b.err != nil {
		return b
	}
	b.headerExists()
	b.notImplementedBefore("2.8")
	s := b.spec("STZ")
	b.assertSegmentInVersion(s)
	b.segment = b.mustAddSegment("STZ")
	b.setField(s, 1, pick(p, "stz_1"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 2, pick(p, "stz_2"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 3, pick(p, "stz_3"), &ValidationRule{Length: lenMinMax(1, 250)})
	b.setField(s, 4, pick(p, "stz_4"), &ValidationRule{Length: lenMinMax(1, 250)})
	return b
}

// toStr coerces a prop value to its string form for ISD.1's String(...) wrap.
func toStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return strconv.FormatInt(int64(asInt(v)), 10)
}

func asInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return 0
}
