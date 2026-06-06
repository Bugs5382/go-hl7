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
	"math/rand"
	"strconv"
	"strings"
)

// randomStringAlphabet is the character set RandomString draws from.
const randomStringAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_"

// RandomString generates a random string of the given length from the
// alphanumeric-plus-underscore alphabet. The usual default length is 20.
func RandomString(length int) string {
	var b strings.Builder
	b.Grow(length)
	for i := 0; i < length; i++ {
		b.WriteByte(randomStringAlphabet[rand.Intn(len(randomStringAlphabet))])
	}
	return b.String()
}

// regexpMeta is the set of regular-expression metacharacters EscapeForRegExp
// escapes.
const regexpMeta = "-/\\^$*+?.()|[]{}"

// EscapeForRegExp escapes regular-expression metacharacters in value by
// prefixing each with a backslash.
func EscapeForRegExp(value string) string {
	var b strings.Builder
	for _, r := range value {
		if strings.ContainsRune(regexpMeta, r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// DecodeHexString decodes a string of ASCII hex digit pairs into the characters
// they encode.
func DecodeHexString(value string) string {
	var b strings.Builder
	for i := 0; i+1 < len(value)+1 && i+2 <= len(value); i += 2 {
		n, err := strconv.ParseInt(value[i:i+2], 16, 32)
		if err != nil {
			continue
		}
		b.WriteRune(rune(n))
	}
	return b.String()
}
