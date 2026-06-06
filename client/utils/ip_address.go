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
	"net"
	"strings"
)

// ValidIPv4 reports whether ip is a valid IPv4 literal. It mirrors the reference's
// validIPv4 (backed by The net.isIPv4, dotted-decimal only).
func ValidIPv4(ip string) bool {
	if ip == "" {
		return false
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	// net.ParseIP accepts both families; restrict to dotted-decimal IPv4 to
	// match The net.isIPv4.
	return strings.Count(ip, ".") == 3 && parsed.To4() != nil
}

// ValidIPv6 reports whether ip is a valid IPv6 literal. It mirrors the reference's
// validIPv6: bracketed forms ([::1]) and zone-id suffixes (fe80::1%eth0) are
// stripped before validation, and IPv4-mapped IPv6 addresses are accepted.
func ValidIPv6(ip string) bool {
	if ip == "" {
		return false
	}
	candidate := strings.TrimSpace(ip)
	if strings.HasPrefix(candidate, "[") && strings.HasSuffix(candidate, "]") {
		candidate = candidate[1 : len(candidate)-1]
	}
	if i := strings.IndexByte(candidate, '%'); i != -1 {
		candidate = candidate[:i]
	}
	parsed := net.ParseIP(candidate)
	if parsed == nil {
		return false
	}
	return strings.Contains(candidate, ":")
}

// DetectIPFamily returns 4 for valid IPv4 literals, 6 for valid IPv6 literals,
// or 0 for hostnames/FQDNs that should be resolved via DNS. It mirrors
// the detectIPFamily.
func DetectIPFamily(value string) int {
	if ValidIPv4(value) {
		return 4
	}
	if ValidIPv6(value) {
		return 6
	}
	return 0
}
