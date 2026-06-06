package builder

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
	"strconv"
	"time"

	"github.com/Bugs5382/go-hl7/client/internal/declaration"
)

// valueNode carries a keyed leaf value. It overrides the
// value coercions and String, and provides a key-based pathCore.
type valueNode struct {
	nodeBase
	key string
}

// initValueNode initializes the embedded value node.
func (v *valueNode) initValueNode(self node, parent node, key, text string, delimiter declaration.Delimiters, hasDelimiter bool) {
	v.initNodeBase(self, parent, text, delimiter, hasDelimiter)
	v.key = key
}

// Bool coerces "Y"/"N"; ok is false otherwise.
func (v *valueNode) Bool() (bool, bool) {
	switch v.self.String() {
	case "N":
		return false, true
	case "Y":
		return true, true
	}
	return false, false
}

// tzRe matches the trailing timezone offset of an HL7 date/time.
var tzRe = regexp.MustCompile(`([+-])(\d{2})(\d{2})$`)

// Date coerces an HL7 date/time to a time.Time. ok is false
// when the leading YYYYMMDD cannot be parsed.
func (v *valueNode) Date() (time.Time, bool) {
	text := v.self.String()
	atoi := func(s string) int { n, _ := strconv.Atoi(s); return n }
	slice := func(a, b int) string {
		if a > len(text) {
			return ""
		}
		if b > len(text) {
			b = len(text)
		}
		return text[a:b]
	}
	if len(text) < 8 {
		return time.Time{}, false
	}
	year := atoi(slice(0, 4))
	month := atoi(slice(4, 6))
	day := atoi(slice(6, 8))
	hourStr := slice(8, 10)
	minStr := slice(10, 12)
	secStr := slice(12, 14)
	hour, minute, second := atoi(hourStr), atoi(minStr), atoi(secStr)
	if year == 0 && month == 0 && day == 0 {
		return time.Time{}, false
	}

	if m := tzRe.FindStringSubmatch(text); m != nil {
		base := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
		sign := -1
		if m[1] != "+" {
			sign = 1
		}
		tzHour, _ := strconv.Atoi(m[2])
		tzMin, _ := strconv.Atoi(m[3])
		offset := time.Duration(sign*((tzHour*60+tzMin)*60)) * time.Second
		return base.Add(offset), true
	}
	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.Local), true
}

// Float coerces to a float. ok is false when unparseable.
func (v *valueNode) Float() (float64, bool) {
	f, err := strconv.ParseFloat(v.self.String(), 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

// Int coerces to an integer. ok is false when unparseable.
func (v *valueNode) Int() (int, bool) {
	n, err := parseLeadingInt(v.self.String())
	if !err {
		return 0, false
	}
	return n, true
}

// rawText returns the de-framed text: the raw value, or the first child's text
// for a composite.
func (v *valueNode) rawText() string {
	ch := v.self.childrenOf()
	if len(ch) == 0 {
		return v.self.Raw()
	}
	return ch[0].String()
}

// pathCore computes the path as the parent path plus this the key.
func (v *valueNode) pathCore() []string {
	return append(append([]string{}, v.parent.Path()...), v.key)
}

// parseLeadingInt parses a leading signed integer (returns ok=false when no
// digits lead).
func parseLeadingInt(s string) (int, bool) {
	i := 0
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		i++
	}
	start := i
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == start {
		return 0, false
	}
	n, err := strconv.Atoi(s[:i])
	if err != nil {
		return 0, false
	}
	return n, true
}
