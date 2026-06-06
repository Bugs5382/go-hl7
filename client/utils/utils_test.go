// Mirrors __tests__/client/hl7.utils.test.ts.
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
	"testing"
	"time"
)

func TestClientUtilityFunctions(t *testing.T) {
	// Fixed date so format checks do not depend on the runner's clock. Feb 3,
	// 2024 04:05:06.078, constructed in the local location to match the
	// local-time getters.
	fixed := time.Date(2024, 2, 3, 4, 5, 6, 78*int(time.Millisecond), time.Local)

	t.Run("createHL7Date / padHL7Date", func(t *testing.T) {
		t.Run("default (no length) to 14-char YYYYMMDDHHMMSS", func(t *testing.T) {
			if got := CreateHL7Date(fixed, ""); got != "20240203040506" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("length 8 to YYYYMMDD", func(t *testing.T) {
			if got := CreateHL7Date(fixed, "8"); got != "20240203" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("length 12 to YYYYMMDDHHMM", func(t *testing.T) {
			if got := CreateHL7Date(fixed, "12"); got != "202402030405" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("length 14 to YYYYMMDDHHMMSS", func(t *testing.T) {
			if got := CreateHL7Date(fixed, "14"); got != "20240203040506" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("length 19 to YYYYMMDDHHMMSS.SSSS", func(t *testing.T) {
			if got := CreateHL7Date(fixed, "19"); got != "20240203040506.0078" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("length 24 includes timezone offset", func(t *testing.T) {
			got := CreateHL7Date(fixed, "24")
			if !regexp.MustCompile(`^20240203040506\.0078[+-]\d{4}$`).MatchString(got) {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("length 26 microseconds + timezone", func(t *testing.T) {
			got := CreateHL7Date(fixed, "26")
			if !regexp.MustCompile(`^20240203040506\.078000[+-]\d{4}$`).MatchString(got) {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("padHL7Date pads numbers to fixed width with zeros", func(t *testing.T) {
			cases := []struct {
				n, w int
				want string
			}{{0, 2, "00"}, {7, 2, "07"}, {42, 2, "42"}, {123, 2, "123"}, {5, 4, "0005"}}
			for _, c := range cases {
				if got := PadHL7Date(c.n, c.w, "0"); got != c.want {
					t.Fatalf("PadHL7Date(%d,%d)=%q want %q", c.n, c.w, got, c.want)
				}
			}
		})
		t.Run("padHL7Date accepts a custom pad character", func(t *testing.T) {
			if got := PadHL7Date(7, 4, "x"); got != "xxx7" {
				t.Fatalf("got %q", got)
			}
		})
	})

	t.Run("isHL7Number / isHL7String", func(t *testing.T) {
		t.Run("isHL7Number recognizes numeric strings and numbers", func(t *testing.T) {
			for _, v := range []string{"42", "0"} {
				if !IsHL7Number(v) {
					t.Fatalf("IsHL7Number(%q) = false", v)
				}
			}
		})
		t.Run("isHL7Number returns true for any input the predicate accepts", func(t *testing.T) {
			if !IsHL7Number("abc") {
				t.Fatalf("IsHL7Number(abc) = false")
			}
		})
		t.Run("isHL7String recognizes strings only", func(t *testing.T) {
			if !IsHL7String("hello") || !IsHL7String("") {
				t.Fatalf("string inputs should be true")
			}
			if IsHL7String(42) || IsHL7String(nil) {
				t.Fatalf("non-string inputs should be false")
			}
		})
	})

	t.Run("isBatch / isFile", func(t *testing.T) {
		t.Run("isFile detects FHS prefix", func(t *testing.T) {
			if !IsFile(`FHS|^~\&|...`) {
				t.Fatal("FHS should be file")
			}
			if IsFile(`MSH|^~\&|...`) || IsFile(`BHS|^~\&|...`) {
				t.Fatal("MSH/BHS should not be file")
			}
		})
		t.Run("isBatch is true for multi-MSH text without BHS", func(t *testing.T) {
			multi := "MSH|^~\\&|||||20240101||ADT^A01|1||2.7\rMSH|^~\\&|||||20240101||ADT^A01|2||2.7"
			if !IsBatch(multi) {
				t.Fatal("multi-MSH should be batch")
			}
		})
		t.Run("isBatch is true for explicit BHS", func(t *testing.T) {
			if !IsBatch(`BHS|^~\&|...`) {
				t.Fatal("BHS should be batch")
			}
		})
		t.Run("isBatch is false for a single MSH", func(t *testing.T) {
			if IsBatch(`MSH|^~\&|||||20240101||ADT^A01|1||2.7`) {
				t.Fatal("single MSH should not be batch")
			}
		})
	})

	t.Run("expBackoff", func(t *testing.T) {
		t.Run("delay is within [step, high]", func(t *testing.T) {
			for attempts := 1; attempts <= 6; attempts++ {
				delay := ExpBackoff(1000, 30000, attempts, 2)
				if delay < 1000 || delay > 30000 {
					t.Fatalf("attempts %d: delay %d out of range", attempts, delay)
				}
			}
		})
		t.Run("delay scales with attempt count (statistically)", func(t *testing.T) {
			sample := func(attempts int) float64 {
				total := 0
				for i := 0; i < 200; i++ {
					total += ExpBackoff(1000, 30000, attempts, 2)
				}
				return float64(total) / 200
			}
			if sample(6) <= sample(1) {
				t.Fatal("higher attempts should average larger delay")
			}
		})
		t.Run("custom exp base is honored", func(t *testing.T) {
			delay := ExpBackoff(500, 5000, 10, 1)
			if delay < 500 || delay > 5000 {
				t.Fatalf("delay %d out of range", delay)
			}
		})
	})

	t.Run("escapeForRegExp", func(t *testing.T) {
		t.Run("escapes regex metacharacters", func(t *testing.T) {
			if got := EscapeForRegExp("a.b*c+d?"); got != `a\.b\*c\+d\?` {
				t.Fatalf("got %q", got)
			}
			if got := EscapeForRegExp("[](){}|^$"); got != `\[\]\(\)\{\}\|\^\$` {
				t.Fatalf("got %q", got)
			}
			if got := EscapeForRegExp("plain"); got != "plain" {
				t.Fatalf("got %q", got)
			}
		})
	})

	t.Run("decodeHexString", func(t *testing.T) {
		t.Run("decodes ASCII hex to characters", func(t *testing.T) {
			if got := DecodeHexString("4869"); got != "Hi" {
				t.Fatalf("got %q", got)
			}
			if got := DecodeHexString("484c37"); got != "HL7" {
				t.Fatalf("got %q", got)
			}
		})
		t.Run("empty string returns empty string", func(t *testing.T) {
			if got := DecodeHexString(""); got != "" {
				t.Fatalf("got %q", got)
			}
		})
	})

	t.Run("randomString", func(t *testing.T) {
		t.Run("returns the requested length", func(t *testing.T) {
			if len(RandomString(20)) != 20 {
				t.Fatal("default length 20")
			}
			if len(RandomString(1)) != 1 {
				t.Fatal("length 1")
			}
			if len(RandomString(50)) != 50 {
				t.Fatal("length 50")
			}
		})
		t.Run("only contains the allowed alphabet", func(t *testing.T) {
			alphabet := regexp.MustCompile(`^[ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0-9_]+$`)
			if !alphabet.MatchString(RandomString(100)) {
				t.Fatal("random string contains a disallowed character")
			}
		})
	})
}
