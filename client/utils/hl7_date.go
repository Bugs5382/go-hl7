// Package utils carries the small pure helpers the client package uses. It
// mirrors the utils/ folder (createHL7Date, randomString, expBackoff,
// ipAddress, is, spilt, getSegIndexes, decodeHexString, escapeForRegExp).
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
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CreateHL7Date formats a time as an HL7-compatible date string. The length
// selects the format, mirroring the createHL7Date:
//
//	"8"  = YYYYMMDD
//	"12" = YYYYMMDDHHMM
//	"14" = YYYYMMDDHHMMSS (default)
//	"19" = YYYYMMDDHHMMSS.SSSS
//	"24" = YYYYMMDDHHMMSS.SSSS+ZZZZ (with timezone offset)
//	"26" = YYYYMMDDHHMMSS.SSSSSS+ZZZZ (microseconds + timezone)
//
// An empty length defaults to the 14-character form. The date is read in its
// own location, mirroring the use of local-time getters.
func CreateHL7Date(date time.Time, length string) string {
	y := date.Year()
	mo := PadHL7Date(int(date.Month()), 2, "0")
	d := PadHL7Date(date.Day(), 2, "0")
	h := PadHL7Date(date.Hour(), 2, "0")
	mi := PadHL7Date(date.Minute(), 2, "0")
	s := PadHL7Date(date.Second(), 2, "0")
	ms := PadHL7Date(date.Nanosecond()/1e6, 4, "0")

	switch length {
	case "12":
		return fmt.Sprintf("%d%s%s%s%s", y, mo, d, h, mi)
	case "14":
		return fmt.Sprintf("%d%s%s%s%s%s", y, mo, d, h, mi, s)
	case "19":
		return fmt.Sprintf("%d%s%s%s%s%s.%s", y, mo, d, h, mi, s, ms)
	case "24":
		base := fmt.Sprintf("%d%s%s%s%s%s.%s", y, mo, d, h, mi, s, ms)
		return base + tzOffset(date)
	case "26":
		base := fmt.Sprintf("%d%s%s%s%s%s", y, mo, d, h, mi, s)
		micros := PadHL7Date((date.Nanosecond()/1e6)*1000, 6, "0")
		return fmt.Sprintf("%s.%s%s", base, micros, tzOffset(date))
	case "8":
		return fmt.Sprintf("%d%s%s", y, mo, d)
	default:
		return fmt.Sprintf("%d%s%s%s%s%s", y, mo, d, h, mi, s)
	}
}

// tzOffset renders the +ZZZZ/-ZZZZ timezone suffix for the date's location.
func tzOffset(date time.Time) string {
	_, offsetSec := date.Zone()
	offsetMin := offsetSec / 60
	sign := "+"
	if offsetMin < 0 {
		sign = "-"
		offsetMin = -offsetMin
	}
	return fmt.Sprintf("%s%s%s", sign, PadHL7Date(offsetMin/60, 2, "0"), PadHL7Date(offsetMin%60, 2, "0"))
}

// PadHL7Date left-pads the decimal form of n to the given width using the pad
// character z. If n is already at least width digits wide it is returned
// unchanged. It mirrors the padHL7Date.
func PadHL7Date(n, width int, z string) string {
	s := strconv.Itoa(n)
	if len(s) >= width {
		return s
	}
	return strings.Repeat(z, width-len(s)) + s
}
