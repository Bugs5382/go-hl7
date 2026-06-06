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

// The per-version builders mirror the 2.1 -> 2.8 class chain. The spec
// expresses version differences by overriding protected _buildXXX / checkMSH;
// most field-level differences, however, are driven by the per-version usage
// catalog the validator already consults (validatorSetField reads b.version),
// so the shared segment builders live on HL7_BASE and are version-aware. The
// structural MSH differences (field 9 vs 9.1/9.2/9.3, 11 vs 11.1/11.2, the
// trailing fields, and the per-version max lengths) are dispatched by version
// in BuildMSH. checkMSH is dispatched by version in CheckMSH.
//
// Go necessity (documented adaptation): the spec embeds the version chain in
// TypeScript class inheritance and narrows withdrawn fields to `never` at the
// type level (e.g. v2.8 ECD.4). Go has no type-level narrowing and no virtual
// dispatch through struct embedding, so every HL7_2_x constructor returns the
// shared *HL7_BASE carrying its version string, and the version narrowing
// collapses to the runtime usage-code check (W/X reject a value), which the
// validator and assertSegmentInVersion enforce identically.

// optsOf returns the first option set, or defaults when none is given. It lets
// the constructors take an optional Options argument like the constructor.
func optsOf(opts []Options) Options {
	if len(opts) > 0 {
		return opts[0]
	}
	return Options{}
}

func newVersion(version string, opts []Options) *HL7_BASE {
	b := &HL7_BASE{}
	b.initBase(opts)
	b.version = version
	b.maxAddSegmentLength, b.hasMaxAddSegment = 60, true
	return b
}

// NewHL7_2_1 constructs a v2.1 builder. It mirrors `new HL7_2_1(properties?)`.
func NewHL7_2_1(opts ...Options) *HL7_BASE { return newVersion("2.1", opts) }

// NewHL7_2_2 constructs a v2.2 builder.
func NewHL7_2_2(opts ...Options) *HL7_BASE { return newVersion("2.2", opts) }

// NewHL7_2_3 constructs a v2.3 builder.
func NewHL7_2_3(opts ...Options) *HL7_BASE { return newVersion("2.3", opts) }

// NewHL7_2_3_1 constructs a v2.3.1 builder.
func NewHL7_2_3_1(opts ...Options) *HL7_BASE { return newVersion("2.3.1", opts) }

// NewHL7_2_4 constructs a v2.4 builder. ECD is introduced here.
func NewHL7_2_4(opts ...Options) *HL7_BASE { return newVersion("2.4", opts) }

// NewHL7_2_5 constructs a v2.5 builder.
func NewHL7_2_5(opts ...Options) *HL7_BASE { return newVersion("2.5", opts) }

// NewHL7_2_5_1 constructs a v2.5.1 builder.
func NewHL7_2_5_1(opts ...Options) *HL7_BASE { return newVersion("2.5.1", opts) }

// NewHL7_2_6 constructs a v2.6 builder.
func NewHL7_2_6(opts ...Options) *HL7_BASE { return newVersion("2.6", opts) }

// NewHL7_2_7 constructs a v2.7 builder. There is no implicit default version:
// the caller selects the spec by which HL7_2_x constructor they invoke.
func NewHL7_2_7(opts ...Options) *HL7_BASE { return newVersion("2.7", opts) }

// NewHL7_2_7_1 constructs a v2.7.1 builder.
func NewHL7_2_7_1(opts ...Options) *HL7_BASE { return newVersion("2.7.1", opts) }

// NewHL7_2_8 constructs a v2.8 builder.
func NewHL7_2_8(opts ...Options) *HL7_BASE { return newVersion("2.8", opts) }
