package hl7

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

import "github.com/Bugs5382/go-hl7/client/hl7/tables"

// codeTable returns the values of a generated HL7 value table for a version and
// id (e.g. version "2.8", id "0076"), or nil when the table carries no fixed
// value set for that version. HL7 table values differ by version, so lookup is
// version-aware; this is the Go bridge to the generated tables registry.
func codeTable(version, id string) []string {
	if v, ok := tables.Tables[version]; ok {
		return v[id]
	}
	return nil
}

// codeTable returns the value table for an id resolved against the builder's
// active HL7 version. Builders use this so a single shared segment builder stays
// version-correct.
func (b *Builder) codeTable(id string) []string {
	return codeTable(b.version, id)
}
