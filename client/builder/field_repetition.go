// MIT License
//
// Copyright (c) 2026 Shane
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package builder

import (
	"strconv"

	"github.com/Bugs5382/go-hl7/client/declaration"
)

// FieldRepetition is one repetition of a field (the FieldRepetition). Its
// children are components.
type FieldRepetition struct{ valueNode }

// newFieldRepetition builds a FieldRepetition with the given key and text.
func newFieldRepetition(parent node, key, text string) *FieldRepetition {
	r := &FieldRepetition{}
	r.initValueNode(r, parent, key, text, declaration.DelimiterComponent, true)
	return r
}

// Read descends into the component at the 1-based head of path (the
// FieldRepetition.read).
func (r *FieldRepetition) Read(path []string) HL7Node {
	idx, _ := strconv.Atoi(path[0])
	rest := path[1:]
	ch := r.childrenOf()
	if idx-1 < 0 || idx-1 >= len(ch) {
		return nil
	}
	component := ch[idx-1]
	if len(rest) > 0 {
		return component.Read(rest)
	}
	return component
}

// createChild builds a Component (the FieldRepetition.createChild).
func (r *FieldRepetition) createChild(text string, index int) HL7Node {
	return newComponent(r, strconv.Itoa(index+1), text)
}

// pathCore returns the parent path (the FieldRepetition.pathCore).
func (r *FieldRepetition) pathCore() []string { return r.parent.Path() }

// writeCore writes into the component at the 1-based head of path (the
// FieldRepetition.writeCore).
func (r *FieldRepetition) writeCore(path []string, value string) HL7Node {
	idx, _ := strconv.Atoi(path[0])
	return r.writeAtIndex(path[1:], value, idx-1, "")
}
