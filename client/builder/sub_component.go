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
	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/internal/declaration"
)

// SubComponent is the leaf value of the tree (the SubComponent). It
// unescapes its raw text on String and treats a non-string as empty.
type SubComponent struct{ valueNode }

// newSubComponent builds a SubComponent with the given key and text.
func newSubComponent(parent node, key, text string) *SubComponent {
	s := &SubComponent{}
	// the spec passes no delimiter for a SubComponent; mirror that by leaving the
	// delimiter unset so it has no further children to split.
	s.initValueNode(s, parent, key, text, declaration.DelimiterSegment, false)
	return s
}

// IsEmpty reports whether the unescaped value is not a string (the
// SubComponent.isEmpty). In Go String always yields a string, so this is false.
func (s *SubComponent) IsEmpty() bool { return false }

// rawText returns the unescaped raw text (the SubComponent.toString).
func (s *SubComponent) rawText() string {
	root := s.messageRoot()
	if root == nil {
		panic(helpers.NewHL7FatalError("this.message is undefined. Unable to continue."))
	}
	return root.unescape(s.Raw())
}

// childrenOf returns no children: a SubComponent has no delimiter to split on.
func (s *SubComponent) childrenOf() []HL7Node { return nil }
