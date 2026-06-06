package client

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
	"testing"
)

// Client option normalization.

func TestClientOptionNormalization(t *testing.T) {
	t.Run("IPv4-only is the default", func(t *testing.T) {
		c, err := NewClient(ClientOptions{Version: "2.7", Host: "hl7.example.com"})
		if err != nil {
			t.Fatal(err)
		}
		if !c.opt.ipv4 || c.opt.ipv6 || c.opt.family != 4 {
			t.Fatalf("ipv4=%v ipv6=%v family=%d, want true,false,4", c.opt.ipv4, c.opt.ipv6, c.opt.family)
		}
	})

	t.Run("explicit ipv4:true alone forces IPv4-only family", func(t *testing.T) {
		c, _ := NewClient(ClientOptions{Version: "2.7", Host: "192.0.2.1", IPv4: ptr(true)})
		if !c.opt.ipv4 || c.opt.ipv6 || c.opt.family != 4 {
			t.Fatalf("ipv4=%v ipv6=%v family=%d, want true,false,4", c.opt.ipv4, c.opt.ipv6, c.opt.family)
		}
	})

	t.Run("explicit ipv6:true alone forces IPv6-only family", func(t *testing.T) {
		c, _ := NewClient(ClientOptions{Version: "2.7", Host: "::1", IPv6: ptr(true)})
		if c.opt.ipv4 || !c.opt.ipv6 || c.opt.family != 6 {
			t.Fatalf("ipv4=%v ipv6=%v family=%d, want false,true,6", c.opt.ipv4, c.opt.ipv6, c.opt.family)
		}
	})

	t.Run("dual-stack opt-in: both true (family 0)", func(t *testing.T) {
		c, _ := NewClient(ClientOptions{Version: "2.7", Host: "hl7.example.com", IPv4: ptr(true), IPv6: ptr(true)})
		if !c.opt.ipv4 || !c.opt.ipv6 || c.opt.family != 0 || !c.opt.autoSelectFamily {
			t.Fatalf("ipv4=%v ipv6=%v family=%d autoSelectFamily=%v", c.opt.ipv4, c.opt.ipv6, c.opt.family, c.opt.autoSelectFamily)
		}
	})

	t.Run("dual-stack with IPv6 literal pins family to 6", func(t *testing.T) {
		c, _ := NewClient(ClientOptions{Version: "2.7", Host: "::1", IPv4: ptr(true), IPv6: ptr(true)})
		if c.opt.family != 6 {
			t.Fatalf("family=%d, want 6", c.opt.family)
		}
	})

	t.Run("dual-stack with IPv4 literal pins family to 4", func(t *testing.T) {
		c, _ := NewClient(ClientOptions{Version: "2.7", Host: "127.0.0.1", IPv4: ptr(true), IPv6: ptr(true)})
		if c.opt.family != 4 {
			t.Fatalf("family=%d, want 4", c.opt.family)
		}
	})

	t.Run("autoSelectFamily can be disabled", func(t *testing.T) {
		c, _ := NewClient(ClientOptions{Version: "2.7", Host: "hl7.example.com", IPv4: ptr(true), IPv6: ptr(true), AutoSelectFamily: ptr(false)})
		if c.opt.autoSelectFamily {
			t.Fatalf("autoSelectFamily = true, want false")
		}
	})

	t.Run("rejects an IPv6 literal when ipv4 is exclusive", func(t *testing.T) {
		_, err := NewClient(ClientOptions{Version: "2.7", Host: "::1", IPv4: ptr(true)})
		if err == nil || err.Error() != "host is not a valid IPv4 address." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("rejects an IPv4 literal when ipv6 is exclusive", func(t *testing.T) {
		_, err := NewClient(ClientOptions{Version: "2.7", Host: "127.0.0.1", IPv6: ptr(true)})
		if err == nil || err.Error() != "host is not a valid IPv6 address." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("validates autoSelectFamilyAttemptTimeout range", func(t *testing.T) {
		_, err := NewClient(ClientOptions{Version: "2.7", Host: "192.0.2.1", AutoSelectFamilyAttemptTimeout: ptr(5)})
		if err == nil || !strings.Contains(err.Error(), "autoSelectFamilyAttemptTimeout must be a number") {
			t.Fatalf("err = %v", err)
		}
	})
}
