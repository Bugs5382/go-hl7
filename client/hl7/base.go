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
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/hl7/metadata"
	"github.com/Bugs5382/go-hl7/client/utils"
)

// Props is a segment-build input: spec property names ("msh_9_1", "pid_11",
// "sendingApplication", ...) mapped to their values (string, time.Time, int,
// or a composite-object map for composite fields). It is the faithful Go
// counterpart to the duck-typed `properties` object literals; modeling
// it as a map (rather than a fixed struct) is the only way to accept the exact
// same heterogeneous inputs the spec accepts, including a composite field that may
// be either a pre-formatted string or a typed component object.
type Props = map[string]any

// Options configures a builder. It mirrors the ClientBuilderOptions: the
// separator characters, the date length, and the hardError flag.
type Options struct {
	// Text is an optional pre-parsed message body (rarely used by builders).
	Text string
	// Date is the HL7 date length ("8"/"12"/"14"/"19"/"24"/"26").
	Date string
	// SeparatorField is the field separator (default "|").
	SeparatorField string
	// SeparatorComponent is the component separator (default "^").
	SeparatorComponent string
	// SeparatorRepetition is the repetition separator (default "~").
	SeparatorRepetition string
	// SeparatorEscape is the escape separator (default "\\").
	SeparatorEscape string
	// SeparatorSubComponent is the sub-component separator (default "&").
	SeparatorSubComponent string
	// HardError forces every deviation to throw regardless of usage softness.
	HardError bool
}

// HL7_BASE is the base class of an HL7 specification builder. It mirrors
// the HL7_BASE: it owns the Message being built, the current segment,
// the spec-driven field validator, and the composite-object assembly. Each
// per-version HL7_2_x embeds HL7_BASE and sets version plus typed BuildXXX
// methods.
type HL7_BASE struct {
	// version is the HL7 spec version, e.g. "2.7". There is no default: each
	// version constructor sets it, so a builder always carries an explicit spec.
	version string

	opt     Options
	message *builder.Message
	// segment is the segment currently being built (the _segment).
	segment *builder.Segment
	// maxAddSegmentLength bounds the ADD segment; set by the version subtype.
	maxAddSegmentLength int
	hasMaxAddSegment    bool
	// hardError forces every deviation to throw (the hardError).
	hardError bool

	// errorHandlers / warningHandlers back the EventEmitter "error"/"warning"
	// events the validator emits. Go necessity: the reference extends an
	// EventEmitter; this is a minimal On(event, handler) over the same event names.
	errorHandlers   []func(string)
	warningHandlers []func(string)
}

// initBase wires the message and options. It mirrors the HL7_BASE constructor.
func (b *HL7_BASE) initBase(opts []Options) {
	opt := optsOf(opts)
	b.opt = normalizeOptions(opt)
	b.hardError = opt.HardError
	m, _ := builder.NewMessage(builder.MessageOptions{
		Date:                  b.opt.Date,
		SeparatorField:        b.opt.SeparatorField,
		SeparatorComponent:    b.opt.SeparatorComponent,
		SeparatorRepetition:   b.opt.SeparatorRepetition,
		SeparatorEscape:       b.opt.SeparatorEscape,
		SeparatorSubComponent: b.opt.SeparatorSubComponent,
	})
	b.message = m
}

// normalizeOptions fills option defaults, mirroring normalizedClientBuilderOptions.
func normalizeOptions(opt Options) Options {
	if opt.SeparatorField == "" {
		opt.SeparatorField = "|"
	}
	if opt.SeparatorComponent == "" {
		opt.SeparatorComponent = "^"
	}
	if opt.SeparatorRepetition == "" {
		opt.SeparatorRepetition = "~"
	}
	if opt.SeparatorEscape == "" {
		opt.SeparatorEscape = "\\"
	}
	if opt.SeparatorSubComponent == "" {
		opt.SeparatorSubComponent = "&"
	}
	return opt
}

// Version returns the HL7 spec version (the version field).
func (b *HL7_BASE) Version() string { return b.version }

// On registers a handler for the "error" or "warning" event, mirroring the
// EventEmitter.on used by the validator.
func (b *HL7_BASE) On(event string, handler func(string)) {
	switch event {
	case "error":
		b.errorHandlers = append(b.errorHandlers, handler)
	case "warning":
		b.warningHandlers = append(b.warningHandlers, handler)
	}
}

func (b *HL7_BASE) emit(event, message string) {
	switch event {
	case "error":
		for _, h := range b.errorHandlers {
			h(message)
		}
	case "warning":
		for _, h := range b.warningHandlers {
			h(message)
		}
	}
}

// ToMessage returns the underlying Message (the toMessage).
func (b *HL7_BASE) ToMessage() *builder.Message { return b.message }

// String returns the entire HL7 message string (the toString).
func (b *HL7_BASE) String() string { return b.message.String() }

// SetDate formats a date at the given HL7 length (the setDate). A zero date
// formats the current time.
func (b *HL7_BASE) SetDate(date time.Time, length string) string {
	if date.IsZero() {
		date = time.Now()
	}
	return utils.CreateHL7Date(date, length)
}

// headerExists panics with an HL7FatalError when the MSH header is not first
// (the headerExists). Go necessity: the spec throws; the BuildXXX shims that
// follow the `headerExists()` therefore panic.
func (b *HL7_BASE) headerExists() {
	first := b.message.GetFirstSegment()
	if first == nil || first.Name() != "MSH" {
		panic(helpers.NewHL7FatalError("MSH Header must be built first."))
	}
}

// buildMSHGuard enforces the single-MSH rule shared by every version's BuildMSH
// (the buildMSH). Go necessity: the spec throws; this panics.
func (b *HL7_BASE) buildMSHGuard() {
	if b.message.TotalSegment("MSH") > 0 {
		panic(helpers.NewHL7FatalError("You can only have one MSH Header per HL7 Message."))
	}
}

// startSegment initializes a new segment and sets it as current (the
// _startSegment).
func (b *HL7_BASE) startSegment(name string) {
	b.segment = mustAddSegment(b.message, name)
}

func mustAddSegment(m *builder.Message, name string) *builder.Segment {
	seg, err := m.AddSegment(name)
	if err != nil {
		panic(err)
	}
	return seg
}

// assertSegmentInVersion rejects building a segment that is not part of the
// current spec version (the _assertSegmentInVersion). Go necessity: panics
// with HL7ValidationError where the spec throws.
func (b *HL7_BASE) assertSegmentInVersion(spec metadata.SegmentSpec) {
	for _, v := range spec.Versions {
		if string(v) == b.version {
			return
		}
	}
	panic(helpers.NewHL7ValidationError(
		fmt.Sprintf("Segment %s is not part of HL7 v%s", spec.Name, b.version)))
}

// BuildSegment builds any HL7 segment by name from its generated SegmentSpec
// (the buildSegment). MSH must use BuildMSH. Go necessity: panics with the
// HL7 error hierarchy where the spec throws.
func (b *HL7_BASE) buildSegmentGeneric(name string, properties Props) {
	upper := strings.ToUpper(name)
	spec, ok := metadata.SEGMENT_SPECS[upper]
	if !ok {
		panic(helpers.NewHL7ValidationError(
			fmt.Sprintf("Unknown HL7 segment %q — no SegmentSpec is registered", name)))
	}
	if upper == "MSH" {
		panic(helpers.NewHL7ValidationError(
			"Use buildMSH() to build the MSH header — buildSegment does not handle MSH framing"))
	}

	b.headerExists()
	b.assertSegmentInVersion(spec)
	b.segment = mustAddSegment(b.message, spec.Name)

	lower := strings.ToLower(spec.Name)
	for _, field := range spec.Fields {
		var value any
		if v, ok := properties[fmt.Sprintf("%s_%d", lower, field.Num)]; ok {
			value = v
		} else if v, ok := properties[strconv.Itoa(field.Num)]; ok {
			value = v
		}
		b.validatorSetField(spec, field.Num, value, nil)
	}
}

// camelComponentRe drops parenthesized clarifications when camelizing a label.
var camelComponentRe = regexp.MustCompile(`\([^)]*\)`)

// camelizeComponentName converts an HL7 component label like "Zip Or Postal
// Code" into the camelCase key zipOrPostalCode (the camelizeComponentName).
func camelizeComponentName(name string) string {
	stripped := camelComponentRe.ReplaceAllString(name, "")
	tokens := strings.FieldsFunc(stripped, func(r rune) bool {
		return (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') && (r < '0' || r > '9')
	})
	if len(tokens) == 0 {
		return ""
	}
	out := strings.ToLower(tokens[0])
	for _, t := range tokens[1:] {
		out += strings.ToUpper(t[:1]) + strings.ToLower(t[1:])
	}
	return out
}

// tailKeyRe caches the per-component "_<num>$" matchers for pickComponentValue.
var tailKeyCache = map[int]*regexp.Regexp{}

// pickComponentValue resolves which key in a typed-component object holds the
// value for a ComponentSpec, trying numeric, numeric-as-string, *_<num>, and
// camelCase keys in that order (the pickComponentValue).
func pickComponentValue(object map[string]any, c metadata.ComponentSpec) any {
	if v, ok := object[strconv.Itoa(c.Num)]; ok {
		return v
	}
	re, ok := tailKeyCache[c.Num]
	if !ok {
		re = regexp.MustCompile(fmt.Sprintf(`_%d$`, c.Num))
		tailKeyCache[c.Num] = re
	}
	for k := range object {
		if re.MatchString(k) {
			return object[k]
		}
	}
	camel := camelizeComponentName(c.Name)
	if camel != "" {
		if v, ok := object[camel]; ok {
			return v
		}
	}
	return nil
}

// composeFromObject converts a typed component object into the HL7 ^-delimited
// composite string, validating each piece against its ComponentSpec (the
// _composeFromObject). Go necessity: panics with HL7ValidationError where the spec
// throws.
func (b *HL7_BASE) composeFromObject(object map[string]any, components []metadata.ComponentSpec, fieldPath string) string {
	var parts []string
	lastFilled := -1
	for _, c := range components {
		v := pickComponentValue(object, c)
		hasValue := v != nil && fmt.Sprint(v) != ""

		if (c.Usage == metadata.UsageWithdrawn || c.Usage == metadata.UsageNotSupport) && hasValue {
			label := "withdrawn"
			if c.Usage == metadata.UsageNotSupport {
				label = "not supported"
			}
			panic(helpers.NewHL7ValidationError(
				fmt.Sprintf("Component %s.%d (%s) is %s", fieldPath, c.Num, c.Name, label)))
		}
		if c.Usage == metadata.UsageRequired && !hasValue {
			panic(helpers.NewHL7ValidationError(
				fmt.Sprintf("Component %s.%d (%s) is required", fieldPath, c.Num, c.Name)))
		}
		if hasValue && c.Length.Set {
			s := fmt.Sprint(v)
			if c.Length.HasExact {
				if len(s) > c.Length.Exact {
					panic(helpers.NewHL7ValidationError(
						fmt.Sprintf("Component %s.%d (%s) must be at most %d characters", fieldPath, c.Num, c.Name, c.Length.Exact)))
				}
			} else {
				if c.Length.Max != 0 && len(s) > c.Length.Max {
					panic(helpers.NewHL7ValidationError(
						fmt.Sprintf("Component %s.%d (%s) must be at most %d characters", fieldPath, c.Num, c.Name, c.Length.Max)))
				}
				if c.Length.Min != 0 && len(s) < c.Length.Min {
					panic(helpers.NewHL7ValidationError(
						fmt.Sprintf("Component %s.%d (%s) must be at least %d characters", fieldPath, c.Num, c.Name, c.Length.Min)))
				}
			}
		}

		// Version-aware HL7 value-table enforcement at the component level: a
		// component bound to a table whose value set exists for this version must
		// carry an in-table value. An empty/absent table is not enforced.
		if hasValue && c.Table != 0 {
			if vals := b.codeTable(fmt.Sprintf("%04d", c.Table)); len(vals) > 0 && !contains(vals, fmt.Sprint(v)) {
				panic(helpers.NewHL7ValidationError(
					fmt.Sprintf("Component %s.%d (%s) must be one of: %s", fieldPath, c.Num, c.Name, strings.Join(vals, ", "))))
			}
		}

		if hasValue {
			parts = append(parts, fmt.Sprint(v))
			lastFilled = len(parts) - 1
		} else {
			parts = append(parts, "")
		}
	}
	return strings.Join(parts[:lastFilled+1], "^")
}

// findField returns the field with the given number, or false.
func findField(spec metadata.SegmentSpec, num int) (metadata.FieldSpec, bool) {
	for _, f := range spec.Fields {
		if f.Num == num {
			return f, true
		}
	}
	return metadata.FieldSpec{}, false
}

// validatorSetField is the spec-driven field setter that consults a field's
// per-version usage code and translates it into validation rules (the
// _validatorSetField). overrides may be nil. Go necessity: panics with
// HL7ValidationError where the spec throws.
func (b *HL7_BASE) validatorSetField(spec metadata.SegmentSpec, fieldNumber int, value any, overrides *ValidationRule) []string {
	field, ok := findField(spec, fieldNumber)
	if !ok {
		panic(helpers.NewHL7ValidationError(
			fmt.Sprintf("Field %s.%d is not defined in the segment spec", spec.Name, fieldNumber)))
	}

	// Composite-object input: if the caller passes a component object (a map,
	// not a Date/array/string) and the field has known components, validate and
	// assemble the ^-delimited composite here. Strings keep working.
	if obj, isObj := value.(map[string]any); isObj && len(field.Components) > 0 {
		value = b.composeFromObject(obj, field.Components, fmt.Sprintf("%s.%d", spec.Name, fieldNumber))
	}

	version := metadata.HL7Version(b.version)
	usage, usageOK := field.Usage[version]
	hasValue := value != nil && fmt.Sprint(value) != ""

	if !usageOK {
		if hasValue {
			panic(helpers.NewHL7ValidationError(
				fmt.Sprintf("Field %s.%d is not available in HL7 v%s", spec.Name, fieldNumber, b.version)))
		}
		return []string{}
	}

	if (usage == metadata.UsageWithdrawn || usage == metadata.UsageNotSupport) && hasValue {
		label := "withdrawn"
		if usage == metadata.UsageNotSupport {
			label = "not supported"
		}
		panic(helpers.NewHL7ValidationError(
			fmt.Sprintf("Field %s.%d is %s in HL7 v%s and cannot be set", spec.Name, fieldNumber, label, b.version)))
	}

	// Merge spec-derived rule with caller overrides. Caller wins for
	// length/type/allowedValues/pattern; spec-derived usage/required are
	// non-overridable.
	var rule ValidationRule
	if overrides != nil {
		rule = *overrides
	}
	if rule.AllowedValues == nil {
		rule.AllowedValues = field.AllowedValues
	}
	// Version-aware HL7 value-table enforcement: when no explicit allowed set is
	// in play and the field is bound to an HL7 table, validate against that
	// table's value set for the active version. An empty/absent table for this
	// version is not enforced (nothing to check against). Out-of-table values
	// then hit the existing hard-error AllowedValues rejection.
	if rule.AllowedValues == nil && field.Table != 0 {
		if vals := b.codeTable(fmt.Sprintf("%04d", field.Table)); len(vals) > 0 {
			rule.AllowedValues = vals
		}
	}
	if rule.DependsOn == nil {
		rule.DependsOn = field.DependsOn
	}
	if rule.HL7Type == "" {
		rule.HL7Type = field.HL7Type
	}
	if !rule.Length.Set {
		rule.Length = field.Length
	}
	rule.Deprecated = usage == metadata.UsageBackward
	rule.Required = usage == metadata.UsageRequired
	rule.Usage = usage
	rule.HasUsage = true

	// Conditional (D) fields are only enforced when the spec carries an explicit
	// dependsOn; prose-only conditions are treated as optional.
	if usage == metadata.UsageDependent && field.DependsOn != nil && hasValue {
		dep := field.DependsOn
		resolvedPath := dep.Path
		if numericPathRe.MatchString(dep.Path) {
			resolvedPath = b.segment.Name() + "." + dep.Path
		}
		depString := b.segment.Get(resolvedPath).String()
		if dep.MustBeSet && depString == "" {
			panic(helpers.NewHL7ValidationError(
				fmt.Sprintf("Field %s.%d is conditional and requires %s to be set in HL7 v%s", spec.Name, fieldNumber, dep.Path, b.version)))
		}
		if dep.HasMustEqual && depString != dep.MustEqual {
			panic(helpers.NewHL7ValidationError(
				fmt.Sprintf("Field %s.%d is conditional and requires %s to equal %q in HL7 v%s", spec.Name, fieldNumber, dep.Path, dep.MustEqual, b.version)))
		}
	}

	return b.validatorSetValue(strconv.Itoa(fieldNumber), value, &rule)
}

// numericPathRe matches a bare numeric/dotted path like "1" or "2.3".
var numericPathRe = regexp.MustCompile(`^\d+(\.\d+)*$`)

// validatorSetValue validates value against rules and writes it when clean
// (the _validatorSetValue). rules may be nil. Go necessity: panics with
// HL7ValidationError on forced/hard errors where the spec throws.
func (b *HL7_BASE) validatorSetValue(fieldPath string, value any, rules *ValidationRule) []string {
	var errors []string
	var warnings []string

	var nr ValidationRule
	if rules != nil {
		nr = *rules
	}
	if !nr.HasType {
		nr.Type = ruleString
		nr.HasType = true
	}

	if nr.AllowedValues != nil && nr.Type != ruleString {
		b.validatorThrowError(&errors, "Type must be string if 'allowedValues' is set.", false)
	}

	if rules != nil && !b.validatorIsVersionCompatible(nr.HL7Support) {
		return []string{}
	}

	normalized := validatorNormalize(value)
	b.validatorCheckDependency(&errors, nr.DependsOn, fieldPath)
	b.validatorCheckValue(&errors, fieldPath, normalized, nr)

	if nr.Deprecated && normalized != nil && fmt.Sprint(normalized) != "" {
		message := fmt.Sprintf("Field %s is deprecated and should not be used in version v%s.", fieldPath, b.version)
		if nr.UseField != "" {
			message += fmt.Sprintf(" Use '%s' instead.", nr.UseField)
		}
		b.validatorWarn(&warnings, message)
	}

	if len(errors) == 0 {
		b.segment.Set(fieldPath, normalized)
	}

	return append(errors, warnings...)
}

func (b *HL7_BASE) validatorCheckDependency(errors *[]string, dep *metadata.Depends, fieldPath string) {
	if dep == nil {
		return
	}
	resolvedPath := dep.Path
	if numericPathRe.MatchString(dep.Path) {
		resolvedPath = b.segment.Name() + "." + dep.Path
	}
	depValue := b.segment.Get(resolvedPath)
	depString := depValue.String()
	isSet := depString != ""

	if dep.MustBeSet && !isSet {
		b.validatorThrowError(errors, fmt.Sprintf("Field %s requires %s to be set", fieldPath, dep.Path), false)
	}
	if dep.HasMustEqual && depString != dep.MustEqual {
		b.validatorThrowError(errors, fmt.Sprintf("Field %s requires %s to equal %q, but got %q", fieldPath, dep.Path, dep.MustEqual, depString), false)
	}
}

var hl7DateRe = regexp.MustCompile(`^\d{8}(\d{4}(\d{2}(\.\d{1,6})?)?)?([+-]\d{4})?$`)

func (b *HL7_BASE) validatorCheckValue(errors *[]string, fieldPath string, value any, rules ValidationRule) {
	if rules.Required && (value == nil || fmt.Sprint(value) == "") {
		b.validatorThrowError(errors, fmt.Sprintf("Field %s is required", fieldPath), true)
	}

	if value == nil {
		return
	}

	if rules.Type == ruleNumber {
		s := fmt.Sprint(value)
		number, err := strconv.ParseFloat(s, 64)
		if err != nil {
			b.validatorThrowError(errors, fmt.Sprintf("Field %s must be a number", fieldPath), false)
		} else if rules.Number != nil {
			if rules.Number.HasMin && number < rules.Number.Min {
				b.validatorThrowError(errors, fmt.Sprintf("Field %s must be at least %v", fieldPath, rules.Number.Min), false)
			}
			if rules.Number.HasMax && number > rules.Number.Max {
				b.validatorThrowError(errors, fmt.Sprintf("Field %s must be at most %v", fieldPath, rules.Number.Max), false)
			}
		}
	}

	if rules.Type == ruleString {
		if _, ok := value.(string); !ok {
			b.validatorThrowError(errors, fmt.Sprintf("Field %s must be a string", fieldPath), false)
		}
	}

	if rules.Type == ruleDate {
		dateString := fmt.Sprint(value)
		if !hl7DateRe.MatchString(dateString) {
			if rules.Required {
				b.validatorThrowError(errors, fmt.Sprintf("Field %s must be a valid HL7 date (YYYYMMDD, YYYYMMDDHHMM, YYYYMMDDHHMMSS[.S+][±HHMM])", fieldPath), false)
			}
			return
		}
	}

	valueString := fmt.Sprint(value)
	length := len(valueString)

	if rules.Length.Set && rules.Length.HasExact && rules.Type != ruleDate && length != rules.Length.Exact {
		b.validatorThrowError(errors, fmt.Sprintf("Field %s must be exactly %d characters", fieldPath, rules.Length.Exact), false)
	}

	if rules.Length.Set && !rules.Length.HasExact {
		if rules.Length.Min != 0 && length < rules.Length.Min {
			b.validatorThrowError(errors, fmt.Sprintf("Field %s must be at least %d characters", fieldPath, rules.Length.Min), false)
		}
		if rules.Length.Max != 0 && length > rules.Length.Max {
			b.validatorThrowError(errors, fmt.Sprintf("Field %s must be at most %d characters", fieldPath, rules.Length.Max), false)
		}
	}

	if rules.Pattern != nil && !rules.Pattern.MatchString(valueString) {
		b.validatorThrowError(errors, fmt.Sprintf("Field %s does not match expected format", fieldPath), false)
	}

	if rules.AllowedValues != nil && !contains(rules.AllowedValues, valueString) {
		b.validatorThrowError(errors, fmt.Sprintf("Field %s must be one of: %s", fieldPath, strings.Join(rules.AllowedValues, ", ")), true)
	}
}

func contains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}

func (b *HL7_BASE) validatorIsVersionCompatible(support []string) bool {
	if len(support) == 0 {
		return true
	}
	satisfies := func(expr string) bool {
		m := versionExprRe.FindStringSubmatch(expr)
		if m == nil {
			return false
		}
		op, version := m[1], m[2]
		cmp := compareVersions(b.version, version)
		switch op {
		case "<":
			return cmp < 0
		case "<=":
			return cmp <= 0
		case "=", "==":
			return cmp == 0
		case ">":
			return cmp > 0
		case ">=":
			return cmp >= 0
		}
		return false
	}
	for _, expr := range support {
		if satisfies(expr) {
			return true
		}
	}
	return false
}

var versionExprRe = regexp.MustCompile(`(<=|>=|<|>|==?)\s*([\d.]+)`)

func compareVersions(a, b string) int {
	pa := strings.Split(a, ".")
	pb := strings.Split(b, ".")
	n := len(pa)
	if len(pb) > n {
		n = len(pb)
	}
	at := func(parts []string, i int) int {
		if i >= len(parts) {
			return 0
		}
		v, _ := strconv.Atoi(parts[i])
		return v
	}
	for i := 0; i < n; i++ {
		if d := at(pa, i) - at(pb, i); d != 0 {
			return d
		}
	}
	return 0
}

func validatorNormalize(value any) any {
	if s, ok := value.(string); ok {
		return strings.TrimSpace(s)
	}
	return value
}

// validatorThrowError records an error, throwing first when hard/forced (the
// _validatorThrowError). Go necessity: panics with HL7ValidationError.
func (b *HL7_BASE) validatorThrowError(errors *[]string, message string, forceThrow bool) {
	if b.hardError || forceThrow {
		panic(helpers.NewHL7ValidationError(message))
	}
	b.emit("error", message)
	*errors = append(*errors, message)
}

func (b *HL7_BASE) validatorWarn(warnings *[]string, message string) {
	b.emit("warning", message)
	*warnings = append(*warnings, message)
}

// timeNow returns the current time. It is a package-level seam so date-bearing
// builders read a single clock (mirrors the `new Date()`).
func timeNow() time.Time { return time.Now() }
