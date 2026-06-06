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

// EmptyNode is the shared sentinel returned for a missing path, so chained
// Get(...).String() never panics. It mirrors the EmptyNode: reads
// self-return, value coercions report not-ok (throws), and writes are
// no-ops returning the node.
type EmptyNode struct{}

// Len returns 0.
func (e *EmptyNode) Len() int { return 0 }

// Name is not implemented for an empty node, mirroring the throw.
func (e *EmptyNode) Name() string { panic("Method not implemented") }

// Path is not implemented for an empty node, mirroring the throw.
func (e *EmptyNode) Path() []string { panic("Method not implemented") }

// Exists always reports false.
func (e *EmptyNode) Exists(path string) bool { return false }

// ExistsIndex always reports false.
func (e *EmptyNode) ExistsIndex(i int) bool { return false }

// ForEach is not implemented for an empty node, mirroring the throw.
func (e *EmptyNode) ForEach(cb func(value HL7Node, index int)) { panic("Method not implemented") }

// Get self-returns the empty node.
func (e *EmptyNode) Get(path string) HL7Node { return e }

// Index self-returns the empty node.
func (e *EmptyNode) Index(i int) HL7Node { return e }

// IsEmpty always reports true.
func (e *EmptyNode) IsEmpty() bool { return true }

// Read is not implemented for an empty node, mirroring the throw.
func (e *EmptyNode) Read(path []string) HL7Node { panic("Method not implemented") }

// Set self-returns the empty node.
func (e *EmptyNode) Set(path string, value any) HL7Node { return e }

// SetIndex self-returns the empty node.
func (e *EmptyNode) SetIndex(i int, value any) HL7Node { return e }

// Write self-returns the empty node.
func (e *EmptyNode) Write(path []string, value string) HL7Node { return e }

// ToArray returns an empty slice.
func (e *EmptyNode) ToArray() []HL7Node { return []HL7Node{} }

// Raw is not implemented for an empty node, mirroring the throw.
func (e *EmptyNode) Raw() string { panic("Method not implemented") }

// String returns the empty string.
func (e *EmptyNode) String() string { return "" }

// Int reports not-ok (throws).
func (e *EmptyNode) Int() (int, bool) { return 0, false }

// Float reports not-ok (throws).
func (e *EmptyNode) Float() (float64, bool) { return 0, false }

// Bool reports not-ok (throws).
func (e *EmptyNode) Bool() (bool, bool) { return false, false }

// Date reports not-ok (throws).
func (e *EmptyNode) Date() (time.Time, bool) { return time.Time{}, false }
