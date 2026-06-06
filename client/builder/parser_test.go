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

// Mirrors __tests__/client/hl7.parser.test.ts.
package builder

import "testing"

func mustMessage(t *testing.T, text string) *Message {
	t.Helper()
	m, err := NewMessage(MessageOptions{Text: text})
	if err != nil {
		t.Fatalf("NewMessage: %v", err)
	}
	return m
}

func TestHL7ParserTests(t *testing.T) {
	t.Run("field separations", func(t *testing.T) {
		t.Run("parses repeating PID.3 with multiple identifier types", func(t *testing.T) {
			text := "MSH|^~\\&|HUBWS|46355||DAL|202412091132||ORM^O01|MZ54932|P|2.3\rPID|1|999-99-9999|CHART^^^CID~MEDICAL^^^MRN||PTLASTNAME^PTFIRSTNAME^||19750825|F|||4690 MAIN STREET^^MASON^OH^45040||^^^^^513^5550124||||||999654321||"
			m := mustMessage(t, text)
			checks := []struct {
				got, want string
			}{
				{m.Get("PID.1").String(), "1"},
				{m.Get("PID.2").String(), "999-99-9999"},
				{m.Get("PID.3").Index(0).Index(0).String(), "CHART"},
				{m.Get("PID.3").Index(0).Index(3).String(), "CID"},
				{m.Get("PID.3").Index(1).Index(0).String(), "MEDICAL"},
				{m.Get("PID.3").Index(1).Index(3).String(), "MRN"},
				{m.Get("PID.5.1").String(), "PTLASTNAME"},
			}
			for i, c := range checks {
				if c.got != c.want {
					t.Fatalf("check %d: got %q want %q", i, c.got, c.want)
				}
			}
		})

		t.Run("parses repeating PID.3 with namespace identifiers", func(t *testing.T) {
			text := "MSH|^~\\&|ZIS|Testziekenhuis^Capelle ad IJssel|||20080917161411||ADT^A08^ADT_A01|CLV1-1649|P|2.4|||AL|NE|NLD|8859/1|NL\rEVN|A08|20080917161412\rPID|1||7137542^^^^PI~123456782^^^NLMINBIZA^NNNLD^^20080917~AA1234567^^^NLMINBIZA^PPN^^20080917||van Test^Jeanet||19600101|F"
			m := mustMessage(t, text)
			if got := m.Get("PID.3").Index(1).Index(0).String(); got != "123456782" {
				t.Fatalf("got %q", got)
			}
		})

		t.Run("parses repeating PID.3 across LF-separated segments", func(t *testing.T) {
			text := "MSH|^~\\&|1.2||||20250126162659||ADT^A03^ADT_A03||P|2.5^FRA^2.5|||||FRA|8859/1|||\n" +
				"PID|||0001^^^MCK&1.2&L^PI~7777^^^ASIP-SANTE-INS-C&1.3&ISO^INS-C^^|"
			m := mustMessage(t, text)
			if got := m.Get("PID.3").Index(1).Index(0).String(); got != "7777" {
				t.Fatalf("got %q", got)
			}
		})
	})
}
