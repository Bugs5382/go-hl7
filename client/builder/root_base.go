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
	"strings"

	"github.com/Bugs5382/go-hl7/client/internal/declaration"
	"github.com/Bugs5382/go-hl7/client/utils"
)

// RootBase is the root of a tree. It holds the assembled
// delimiter string and the escape/unescape machinery.
type RootBase struct {
	nodeBase
	delimiters    string
	matchUnescape *regexp.Regexp
	matchEscape   *regexp.Regexp
}

// initRootBase initializes the embedded root from normalized options.
func (r *RootBase) initRootBase(self node, opt builderOptions) {
	r.initNodeBase(self, nil, opt.Text, declaration.DelimiterSegment, true)
	r.delimiters = opt.NewLine + opt.SeparatorField + opt.SeparatorComponent + opt.SeparatorRepetition + opt.SeparatorEscape + opt.SeparatorSubComponent
	r.matchUnescape = makeMatchUnescape(r.delimiters)
	r.matchEscape = makeMatchEscape(r.delimiters)
}

// asRoot returns the embedded RootBase for message-root traversal.
func (r *RootBase) asRoot() *RootBase { return r }

// Delimiters returns the assembled delimiter string (the delimiters getter).
func (r *RootBase) Delimiters() string { return r.delimiters }

// makeMatchUnescape builds the escape-sequence matcher for the delimiter set.
func makeMatchUnescape(delimiters string) *regexp.Regexp {
	d := []rune(delimiters)
	esc := utils.EscapeForRegExp(string(d[declaration.DelimiterEscape]))
	return regexp.MustCompile(esc + "[^" + esc + "]*" + esc)
}

// makeMatchEscape builds the matcher for the raw delimiter characters that must
// be encoded when serializing a value. It reads the characters from the
// message's own delimiter set so custom encoding characters are honored, and
// deduplicates so a delimiter set that reuses a character stays a valid pattern.
func makeMatchEscape(delimiters string) *regexp.Regexp {
	d := []rune(delimiters)
	seen := make(map[string]bool)
	var alts []string
	for _, idx := range []declaration.Delimiters{
		declaration.DelimiterEscape,
		declaration.DelimiterField,
		declaration.DelimiterRepetition,
		declaration.DelimiterComponent,
		declaration.DelimiterSubComponent,
	} {
		ch := string(d[int(idx)])
		if seen[ch] {
			continue
		}
		seen[ch] = true
		alts = append(alts, utils.EscapeForRegExp(ch))
	}
	return regexp.MustCompile(strings.Join(alts, "|"))
}

// escape encodes every embedded delimiter character in text as its HL7 escape
// sequence (\F\, \S\, \T\, \R\, \E\). It is the leaf-level (fully non-structural)
// form used when a value can carry no further sub-structure.
func (r *RootBase) escape(text string) string {
	return r.escapeForLevel(text, nil)
}

// escapeForLevel encodes the embedded delimiters in text as HL7 escape
// sequences, leaving intact the delimiters listed in structural (those that
// still subdivide the value at the level it is being written to). The escape
// character is always encoded first, in a single pass, so the sequences this
// introduces are never re-scanned or double-escaped. The characters and their
// sequences are read from the message's delimiter set, so custom encoding
// characters round-trip with unescape.
func (r *RootBase) escapeForLevel(text string, structural map[declaration.Delimiters]bool) string {
	if text == "" {
		return text
	}
	d := []rune(r.delimiters)
	esc := string(d[declaration.DelimiterEscape])
	field := string(d[declaration.DelimiterField])
	rep := string(d[declaration.DelimiterRepetition])
	comp := string(d[declaration.DelimiterComponent])
	sub := string(d[declaration.DelimiterSubComponent])

	return r.matchEscape.ReplaceAllStringFunc(text, func(match string) string {
		switch match {
		case esc:
			return esc + "E" + esc
		case field:
			if structural[declaration.DelimiterField] {
				return match
			}
			return esc + "F" + esc
		case rep:
			if structural[declaration.DelimiterRepetition] {
				return match
			}
			return esc + "R" + esc
		case comp:
			if structural[declaration.DelimiterComponent] {
				return match
			}
			return esc + "S" + esc
		case sub:
			if structural[declaration.DelimiterSubComponent] {
				return match
			}
			return esc + "T" + esc
		}
		return match
	})
}

// unescape resolves HL7 escape sequences back to their delimiter characters.
func (r *RootBase) unescape(text string) string {
	d := []rune(r.delimiters)
	escCh := string(d[declaration.DelimiterEscape])
	if !strings.Contains(text, escCh) {
		return text
	}
	return r.matchUnescape.ReplaceAllStringFunc(text, func(match string) string {
		mr := []rune(match)
		if len(mr) < 2 {
			return match
		}
		switch string(mr[1]) {
		case "C", "H", "M", "N", "Z":
			return ""
		case "E":
			return string(d[declaration.DelimiterEscape])
		case "F":
			return string(d[declaration.DelimiterField])
		case "R":
			return string(d[declaration.DelimiterRepetition])
		case "S":
			return string(d[declaration.DelimiterComponent])
		case "T":
			return string(d[declaration.DelimiterSubComponent])
		case "X":
			return utils.DecodeHexString(string(mr[2 : len(mr)-1]))
		default:
			return match
		}
	})
}
