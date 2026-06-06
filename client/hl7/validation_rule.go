// Package hl7 is the spec-driven, per-version HL7 v2 message builder. It mirrors
// the hl7/ tree: the Builder base builder, the per-version HL7_2_x
// classes (a 2.1 -> 2.8 inheritance chain), the spec-driven field validator
// keyed off the generated metadata catalog, and the composite-object field
// assembly. The base builder writes into a builder.Message and toString()
// serializes the wire format.
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
	"regexp"

	"github.com/Bugs5382/go-hl7/client/hl7/metadata"
)

// ruleType is the input type for a ValidationRule, controlling which checks
// apply. It mirrors the ValidationRule.type union ("date"|"number"|
// "string"), defaulting to "string".
type ruleType string

const (
	// ruleString validates the value as a string (the default).
	ruleString ruleType = "string"
	// ruleNumber validates the value as a number.
	ruleNumber ruleType = "number"
	// ruleDate validates the value as an HL7 date/time.
	ruleDate ruleType = "date"
)

// numberBound mirrors the ValidationRule.number min/max bounds. The
// Has* flags stand in for TypeScript's optional members.
type numberBound struct {
	Min    float64
	HasMin bool
	Max    float64
	HasMax bool
}

// ValidationRule is the per-field validation rule consumed by the validator. It
// mirrors the ValidationRule type. TypeScript optional members become
// pointer/Has* flags (Go necessity), and the regexp.Regexp pattern replaces
// the RegExp.
type ValidationRule struct {
	// AllowedValues is the set of valid values. Requires Type to be ruleString.
	AllowedValues []string
	// DependsOn is the conditional dependency on another field.
	DependsOn *metadata.Depends
	// Deprecated, when true, warns the field should not be used (value still
	// serializes).
	Deprecated bool
	// HL7Support is the version expression(s) this field is valid for, e.g.
	// ">=2.3" or [">=2.1", "<=2.6"].
	HL7Support []string
	// HL7Type is the HL7 data type identifier (documentation only).
	HL7Type string
	// Length is the exact length or min/max bounds for string values.
	Length metadata.Length
	// Number is the min/max bounds for numeric values. Requires Type ruleNumber.
	Number *numberBound
	// Pattern is a regexp the value must match.
	Pattern *regexp.Regexp
	// Required, when true, requires the field to be present and non-empty.
	Required bool
	// Type controls which validations apply. Defaults to ruleString.
	Type ruleType
	// HasType reports whether Type was explicitly set (Go stand-in for the
	// TypeScript optional member; lets the validator default it to ruleString).
	HasType bool
	// Usage is the HL7 usage code driving Required/Deprecated and W/X rejection.
	Usage metadata.HL7UsageCode
	// HasUsage reports whether Usage was set.
	HasUsage bool
	// UseField names the replacement field when Deprecated is true.
	UseField string
}
