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

// The builder spans the 2.1 -> 2.8 spec range. Most field-level differences are
// driven by the per-version usage catalog the validator already consults
// (validatorSetField reads b.version), so the shared segment builders live on
// Builder and are version-aware. The structural MSH differences (field 9 vs
// 9.1/9.2/9.3, 11 vs 11.1/11.2, the trailing fields, and the per-version max
// lengths) are dispatched by version in BuildMSH. checkMSH is dispatched by
// version in CheckMSH.
//
// Go necessity (documented adaptation): the spec embeds the version chain in
// TypeScript class inheritance and narrows withdrawn fields to `never` at the
// type level (e.g. v2.8 ECD.4). Go has no type-level narrowing and no virtual
// dispatch through struct embedding, so New returns a single *Builder carrying
// its version string, and the version narrowing collapses to the runtime
// usage-code check (W/X reject a value), which the validator and
// assertSegmentInVersion enforce identically.

// Version is an HL7 spec version selector passed to New. Use the exported
// V2_x constants rather than a bare string literal.
type Version string

// The supported HL7 spec versions. New requires one; there is no default, so a
// builder always carries an explicit spec.
const (
	V2_1   Version = "2.1"
	V2_2   Version = "2.2"
	V2_3   Version = "2.3"
	V2_3_1 Version = "2.3.1"
	V2_4   Version = "2.4"
	V2_5   Version = "2.5"
	V2_5_1 Version = "2.5.1"
	V2_6   Version = "2.6"
	V2_7   Version = "2.7"
	V2_7_1 Version = "2.7.1"
	V2_8   Version = "2.8"
)

// optsOf returns the first option set, or defaults when none is given. It lets
// New take an optional Options argument.
func optsOf(opts []Options) Options {
	if len(opts) > 0 {
		return opts[0]
	}
	return Options{}
}

func newVersion(version string, opts []Options) *Builder {
	b := &Builder{}
	b.initBase(opts)
	b.version = version
	b.maxAddSegmentLength, b.hasMaxAddSegment = 60, true
	return b
}

// New constructs a Builder for the given HL7 spec version, e.g.
// hl7.New(hl7.V2_5). ECD is introduced at V2_4.
func New(v Version, opts ...Options) *Builder { return newVersion(string(v), opts) }
