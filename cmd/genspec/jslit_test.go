package main

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

import "testing"

func TestParseExportConstObject(t *testing.T) {
	src := `import { SegmentSpec } from "@/x";
/** doc */
export const ECD_SPEC: SegmentSpec = {
  description: "Equipment Command",
  fields: [
    { hl7Type: "NM", name: "Ref", num: 1, usage: { "2.4": "R", "2.8": "R" } },
    {
      hl7Type: "ST",
      length: { max: 10_240 },
      name: "Big",
      num: 2,
      table: 396,
      usage: { "2.5.1": "B" },
    },
  ],
  name: "ECD",
  versions: ["2.4", "2.8"],
};`
	name, val, err := parseExportConst(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if name != "ECD_SPEC" {
		t.Fatalf("name = %q", name)
	}
	obj := val.(map[string]jsValue)
	if asString(obj["name"]) != "ECD" {
		t.Fatalf("name field = %q", asString(obj["name"]))
	}
	fields := obj["fields"].([]jsValue)
	if len(fields) != 2 {
		t.Fatalf("fields = %d", len(fields))
	}
	f2 := fields[1].(map[string]jsValue)
	if max, set := lengthMax(f2["length"]); !set || max != 10240 {
		t.Fatalf("length underscore separator: max=%d set=%v", max, set)
	}
	if asInt(f2["table"]) != 396 {
		t.Fatalf("table = %d", asInt(f2["table"]))
	}
	usage := f2["usage"].(map[string]jsValue)
	if asString(usage["2.5.1"]) != "B" {
		t.Fatalf("dotted version key lost: %v", usage)
	}
}

func TestParseArrayTable(t *testing.T) {
	src := `export const TABLE_0002 = [
  "A",
  "B",
  "C",
];`
	name, val, err := parseExportConst(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if name != "TABLE_0002" {
		t.Fatalf("name = %q", name)
	}
	arr := val.([]jsValue)
	if len(arr) != 3 || asString(arr[0]) != "A" || asString(arr[2]) != "C" {
		t.Fatalf("array = %#v", arr)
	}
}

func TestParseSpecMapPreservesOrder(t *testing.T) {
	src := `export const SEGMENT_SPECS: Readonly<Record<string, SegmentSpec>> = {
  ABS: ABS_SPEC,
  ZL7: ZL7_SPEC,
  Zxx: Zxx_SPEC,
};`
	entries, err := parseSpecMap(src, "SEGMENT_SPECS")
	if err != nil {
		t.Fatalf("parseSpecMap: %v", err)
	}
	want := []specMapEntry{
		{"ABS", "ABS_SPEC"},
		{"ZL7", "ZL7_SPEC"},
		{"Zxx", "Zxx_SPEC"},
	}
	if len(entries) != len(want) {
		t.Fatalf("entries = %d, want %d", len(entries), len(want))
	}
	for i, e := range entries {
		if e != want[i] {
			t.Fatalf("entry %d = %+v, want %+v", i, e, want[i])
		}
	}
}
