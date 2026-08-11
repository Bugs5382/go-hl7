package modules

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
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/htmlindex"
	"golang.org/x/text/encoding/ianaindex"

	"github.com/Bugs5382/go-hl7/client/helpers"
)

// hl7CharsetAliases maps HL7 MSH-18 (HL7 Table 0211) charset codes, and a few
// convenient short forms, to concrete text encodings. The 8859/x codes are
// mapped straight to the charmap encodings so they decode as true ISO-8859-x
// (the WHATWG htmlindex would otherwise fold "iso-8859-1" onto windows-1252).
// Keys are upper-cased; callers normalize before the lookup.
var hl7CharsetAliases = map[string]encoding.Encoding{
	"8859/1":  charmap.ISO8859_1,
	"8859/2":  charmap.ISO8859_2,
	"8859/3":  charmap.ISO8859_3,
	"8859/4":  charmap.ISO8859_4,
	"8859/5":  charmap.ISO8859_5,
	"8859/6":  charmap.ISO8859_6,
	"8859/7":  charmap.ISO8859_7,
	"8859/8":  charmap.ISO8859_8,
	"8859/9":  charmap.ISO8859_9,
	"8859/15": charmap.ISO8859_15,

	"ISO-8859-1":  charmap.ISO8859_1,
	"ISO-8859-2":  charmap.ISO8859_2,
	"ISO-8859-3":  charmap.ISO8859_3,
	"ISO-8859-4":  charmap.ISO8859_4,
	"ISO-8859-5":  charmap.ISO8859_5,
	"ISO-8859-6":  charmap.ISO8859_6,
	"ISO-8859-7":  charmap.ISO8859_7,
	"ISO-8859-8":  charmap.ISO8859_8,
	"ISO-8859-9":  charmap.ISO8859_9,
	"ISO-8859-15": charmap.ISO8859_15,

	"WINDOWS-1250": charmap.Windows1250,
	"WINDOWS-1251": charmap.Windows1251,
	"WINDOWS-1252": charmap.Windows1252,
}

// ResolveCharset resolves an HL7 MSH-18 charset code or an IANA/WHATWG name to a
// text encoding. It returns a nil encoding for the empty string and for the
// UTF-8/ASCII names, signaling the UTF-8 byte-slice pass-through the codec uses
// by default. An unrecognized name returns an *HL7FatalError so callers can fail
// cleanly (issue #26).
func ResolveCharset(charset string) (encoding.Encoding, error) {
	name := strings.ToUpper(strings.TrimSpace(charset))
	if name == "" {
		return nil, nil
	}

	// ASCII is a UTF-8 subset and UTF-8 is the native representation, so both use
	// the pass-through (nil) path rather than an encoding transform.
	switch name {
	case "ASCII", "US-ASCII", "USASCII":
		return nil, nil
	case "UNICODE", "UNICODE UTF-8", "UTF-8", "UTF8":
		return nil, nil
	}

	if enc, ok := hl7CharsetAliases[name]; ok {
		return enc, nil
	}

	// Fall back to the IANA registry, then to the WHATWG index, for any other
	// standard charset name (for example "windows-1256" or "gb18030").
	if enc, err := ianaindex.MIME.Encoding(charset); err == nil && enc != nil {
		return enc, nil
	}
	if enc, err := htmlindex.Get(charset); err == nil && enc != nil {
		return enc, nil
	}

	return nil, helpers.NewHL7FatalErrorf("unsupported charset: %q", charset)
}
