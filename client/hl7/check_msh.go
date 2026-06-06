package hl7

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
	"github.com/Bugs5382/go-hl7/client/helpers"
)

// CheckMSH validates an MSH header against the current spec version, dispatching
// the per-version checks. It returns nil when the header is
// valid and an error (wrapping ErrFatal/ErrValidation) describing the first
// failure otherwise. The messages are preserved verbatim so callers matching on
// the text see identical behavior. v2.1 has no checkMSH and returns a fatal
// "Not Implemented".
func (b *Builder) CheckMSH(msh Props) error {
	switch b.version {
	case "2.1":
		return helpers.NewHL7FatalError("Not Implemented")
	case "2.2":
		return checkMSHBase(msh)
	case "2.3", "2.3.1":
		return checkMSH23(msh)
	default: // 2.4, 2.5, 2.5.1, 2.6, 2.7, 2.7.1, 2.8
		return checkMSH24Plus(b.version, msh)
	}
}

// has reports whether key is present with a non-nil value.
func has(p Props, key string) (string, bool) {
	v, ok := p[key]
	if !ok || v == nil {
		return "", false
	}
	s, _ := v.(string)
	return s, true
}

// checkMSHBase validates the base MSH rules (9.1/9.2 len 3, msh_10 max 20).
func checkMSHBase(msh Props) error {
	s91, ok91 := has(msh, "msh_9_1")
	s92, ok92 := has(msh, "msh_9_2")
	if !ok91 || !ok92 {
		return helpers.NewHL7ValidationError("MSH.9.1 & MSH 9.2 must be defined.")
	}
	if len(s91) != 3 {
		return helpers.NewHL7ValidationError("MSH.9.1 must be 3 characters in length.")
	}
	if len(s92) != 3 {
		return helpers.NewHL7ValidationError("MSH.9.2 must be 3 characters in length.")
	}
	if s10, ok := has(msh, "msh_10"); ok && (len(s10) == 0 || len(s10) > 20) {
		return helpers.NewHL7ValidationError("MSH.10 must be greater than 0 and less than 20 characters.")
	}
	return nil
}

// checkMSH23 validates the v2.3 MSH rules (base + 11.1/11.2 length checks).
func checkMSH23(msh Props) error {
	if err := checkMSHBase(msh); err != nil {
		return err
	}
	s111, _ := has(msh, "msh_11_1")
	if len(s111) > 1 {
		return helpers.NewHL7ValidationError("MSH.11.1 has to be 1 character long. Valid inputs are: D, P, or T")
	}
	if s112, ok := has(msh, "msh_11_2"); ok && (len(s112) > 1 || s112 == "") {
		return helpers.NewHL7ValidationError("MSH.11.2 can either be undefined/blank and 1 character long.")
	}
	return nil
}

// checkMSH24Plus validates the v2.4 rules (base + MSH.9.3 length) and the v2.7
// override (9.1/9.2 len 3, msh_10 max 199). v2.4-2.6 use the 2.3-derived base;
// v2.7+ uses the 2.7 rules.
func checkMSH24Plus(version string, msh Props) error {
	switch version {
	case "2.4", "2.5", "2.5.1", "2.6":
		if err := checkMSH23(msh); err != nil {
			return err
		}
		if s93, ok := has(msh, "msh_9_3"); ok && (len(s93) < 3 || len(s93) > 10) {
			return helpers.NewHL7ValidationError("MSH.9.3 must be 3 to 10 characters in length if specified.")
		}
		return nil
	default: // 2.7, 2.7.1, 2.8
		return checkMSH27(msh)
	}
}

// checkMSH27 validates the v2.7 MSH rules (9.1/9.2 len 3, msh_10 max 199).
func checkMSH27(msh Props) error {
	s91, ok91 := has(msh, "msh_9_1")
	s92, ok92 := has(msh, "msh_9_2")
	if !ok91 || !ok92 {
		return helpers.NewHL7ValidationError("MSH.9.1 & MSH 9.2 must be defined.")
	}
	if len(s91) != 3 {
		return helpers.NewHL7ValidationError("MSH.9.1 must be 3 characters in length.")
	}
	if len(s92) != 3 {
		return helpers.NewHL7ValidationError("MSH.9.2 must be 3 characters in length.")
	}
	if s10, ok := has(msh, "msh_10"); ok && len(s10) > 199 {
		return helpers.NewHL7ValidationError("MSH.10 must be greater than 0 and less than 199 characters.")
	}
	return nil
}
