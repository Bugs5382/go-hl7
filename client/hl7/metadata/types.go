// Package metadata carries the machine-readable HL7 segment and data-type
// specs that drive the spec-driven builder and per-version usage validation. It
// mirrors the hl7/metadata tree (types.ts plus the generated segment and
// data-type catalogs aggregated into SegmentSpecs and DataTypes).
package metadata

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

// HL7Version is a supported HL7 v2 spec version. It mirrors the reference's
// HL7Version union ("2.1".."2.8").
type HL7Version string

// Supported HL7 versions. The list mirrors the HL7Version union.
const (
	V21  HL7Version = "2.1"
	V22  HL7Version = "2.2"
	V23  HL7Version = "2.3"
	V231 HL7Version = "2.3.1"
	V24  HL7Version = "2.4"
	V25  HL7Version = "2.5"
	V251 HL7Version = "2.5.1"
	V26  HL7Version = "2.6"
	V27  HL7Version = "2.7"
	V271 HL7Version = "2.7.1"
	V28  HL7Version = "2.8"
)

// KnownVersions lists every supported HL7 v2 version, in ascending order. It is
// the canonical set used to validate a connection's or listener's required
// version.
var KnownVersions = []HL7Version{V21, V22, V23, V231, V24, V25, V251, V26, V27, V271, V28}

// IsKnownVersion reports whether v is one of the supported HL7 v2 versions.
func IsKnownVersion(v string) bool {
	for _, known := range KnownVersions {
		if string(known) == v {
			return true
		}
	}
	return false
}

// HL7UsageCode is an HL7 v2 spec usage code as defined by the HL7 standard.
// It mirrors the HL7UsageCode union.
//
//   - R Required: field must be populated.
//   - O Optional: field may or may not be present.
//   - B Backward Compatibility: kept for older versions, deprecated.
//   - W Withdrawn: element is not to be used; treated as X.
//   - D Dependent / Conditional: usage depends on another field or trigger.
//   - X Not Supported: element cannot be present.
type HL7UsageCode string

// Usage codes. The set mirrors the HL7UsageCode union.
const (
	UsageRequired   HL7UsageCode = "R"
	UsageOptional   HL7UsageCode = "O"
	UsageBackward   HL7UsageCode = "B"
	UsageWithdrawn  HL7UsageCode = "W"
	UsageDependent  HL7UsageCode = "D"
	UsageNotSupport HL7UsageCode = "X"
)

// Length is the exact length or min/max bounds for a field or component value.
// The reference models this as `number | { min?, max? }`; Go has no union, so Length
// carries an exact value or a min/max with a flag indicating which form is set.
type Length struct {
	// Exact is the exact length when HasExact is true.
	Exact int
	// Min is the lower bound when HasExact is false.
	Min int
	// Max is the upper bound when HasExact is false.
	Max int
	// HasExact reports whether Exact (rather than Min/Max) carries the bound.
	HasExact bool
	// Set reports whether any length bound was published for this field.
	Set bool
}

// Depends is a conditional dependency used when the usage code is "D". If
// present, the dependency must resolve before the field is considered
// satisfied. It mirrors the dependsOn shape.
type Depends struct {
	// Path is the dependent field path.
	Path string
	// MustBeSet requires the dependent path to be populated.
	MustBeSet bool
	// MustEqual requires the dependent path to equal this value, when set.
	MustEqual string
	// HasMustEqual reports whether MustEqual carries a value.
	HasMustEqual bool
}

// ComponentSpec describes one sub-component inside a composite field. For
// composite HL7 data types (XAD, XPN, CE/CWE, CX, ...) the field value is a
// ^-delimited list whose pieces each have their own data type, length,
// optionality, and possibly a table reference. It mirrors the reference's
// ComponentSpec.
type ComponentSpec struct {
	// Num is the 1-based position within the field.
	Num int
	// Name is the human-readable component name.
	Name string
	// HL7Type is the HL7 data type, e.g. "ST", "DTM", "SAD".
	HL7Type string
	// Length is the exact length or min/max bounds.
	Length Length
	// Table is the numeric HL7 table id, if the component is enumerated.
	Table int
	// Rpt is the cardinality from Caristix ("1" single, "*" unbounded, ...).
	Rpt string
	// Usage is the HL7 usage code at this component position.
	Usage HL7UsageCode
}

// FieldSpec describes a single field within a segment, with per-version usage
// codes. A version missing from Usage means the field does NOT exist in that
// version of the spec; attempting to set it must be rejected at runtime. It
// mirrors the FieldSpec.
type FieldSpec struct {
	// Num is the field number within the segment, e.g. ECD.4 -> 4.
	Num int
	// Name is the human-readable field name from the HL7 spec.
	Name string
	// HL7Type is the HL7 data type identifier, e.g. "ST", "NM", "DTM", "CWE".
	HL7Type string
	// Length is the exact length or min/max bounds for string values.
	Length Length
	// Table is the HL7 table id used to validate this field, if any.
	Table int
	// AllowedValues is a restricted set of allowed string values.
	AllowedValues []string
	// Components are the sub-components for composite data types. Empty/absent
	// for primitive fields.
	Components []ComponentSpec
	// DependsOn is the conditional dependency used when Usage is "D".
	DependsOn *Depends
	// Usage is the per-version usage code map. A version key being absent means
	// the field is not part of that version of the spec at all.
	Usage map[HL7Version]HL7UsageCode
}

// SegmentSpec describes an HL7 segment, with per-version availability and field
// set. It mirrors the SegmentSpec.
type SegmentSpec struct {
	// Name is the three-letter segment name, e.g. "ECD".
	Name string
	// Description is the human-readable description of the segment.
	Description string
	// Versions are the HL7 versions in which the segment exists at all. A
	// builder for a version not in this list must reject building the segment.
	Versions []HL7Version
	// Fields are all fields defined for this segment across any version.
	Fields []FieldSpec
}

// DataTypeSpec describes a composite HL7 data type (XAD, XPN, CE, ...) as a
// list of component positions. It mirrors the data-type definitions used
// by the composite-object field assembly.
type DataTypeSpec struct {
	// Name is the data type name, e.g. "XAD".
	Name string
	// Description is the human-readable description.
	Description string
	// Components are the ordered component positions.
	Components []ComponentSpec
}
