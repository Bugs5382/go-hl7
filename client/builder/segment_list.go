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

import "github.com/Bugs5382/go-hl7/client/internal/declaration"

// SegmentList wraps one or more segments of the same name so callers can
// iterate even when there is a single match.
type SegmentList struct {
	nodeBase
	segments []*Segment
}

// newSegmentList builds a SegmentList over the given segments.
func newSegmentList(parent node, segments []*Segment) *SegmentList {
	l := &SegmentList{}
	l.initNodeBase(l, parent, "", declaration.DelimiterSegment, false)
	l.segments = segments
	return l
}

// Name returns the name of the first segment.
func (l *SegmentList) Name() string { return l.segments[0].Name() }

// childrenOf returns the wrapped segments.
func (l *SegmentList) childrenOf() []HL7Node {
	out := make([]HL7Node, len(l.segments))
	for i, s := range l.segments {
		out[i] = s
	}
	return out
}

// Segments returns the wrapped segments for iteration.
func (l *SegmentList) Segments() []*Segment { return l.segments }

// Read delegates to the first segment.
func (l *SegmentList) Read(path []string) HL7Node { return l.segments[0].Read(path) }

// rawText delegates to the first segment.
func (l *SegmentList) rawText() string { return l.segments[0].String() }

// pathCore returns the first segment's path.
func (l *SegmentList) pathCore() []string { return l.segments[0].Path() }

// writeCore delegates to the first segment.
func (l *SegmentList) writeCore(path []string, value string) HL7Node {
	return l.segments[0].Write(path, value)
}
