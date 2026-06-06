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

import "testing"

// Mirrors hl7.dualstack.test.ts "validIPv4 / validIPv6 / detectIPFamily".

func TestValidIPv4(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "0.0.0.0", "192.0.2.1", "255.255.255.255"} {
		if !ValidIPv4(ip) {
			t.Errorf("ValidIPv4(%q) = false, want true", ip)
		}
	}
	for _, ip := range []string{"256.0.0.1", "1.2.3", "1.2.3.4.5", "a.b.c.d", "", "::1"} {
		if ValidIPv4(ip) {
			t.Errorf("ValidIPv4(%q) = true, want false", ip)
		}
	}
}

func TestValidIPv6(t *testing.T) {
	valid := []string{
		"2001:0db8:85a3:0000:0000:8a2e:0370:7334",
		"fe80:0000:0000:0000:0000:0000:0000:0001",
		"::1", "::", "fe80::1", "2001:db8::1", "2001:db8::",
		"::ffff:192.168.1.1", "::ffff:0.0.0.0",
		"[::1]", "fe80::1%eth0", "[fe80::1%eth0]",
	}
	for _, ip := range valid {
		if !ValidIPv6(ip) {
			t.Errorf("ValidIPv6(%q) = false, want true", ip)
		}
	}
	for _, ip := range []string{"2001:0db8:85a3:0000:zz00:8a2e:0370:7334", "not-an-ip", "", "1.2.3.4"} {
		if ValidIPv6(ip) {
			t.Errorf("ValidIPv6(%q) = true, want false", ip)
		}
	}
}

func TestDetectIPFamily(t *testing.T) {
	cases := map[string]int{
		"127.0.0.1":       4,
		"::1":             6,
		"fe80::1":         6,
		"hl7.example.com": 0,
		"localhost":       0,
	}
	for host, want := range cases {
		if got := DetectIPFamily(host); got != want {
			t.Errorf("DetectIPFamily(%q) = %d, want %d", host, got, want)
		}
	}
}
