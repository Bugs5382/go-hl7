// Package modules carries the MLLP codec, parser plan, and option/handler type
// aliases.
package modules

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

// ParserPlan derives the five delimiter characters from the MSH/BHS/FHS header
// bytes, applying defaults when the input is shorter than expected.
type ParserPlan struct {
	// SeparatorField is the field separator (default "|").
	SeparatorField string
	// SeparatorComponent is the component separator (default "^").
	SeparatorComponent string
	// SeparatorRepetition is the repetition separator (default "~").
	SeparatorRepetition string
	// SeparatorEscape is the escape character (default "\\").
	SeparatorEscape string
	// SeparatorSubComponent is the sub-component separator (default "&").
	SeparatorSubComponent string
}

// NewParserPlan derives a ParserPlan from the delimiter run that follows the
// segment name (the bytes at offsets 3..8).
func NewParserPlan(data string) *ParserPlan {
	seps := []rune(data)
	p := &ParserPlan{
		SeparatorRepetition:   "~",
		SeparatorComponent:    "^",
		SeparatorSubComponent: "&",
		SeparatorEscape:       "\\",
	}
	if len(seps) > 0 {
		p.SeparatorField = string(seps[0])
	}
	if len(seps) > 2 {
		p.SeparatorRepetition = string(seps[2])
	}
	if len(seps) > 1 {
		p.SeparatorComponent = string(seps[1])
	}
	if len(seps) > 4 {
		p.SeparatorSubComponent = string(seps[4])
	}
	if len(seps) > 3 {
		p.SeparatorEscape = string(seps[3])
	}
	return p
}
