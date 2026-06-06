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

package main

import (
	"fmt"
	"strconv"
	"strings"
)

// jsValue is a parsed JavaScript-literal value: a string, a float64 (number), a
// bool, nil, a []jsValue (array), or a map[string]jsValue (object). It is the
// shape the genspec transpiler reads the committed *.ts metadata into.
type jsValue interface{}

// jsParser is a minimal recursive-descent parser for the JavaScript object and
// array literals that the generated metadata files contain. It handles
// double-quoted strings, numbers (including underscore digit separators like
// 10_240), bare identifiers (treated as strings: true/false/null are special),
// nested objects/arrays, and trailing commas. It is intentionally small: the
// committed catalog uses only this subset.
type jsParser struct {
	src string
	pos int
}

func newJSParser(src string) *jsParser { return &jsParser{src: src} }

func (p *jsParser) errf(format string, a ...any) error {
	return fmt.Errorf("genspec: parse at offset %d: %s", p.pos, fmt.Sprintf(format, a...))
}

// skipWS advances past whitespace and // line / /* block */ comments.
func (p *jsParser) skipWS() {
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			p.pos++
		case c == '/' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '/':
			for p.pos < len(p.src) && p.src[p.pos] != '\n' {
				p.pos++
			}
		case c == '/' && p.pos+1 < len(p.src) && p.src[p.pos+1] == '*':
			p.pos += 2
			for p.pos+1 < len(p.src) && (p.src[p.pos] != '*' || p.src[p.pos+1] != '/') {
				p.pos++
			}
			p.pos += 2
		default:
			return
		}
	}
}

func (p *jsParser) peek() byte {
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

// parseValue parses a single JS-literal value.
func (p *jsParser) parseValue() (jsValue, error) {
	p.skipWS()
	c := p.peek()
	switch {
	case c == '{':
		return p.parseObject()
	case c == '[':
		return p.parseArray()
	case c == '"' || c == '\'':
		return p.parseString()
	case c == '-' || (c >= '0' && c <= '9'):
		return p.parseNumber()
	case isIdentStart(c):
		return p.parseIdent()
	default:
		return nil, p.errf("unexpected character %q", c)
	}
}

func (p *jsParser) parseObject() (jsValue, error) {
	p.pos++ // consume {
	obj := map[string]jsValue{}
	for {
		p.skipWS()
		if p.peek() == '}' {
			p.pos++
			return obj, nil
		}
		key, err := p.parseKey()
		if err != nil {
			return nil, err
		}
		p.skipWS()
		if p.peek() != ':' {
			return nil, p.errf("expected ':' after key %q", key)
		}
		p.pos++ // consume :
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		obj[key] = val
		p.skipWS()
		if p.peek() == ',' {
			p.pos++
			continue
		}
		if p.peek() == '}' {
			p.pos++
			return obj, nil
		}
		return nil, p.errf("expected ',' or '}' in object")
	}
}

func (p *jsParser) parseArray() (jsValue, error) {
	p.pos++ // consume [
	arr := []jsValue{}
	for {
		p.skipWS()
		if p.peek() == ']' {
			p.pos++
			return arr, nil
		}
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		arr = append(arr, val)
		p.skipWS()
		if p.peek() == ',' {
			p.pos++
			continue
		}
		if p.peek() == ']' {
			p.pos++
			return arr, nil
		}
		return nil, p.errf("expected ',' or ']' in array")
	}
}

// parseKey parses an object key: a quoted string or a bare identifier.
func (p *jsParser) parseKey() (string, error) {
	p.skipWS()
	c := p.peek()
	if c == '"' || c == '\'' {
		v, err := p.parseString()
		if err != nil {
			return "", err
		}
		return v.(string), nil
	}
	if isIdentStart(c) || (c >= '0' && c <= '9') {
		start := p.pos
		for p.pos < len(p.src) && (isIdentPart(p.src[p.pos]) || p.src[p.pos] == '.') {
			p.pos++
		}
		return p.src[start:p.pos], nil
	}
	return "", p.errf("expected object key")
}

func (p *jsParser) parseString() (jsValue, error) {
	quote := p.src[p.pos]
	p.pos++ // consume opening quote
	var sb strings.Builder
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if c == '\\' {
			p.pos++
			if p.pos >= len(p.src) {
				break
			}
			esc := p.src[p.pos]
			switch esc {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\', '"', '\'', '/':
				sb.WriteByte(esc)
			default:
				sb.WriteByte(esc)
			}
			p.pos++
			continue
		}
		if c == quote {
			p.pos++
			return sb.String(), nil
		}
		sb.WriteByte(c)
		p.pos++
	}
	return nil, p.errf("unterminated string")
}

func (p *jsParser) parseNumber() (jsValue, error) {
	start := p.pos
	for p.pos < len(p.src) {
		c := p.src[p.pos]
		if (c >= '0' && c <= '9') || c == '.' || c == '-' || c == '+' ||
			c == 'e' || c == 'E' || c == '_' || c == 'x' || c == 'X' {
			p.pos++
			continue
		}
		break
	}
	raw := strings.ReplaceAll(p.src[start:p.pos], "_", "")
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, p.errf("bad number %q: %v", raw, err)
	}
	return f, nil
}

func (p *jsParser) parseIdent() (jsValue, error) {
	start := p.pos
	for p.pos < len(p.src) && isIdentPart(p.src[p.pos]) {
		p.pos++
	}
	word := p.src[start:p.pos]
	switch word {
	case "true":
		return true, nil
	case "false":
		return false, nil
	case "null", "undefined":
		return nil, nil
	default:
		return word, nil
	}
}

func isIdentStart(c byte) bool {
	return c == '_' || c == '$' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isIdentPart(c byte) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

// parseAssignment finds the right-hand side of the first `export const NAME =
// <value>;` (or `const NAME = <value>`) assignment in the source and parses it.
// It is the entry point for the *.ts metadata files, each of which exports a
// single literal.
func parseExportConst(src string) (string, jsValue, error) {
	idx := strings.Index(src, "export const ")
	off := len("export const ")
	if idx < 0 {
		idx = strings.Index(src, "const ")
		off = len("const ")
	}
	if idx < 0 {
		return "", nil, fmt.Errorf("genspec: no export const found")
	}
	rest := src[idx+off:]
	eq := strings.Index(rest, "=")
	if eq < 0 {
		return "", nil, fmt.Errorf("genspec: no '=' after const name")
	}
	namePart := strings.TrimSpace(rest[:eq])
	// Strip any `: Type` annotation from the name.
	if colon := strings.Index(namePart, ":"); colon >= 0 {
		namePart = strings.TrimSpace(namePart[:colon])
	}
	p := newJSParser(rest[eq+1:])
	val, err := p.parseValue()
	if err != nil {
		return "", nil, err
	}
	return namePart, val, nil
}
