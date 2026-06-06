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

package utils

import (
	"math"
	"math/rand"
)

// ExpBackoff calculates an exponential backoff/retry delay (where attempts >= 1
// and exp > 1). It mirrors the expBackoff. The returned delay falls in
// [step, max] where max grows with the attempt count up to high. Pass exp = 2
// for the default exponent base.
func ExpBackoff(step, high, attempts, exp int) int {
	slots := int(math.Ceil(math.Min(float64(high)/float64(step), math.Pow(float64(exp), float64(attempts)))))
	maxDelay := slots * step
	if maxDelay > high {
		maxDelay = high
	}
	return int(math.Floor(rand.Float64()*float64(maxDelay-step) + float64(step)))
}
