package hl7_test

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
	"testing"
	"time"

	"github.com/Bugs5382/go-hl7/client/hl7"
)

// These tests mirror the hl7.builder-chaining.test.ts: BuildXXX returns
// the receiver, and chained and imperative builds produce identical output.

func TestBuilderChaining(t *testing.T) {
	date := time.Date(2024, 1, 15, 10, 20, 30, 0, time.UTC)

	t.Run("buildXXX returns the builder instance", func(t *testing.T) {
		b := hl7.New(hl7.V2_8)
		rv := b.BuildMSH(hl7.Props{"msh_10": "X", "msh_11_1": "P", "msh_7": date, "msh_9_1": "ADT", "msh_9_2": "A01"})
		if rv != b {
			t.Fatal("expected BuildMSH to return the receiver")
		}
	})

	t.Run("chained calls keep working on the version builder", func(t *testing.T) {
		out := hl7.New(hl7.V2_4).
			BuildMSH(hl7.Props{"msh_10": "X", "msh_11_1": "P", "msh_7": date, "msh_9_1": "ADT", "msh_9_2": "A01"}).
			BuildECD(hl7.Props{"ecd_1": "1", "ecd_2": "RC"}).
			String()
		contains(t, out, "MSH|")
		contains(t, out, "ECD|")
	})

	t.Run("chained and imperative builds produce identical output", func(t *testing.T) {
		chained := hl7.New(hl7.V2_4).
			BuildMSH(hl7.Props{"msh_10": "ID1", "msh_11_1": "P", "msh_7": date, "msh_9_1": "ADT", "msh_9_2": "A01"}).
			BuildECD(hl7.Props{"ecd_1": "1", "ecd_2": "RC", "ecd_3": "Y"}).
			String()

		imperative := hl7.New(hl7.V2_4)
		imperative.BuildMSH(hl7.Props{"msh_10": "ID1", "msh_11_1": "P", "msh_7": date, "msh_9_1": "ADT", "msh_9_2": "A01"})
		imperative.BuildECD(hl7.Props{"ecd_1": "1", "ecd_2": "RC", "ecd_3": "Y"})

		if chained != imperative.String() {
			t.Fatalf("chained %q != imperative %q", chained, imperative.String())
		}
	})
}
