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

	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/internal/declaration"
)

// numericPathRe matches a fully numeric dotted path like "5" or "5.1".
var numericPathRe = regexp.MustCompile(`^\d+(\.\d+)*$`)

// Segment is one segment line of a message. Its children are
// fields. MSH/BHS/FHS are special-cased: field 1 is the field separator and
// other indices shift by one.
type Segment struct {
	nodeBase
	segmentName string
}

// newSegment builds a Segment from its text. The first three characters name
// the segment (the Segment constructor).
func newSegment(parent node, text string) *Segment {
	if text == "" {
		panic(helpers.NewHL7FatalError("Segment must have a name."))
	}
	s := &Segment{}
	s.initNodeBase(s, parent, text, declaration.DelimiterField, true)
	s.segmentName = text[:min3(len(text))]
	s.nameCache = s.segmentName
	return s
}

// min3 returns the smaller of n and 3.
func min3(n int) int {
	if n < 3 {
		return n
	}
	return 3
}

// Name returns the three-letter segment name.
func (s *Segment) Name() string { return s.segmentName }

// Read resolves a field path with MSH/BHS/FHS index handling.
func (s *Segment) Read(path []string) HL7Node {
	index, _ := strconv.Atoi(path[0])
	path = path[1:]
	if index < 1 {
		panic(helpers.NewHL7FatalError("Index must be 1 or greater."))
	}
	if s.segmentName == "MSH" || s.segmentName == "BHS" || s.segmentName == "FHS" {
		root := s.messageRoot()
		if root != nil && index == 1 {
			return newSubComponent(s, "1", string([]rune(root.delimiters)[int(declaration.DelimiterField)]))
		}
		index = index - 1
	}
	ch := s.childrenOf()
	if index < 0 || index >= len(ch) {
		return nil
	}
	field := ch[index]
	if field != nil && len(path) > 0 {
		return field.Read(path)
	}
	return field
}

// Set writes value at a path, resolving numeric-only paths under the segment
// name.
func (s *Segment) Set(path string, value any) HL7Node {
	resolved := path
	if numericPathRe.MatchString(path) {
		resolved = s.segmentName + "." + path
	}
	if arr, ok := value.([]any); ok {
		for i, item := range arr {
			s.Set(resolved+"."+strconv.Itoa(i+1), item)
		}
		return s
	}
	p := s.preparePath(resolved)
	s.Write(p, s.prepareValue(value))
	return s
}

// createChild builds a Field.
func (s *Segment) createChild(text string, index int) HL7Node {
	return newField(s, strconv.Itoa(index), text)
}

// pathCore returns the segment name.
func (s *Segment) pathCore() []string { return []string{s.segmentName} }

// writeCore writes a field with MSH/BHS/FHS index handling.
func (s *Segment) writeCore(path []string, value string) HL7Node {
	index, err := strconv.Atoi(path[0])
	path = path[1:]
	if err != nil || index < 1 {
		panic(helpers.NewHL7FatalError("Can't have an index < 1 or not be a valid number."))
	}
	if (s.segmentName == "MSH" || s.segmentName == "BHS" || s.segmentName == "FHS") && index != 1 && index != 2 {
		index = index - 1
	}
	return s.writeAtIndex(path, value, index, "")
}
