package utils

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
	"sort"
	"strconv"
)

// GetSegIndexes returns the byte offsets in data where any of the named
// segments begins (a segment-header name immediately followed by "|"). It
// mirrors the getSegIndexes. Offsets are returned as strings, matching
// the representation, in match order across the names.
func GetSegIndexes(names []string, data string, list []string) []string {
	if list == nil {
		list = []string{}
	}
	for _, name := range names {
		// the spec uses JS multiline anchors that also match after a line
		// terminator. Go's (?m) only treats LF as a boundary, so match the
		// header at input start or after a CR or LF, then report the captured
		// group offset.
		re := regexp.MustCompile("(?:^|[\r\n])(" + regexp.QuoteMeta(name) + ")\\|")
		for _, loc := range re.FindAllStringSubmatchIndex(data, -1) {
			list = append(list, strconv.Itoa(loc[2]))
		}
	}
	return list
}

// Split splits an HL7 stream into its constituent segments by the offsets of
// the FHS/BHS/MSH/BTS/FTS header segments. It mirrors the split.
func Split(data string, segments []string) []string {
	if segments == nil {
		segments = []string{}
	}
	idxStr := GetSegIndexes([]string{"FHS", "BHS", "MSH", "BTS", "FTS"}, data, nil)
	idx := make([]int, 0, len(idxStr))
	for _, s := range idxStr {
		n, _ := strconv.Atoi(s)
		idx = append(idx, n)
	}
	sort.Ints(idx)
	for i := 0; i < len(idx); i++ {
		start := idx[i]
		var end int
		if i+1 == len(idx) {
			end = len(data)
		} else {
			end = idx[i+1]
		}
		segments = append(segments, data[start:end])
	}
	return segments
}
