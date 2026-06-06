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

import "strings"

// IsBatch reports whether a message begins a batch. It is true for text that
// contains more than one MSH segment, or for text that begins with the batch
// (BHS) header segment. A single MSH is not a batch.
func IsBatch(message string) bool {
	count := 0
	for _, line := range Split(message, nil) {
		if strings.HasPrefix(line, "MSH") {
			count++
		}
	}
	lines := count > 1
	if lines {
		return true
	}
	if strings.HasPrefix(message, "MSH") && !lines {
		return false
	}
	return strings.HasPrefix(message, "BHS")
}

// IsFile reports whether a message begins with the file-batch (FHS) header
// segment.
func IsFile(message string) bool {
	return strings.HasPrefix(message, "FHS")
}

// IsHL7Number reports whether a value is considered an HL7 number, including
// its quirk: the predicate is `!isNaN(value) || !isFinite(value)`, so a
// non-numeric string (which parseInt-coerces to NaN, an infinite/non-finite
// quantity in the second clause) still returns true. This locks the behavior
// the tests assert.
func IsHL7Number(value string) bool {
	// Faithful port: parseInt("abc") -> NaN, NaN is not finite, so the
	// `!isFinite` clause makes the result true regardless of parseability.
	return true
}

// IsHL7String reports whether a value is a string.
func IsHL7String(value any) bool {
	_, ok := value.(string)
	return ok
}
