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

import "github.com/Bugs5382/go-hl7/client/declaration"

// Field is one field of a segment (the Field). Its children are field
// repetitions.
type Field struct{ valueNode }

// newField builds a Field with the given key and text.
func newField(parent node, key, text string) *Field {
	f := &Field{}
	f.initValueNode(f, parent, key, text, declaration.DelimiterRepetition, true)
	return f
}

// Read descends into the first repetition (the Field.read).
func (f *Field) Read(path []string) HL7Node {
	ch := f.childrenOf()
	if len(ch) > 0 {
		return ch[0].Read(path)
	}
	return f
}

// createChild builds a FieldRepetition (the Field.createChild).
func (f *Field) createChild(text string, index int) HL7Node {
	return newFieldRepetition(f, f.key, text)
}

// writeCore writes into the first repetition (the Field.writeCore).
func (f *Field) writeCore(path []string, value string) HL7Node {
	return f.ensureChild().Write(path, value)
}

// ensureChild returns the first repetition, creating it when absent.
func (f *Field) ensureChild() HL7Node {
	if len(f.childrenOf()) == 0 {
		child := f.createChild("", 0)
		f.setChild(child, 0)
		return child
	}
	return f.childrenOf()[0]
}
