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
	"strconv"

	"github.com/Bugs5382/go-hl7/client/internal/declaration"
)

// Component is one component of a field repetition (the Component). Its
// children are sub-components.
type Component struct{ valueNode }

// newComponent builds a Component with the given key and text.
func newComponent(parent node, key, text string) *Component {
	c := &Component{}
	c.initValueNode(c, parent, key, text, declaration.DelimiterSubComponent, true)
	return c
}

// Read returns the sub-component at the 1-based head of path (the
// Component.read). An out-of-range index returns nil, mirroring the
// `children[i]` yielding undefined (which get() resolves to an EmptyNode).
func (c *Component) Read(path []string) HL7Node {
	idx, _ := strconv.Atoi(path[0])
	ch := c.childrenOf()
	if idx-1 < 0 || idx-1 >= len(ch) {
		return nil
	}
	return ch[idx-1]
}

// createChild builds a SubComponent (the Component.createChild).
func (c *Component) createChild(text string, index int) HL7Node {
	return newSubComponent(c, strconv.Itoa(index+1), text)
}

// writeCore writes into the sub-component at the 1-based head of path (the
// Component.writeCore).
func (c *Component) writeCore(path []string, value string) HL7Node {
	idx, _ := strconv.Atoi(path[0])
	return c.writeAtIndex(path[1:], value, idx-1, "")
}
