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
	"strings"
	"time"

	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/internal/declaration"
	"github.com/Bugs5382/go-hl7/client/utils"
)

// emptySingleton is the shared sentinel returned for a missing path, so chained
// Get(...).String() never panics.
var emptySingleton = &EmptyNode{}

// nodeBase carries the shared behavior of every tree node. The self field holds
// the most-derived node so base methods dispatch overridden behavior (Go's
// stand-in for virtual methods).
type nodeBase struct {
	// self is the most-derived node, set by each constructor for virtual
	// dispatch: base methods call self to reach overrides.
	self node
	// parent is the owning node, or nil for a root.
	parent node

	children      []HL7Node
	delimiter     declaration.Delimiters
	hasDelimiter  bool
	delimiterText string
	dirty         bool
	nameCache     string
	pathCache     []string
	hasPathCache  bool
	text          string
	root          *RootBase
}

// initNodeBase initializes the embedded base. parent may be nil (a root).
func (n *nodeBase) initNodeBase(self node, parent node, text string, delimiter declaration.Delimiters, hasDelimiter bool) {
	n.self = self
	n.parent = parent
	n.children = nil
	n.delimiter = delimiter
	n.hasDelimiter = hasDelimiter
	n.delimiterText = ""
	n.dirty = false
	n.nameCache = ""
	n.text = text
}

// messageRoot walks to the owning RootBase (the message getter).
func (n *nodeBase) messageRoot() *RootBase {
	if n.root != nil {
		return n.root
	}
	if n.parent == nil {
		if rb, ok := n.self.(interface{ asRoot() *RootBase }); ok {
			n.root = rb.asRoot()
		}
	} else if pb, ok := n.parent.(interface{ messageRootPublic() *RootBase }); ok {
		n.root = pb.messageRootPublic()
	}
	return n.root
}

// messageRootPublic exposes messageRoot for parent traversal across types.
func (n *nodeBase) messageRootPublic() *RootBase { return n.messageRoot() }

// delimiterTextOf resolves the delimiter character for this the level.
func (n *nodeBase) delimiterTextOf() string {
	root := n.messageRoot()
	if root == nil || !n.hasDelimiter {
		panic(helpers.NewHL7FatalError("this.message is not defined."))
	}
	n.delimiterText = string([]rune(root.delimiters)[int(n.delimiter)])
	return n.delimiterText
}

// childrenOf lazily parses and returns the child nodes (the children getter).
func (n *nodeBase) childrenOf() []HL7Node {
	if n.text != "" && len(n.children) == 0 {
		parts := strings.Split(n.text, n.delimiterTextOf())
		ch := make([]HL7Node, len(parts))
		for i, p := range parts {
			ch[i] = n.self.createChild(p, i)
		}
		n.children = ch
	}
	return n.children
}

// Len returns the number of children (the length getter).
func (n *nodeBase) Len() int { return len(n.self.childrenOf()) }

// Name returns the dotted path name (the name getter).
func (n *nodeBase) Name() string {
	if n.nameCache != "" {
		return n.nameCache
	}
	n.nameCache = strings.Join(n.self.Path(), ".")
	return n.nameCache
}

// Path returns the path segments (the path getter).
func (n *nodeBase) Path() []string {
	if n.hasPathCache {
		return n.pathCache
	}
	n.pathCache = n.self.pathCore()
	n.hasPathCache = true
	return n.pathCache
}

// Exists reports whether a string path resolves to a non-empty node.
func (n *nodeBase) Exists(path string) bool {
	v := n.self.Get(path)
	if v == nil {
		return false
	}
	return !v.IsEmpty()
}

// ExistsIndex reports whether a child index resolves to a non-empty node.
func (n *nodeBase) ExistsIndex(i int) bool {
	v := n.self.Index(i)
	if v == nil {
		return false
	}
	return !v.IsEmpty()
}

// ForEach iterates over the children.
func (n *nodeBase) ForEach(cb func(value HL7Node, index int)) {
	ch := n.self.childrenOf()
	for i, c := range ch {
		cb(c, i)
	}
}

// Index returns the child at a 0-based position (the get(number)).
func (n *nodeBase) Index(i int) HL7Node {
	ch := n.self.childrenOf()
	if i >= 0 && i < len(ch) {
		return ch[i]
	}
	return emptySingleton
}

// Get resolves a string path like "PID.5.1" (the get(string)).
func (n *nodeBase) Get(path string) HL7Node {
	if path == "" {
		return emptySingleton
	}
	p := n.preparePath(path)
	rv := n.self.Read(p)
	if rv == nil {
		return emptySingleton
	}
	return rv
}

// IsEmpty reports whether the node has no children.
func (n *nodeBase) IsEmpty() bool { return len(n.self.childrenOf()) == 0 }

// Read resolves an already-split path. The base implementation is not
// implemented; concrete nodes override it.
func (n *nodeBase) Read(path []string) HL7Node {
	panic(helpers.NewHL7FatalError("Method not implemented."))
}

// Set writes value at a string path (the set(string, value)).
func (n *nodeBase) Set(path string, value any) HL7Node {
	if arr, ok := value.([]any); ok {
		for i, item := range arr {
			n.self.Set(path+"."+strconv.Itoa(i+1), item)
		}
		return n.self
	}
	p := n.preparePath(path)
	n.self.Write(p, n.prepareValue(value))
	return n.self
}

// SetIndex writes value at a 0-based child index (the set(number, value)).
func (n *nodeBase) SetIndex(i int, value any) HL7Node {
	if arr, ok := value.([]any); ok {
		child := n.ensureIndex(i)
		for j, item := range arr {
			child.SetIndex(j, item)
		}
		return n.self
	}
	n.setChild(n.self.createChild(n.prepareValue(value), i), i)
	return n.self
}

// Write writes value at an already-split path.
func (n *nodeBase) Write(path []string, value string) HL7Node {
	n.setDirty()
	if value == "" {
		return n.self.writeCore(path, "")
	}
	return n.self.writeCore(path, value)
}

// ToArray returns the child nodes.
func (n *nodeBase) ToArray() []HL7Node { return n.self.childrenOf() }

// Raw renders the raw text.
func (n *nodeBase) Raw() string {
	if !n.dirty {
		return n.text
	}
	n.dirty = false
	parts := make([]string, 0, len(n.self.childrenOf()))
	for _, c := range n.self.childrenOf() {
		parts = append(parts, c.Raw())
	}
	n.text = strings.Join(parts, n.delimiterTextOf())
	return n.text
}

// String renders the de-framed text. The base delegates to rawText.
func (n *nodeBase) String() string { return n.self.rawText() }

// rawText is the default String backing, returning the raw text.
func (n *nodeBase) rawText() string { return n.self.Raw() }

// Int coerces to an integer; not implemented on the base (ValueNode overrides).
func (n *nodeBase) Int() (int, bool) { return 0, false }

// Float coerces to a float; not implemented on the base.
func (n *nodeBase) Float() (float64, bool) { return 0, false }

// Bool coerces to a boolean; not implemented on the base.
func (n *nodeBase) Bool() (bool, bool) { return false, false }

// Date coerces to a time.Time; not implemented on the base.
func (n *nodeBase) Date() (time.Time, bool) { return time.Time{}, false }

// addChild appends a child built from text.
func (n *nodeBase) addChild(text string) HL7Node {
	n.setDirty()
	child := n.self.createChild(text, len(n.self.childrenOf()))
	n.children = append(n.self.childrenOf(), child)
	return child
}

// createChild builds a child node; concrete nodes override it.
func (n *nodeBase) createChild(text string, index int) HL7Node {
	panic(helpers.NewHL7FatalError("Method not implemented."))
}

// ensureIndex resolves a child index, creating it when absent.
func (n *nodeBase) ensureIndex(i int) HL7Node {
	rv := n.self.Index(i)
	if rv != HL7Node(emptySingleton) {
		return rv
	}
	return n.setChild(n.self.createChild("", i), i)
}

// pathCore computes the path; concrete nodes override it.
func (n *nodeBase) pathCore() []string {
	panic(helpers.NewHL7FatalError("Method not implemented."))
}

// preparePath splits and validates a path relative to this node.
func (n *nodeBase) preparePath(path string) []string {
	parts := strings.Split(path, ".")
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
		parts = append(append([]string{}, n.self.Path()...), parts...)
	}
	if !n.isSubPath(parts) {
		panic(helpers.NewHL7FatalError("'" + strings.Join(parts, ",") + "' is not a sub-path of '" + strings.Join(n.self.Path(), ",") + "'"))
	}
	return n.remainderOf(parts)
}

// prepareValue coerces a Go value to its HL7 string form.
func (n *nodeBase) prepareValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		if root := n.messageRoot(); root != nil {
			return root.escape(v)
		}
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return "Y"
		}
		return "N"
	case time.Time:
		return formatDateTime(v)
	default:
		return ""
	}
}

// setChild places a child at index, growing with empties as needed.
func (n *nodeBase) setChild(child HL7Node, index int) HL7Node {
	n.setDirty()
	ch := n.self.childrenOf()
	if index < len(ch) {
		ch[index] = child
		n.children = ch
		return child
	}
	for i := len(ch); i < index; i++ {
		ch = append(ch, n.self.createChild("", i))
	}
	ch = append(ch, child)
	n.children = ch
	return child
}

// setDirty marks the node and its ancestors dirty.
func (n *nodeBase) setDirty() {
	if !n.dirty {
		n.dirty = true
		if n.parent != nil {
			if pb, ok := n.parent.(interface{ setDirtyPublic() }); ok {
				pb.setDirtyPublic()
			}
		}
	}
}

// setDirtyPublic exposes setDirty for parent traversal across types.
func (n *nodeBase) setDirtyPublic() { n.setDirty() }

// writeAtIndex writes value into the child at index, descending the remaining
// path.
func (n *nodeBase) writeAtIndex(path []string, value string, index int, emptyValue string) HL7Node {
	var child HL7Node
	ch := n.self.childrenOf()
	if len(path) == 0 {
		v := value
		if value == "" {
			v = emptyValue
		}
		child = n.self.createChild(v, index)
	} else if index < len(ch) {
		child = ch[index]
	} else {
		child = n.self.createChild(emptyValue, index)
	}

	n.setChild(child, index)

	if len(path) > 0 {
		return child.Write(path, value)
	}
	return child
}

// writeCore performs the type-specific write; concrete nodes override it.
func (n *nodeBase) writeCore(path []string, value string) HL7Node {
	panic(helpers.NewHL7FatalError("Method not implemented."))
}

// isSubPath reports whether other extends this the path.
func (n *nodeBase) isSubPath(other []string) bool {
	p := n.self.Path()
	if len(p) >= len(other) {
		return false
	}
	for i := range p {
		if p[i] != other[i] {
			return false
		}
	}
	return true
}

// remainderOf returns the portion of other beyond this the path.
func (n *nodeBase) remainderOf(other []string) []string {
	return other[len(n.self.Path()):]
}

// formatDate renders a Date as YYYYMMDD.
func formatDate(date time.Time) string {
	return strconv.Itoa(date.Year()) + utils.PadHL7Date(int(date.Month()), 2, "0") + utils.PadHL7Date(date.Day(), 2, "0")
}

// formatDateTime renders a Date, adding the time when present.
func formatDateTime(date time.Time) string {
	if date.Hour() != 0 || date.Minute() != 0 || date.Second() != 0 || date.Nanosecond() != 0 {
		return formatDate(date) + utils.PadHL7Date(date.Hour(), 2, "0") + utils.PadHL7Date(date.Minute(), 2, "0") + utils.PadHL7Date(date.Second(), 2, "0")
	}
	return formatDate(date)
}
