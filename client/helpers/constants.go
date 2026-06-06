package helpers

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

import "regexp"

// MLLP control bytes. These mirror the PROTOCOL_MLLP_HEADER/END/FOOTER.
const (
	// ProtocolMLLPHeader is the MLLP start-block byte (VT, 0x0B).
	ProtocolMLLPHeader byte = 0x0b
	// ProtocolMLLPEnd is the MLLP end-block byte (FS, 0x1C).
	ProtocolMLLPEnd byte = 0x1c
	// ProtocolMLLPFooter is the MLLP carriage-return terminator byte (0x0D).
	ProtocolMLLPFooter byte = 0x0d
)

// NameFormat matches characters that are not permitted in a listener/connection
// name. It mirrors the NAME_FORMAT regular expression.
var NameFormat = regexp.MustCompile("[ `!@#$%^&*()+\\-=\\[\\]{};':\"\\\\|,.<>/?~]")
