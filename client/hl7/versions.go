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

	"github.com/Bugs5382/go-hl7/client/helpers"
)

// The builder spans the 2.1 -> 2.8 spec range. Most field-level differences are
// driven by the per-version usage catalog the validator already consults
// (validatorSetField reads b.version), so the shared segment builders live on
// Builder and are version-aware. The structural MSH differences (field 9 vs
// 9.1/9.2/9.3, 11 vs 11.1/11.2, the trailing fields, and the per-version max
// lengths) are dispatched by version in BuildMSH. checkMSH is dispatched by
// version in CheckMSH.
//
// New returns a single *Builder carrying its version string; per-version field
// availability is enforced at runtime by the usage-code check (W/X reject a
// value), which the validator and assertSegmentInVersion apply identically.

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

// knownVersions is the set of spec versions New accepts. An unknown version is
// recorded onto the builder so it surfaces at ToMessage/Err rather than building
// against a spec that does not exist.
var knownVersions = map[Version]bool{
	V2_1: true, V2_2: true, V2_3: true, V2_3_1: true, V2_4: true,
	V2_5: true, V2_5_1: true, V2_6: true, V2_7: true, V2_7_1: true, V2_8: true,
}

func newVersion(version string, opts []Options) *Builder {
	b := &Builder{}
	b.initBase(opts)
	b.version = version
	b.maxAddSegmentLength, b.hasMaxAddSegment = 60, true
	return b
}

// New constructs a Builder for the given HL7 spec version, e.g.
// hl7.New(hl7.V2_5). ECD is introduced at V2_4. An unrecognized version does not
// fail New itself; the builder records an HL7ValidationError that the first
// Build* call short-circuits on and ToMessage/Err returns.
func New(v Version, opts ...Options) *Builder {
	b := newVersion(string(v), opts)
	if !knownVersions[v] {
		b.fail(helpers.NewHL7ValidationError(
			fmt.Sprintf("Unknown HL7 version %q — use one of the hl7.V2_x constants", string(v))))
	}
	return b
}
