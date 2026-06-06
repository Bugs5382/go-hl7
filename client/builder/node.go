// Package builder is the wire message model (delimiter-level, version-agnostic)
// for HL7 v2 messages: the Message, Batch, FileBatch roots and the uniform node
// hierarchy (NodeBase, RootBase, Segment, SegmentList, Field, FieldRepetition,
// Component, SubComponent, ValueNode, EmptyNode).
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

import "time"

// HL7Node is the uniform interface every element of the HL7 tree satisfies.
// Lookup splits into Get(path string) and Index(i int); the value coercions
// return (T, ok) rather than panicking on bad input.
type HL7Node interface {
	// Name returns the dotted path name of this node.
	Name() string
	// Len returns the number of child nodes.
	Len() int

	// Get resolves a string path like "PID.5.1" (1-based HL7 indices).
	Get(path string) HL7Node
	// Index returns the child at a 0-based position.
	Index(i int) HL7Node

	// Set writes value at a string path, returning the receiver for chaining.
	Set(path string, value any) HL7Node
	// SetIndex writes value at a 0-based child index.
	SetIndex(i int, value any) HL7Node
	// Exists reports whether a string path resolves to a non-empty node.
	Exists(path string) bool
	// ExistsIndex reports whether a child index resolves to a non-empty node.
	ExistsIndex(i int) bool

	// Read resolves an already-split path.
	Read(path []string) HL7Node
	// Write writes value at an already-split path.
	Write(path []string, value string) HL7Node

	// String returns the de-framed HL7 text for this node.
	String() string
	// Raw returns the raw text for this node.
	Raw() string
	// ToArray returns the child nodes.
	ToArray() []HL7Node
	// ForEach iterates over the children.
	ForEach(cb func(value HL7Node, index int))
	// IsEmpty reports whether the node has no children.
	IsEmpty() bool

	// Path returns the path segments.
	Path() []string

	// Int coerces to an integer; ok is false when the value is unparseable.
	Int() (int, bool)
	// Float coerces to a float; ok is false when unparseable.
	Float() (float64, bool)
	// Bool coerces "Y"/"N" to a boolean; ok is false otherwise.
	Bool() (bool, bool)
	// Date coerces an HL7 date/time to a time.Time; ok is false when invalid.
	Date() (time.Time, bool)
}

// node is the internal virtual-method contract used for inheritance-style
// dispatch. Each subclass overrides the hooks (createChild, pathCore,
// writeCore, read, and the string/name/children accessors); the interface plus
// a self back-reference on base lets a base method that needs the
// "most-derived" behavior call through self.
type node interface {
	HL7Node

	// createChild builds a child node for text at the given index.
	createChild(text string, index int) HL7Node
	// pathCore computes this the path segments.
	pathCore() []string
	// writeCore performs the type-specific write for an already-split path.
	writeCore(path []string, value string) HL7Node
	// childrenOf returns the (possibly lazily-parsed) child slice.
	childrenOf() []HL7Node
	// rawText returns the de-framed text via the type's String semantics.
	rawText() string
}
