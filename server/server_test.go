package server

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
	"errors"
	"strings"
	"testing"

	"github.com/Bugs5382/go-hl7/server/utils"
)

func newServer(t *testing.T, opts *ServerOptions) *Server {
	t.Helper()
	s, err := NewServer(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return s
}

func TestInboundRequestModule(t *testing.T) {
	t.Run("getMessage panics when no Message was provided", func(t *testing.T) {
		defer func() {
			r := recover()
			err, ok := r.(error)
			if !ok || err.Error() != "Message is not defined." {
				t.Fatalf("recover = %v, want HL7ListenerError 'Message is not defined.'", r)
			}
			var listenerErr *utils.HL7ListenerError
			if !errors.As(err, &listenerErr) {
				t.Fatalf("recover type = %T, want *HL7ListenerError", err)
			}
		}()
		empty := NewInboundRequest(nil, InboundRequestProps{Type: "file"})
		empty.GetMessage()
	})

	t.Run("getType returns the configured request type", func(t *testing.T) {
		request := NewInboundRequest(nil, InboundRequestProps{Type: "file"})
		if got := request.GetType(); got != "file" {
			t.Fatalf("GetType() = %q, want file", got)
		}
	})
}

func TestServerClass(t *testing.T) {
	t.Run("accepts ipv4 and ipv6 both true (dual-stack)", func(t *testing.T) {
		if _, err := NewServer(&ServerOptions{IPv4: ptr(true), IPv6: ptr(true)}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects ipv4 and ipv6 both false", func(t *testing.T) {
		_, err := NewServer(&ServerOptions{IPv4: ptr(false), IPv6: ptr(false)})
		want := "ipv4 and ipv6 cannot both be disabled — at least one address family must be enabled."
		if err == nil || err.Error() != want {
			t.Fatalf("err = %v, want %q", err, want)
		}
	})

	t.Run("rejects empty bindAddress when ipv4 is exclusive", func(t *testing.T) {
		_, err := NewServer(&ServerOptions{BindAddress: ptr(""), IPv4: ptr(true)})
		if err == nil || err.Error() != "bindAddress is an invalid ipv4 address." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("rejects malformed bindAddress when ipv4 is exclusive", func(t *testing.T) {
		_, err := NewServer(&ServerOptions{BindAddress: ptr("123.34.52.455"), IPv4: ptr(true)})
		if err == nil || err.Error() != "bindAddress is an invalid ipv4 address." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("rejects garbage bindAddress in dual-stack mode", func(t *testing.T) {
		_, err := NewServer(&ServerOptions{BindAddress: ptr("not-an-ip"), IPv4: ptr(true), IPv6: ptr(true)})
		if err == nil || err.Error() != "bindAddress is not a valid IPv4 or IPv6 address." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("rejects empty bindAddress when ipv6 is exclusive", func(t *testing.T) {
		_, err := NewServer(&ServerOptions{BindAddress: ptr(""), IPv4: ptr(false), IPv6: ptr(true)})
		if err == nil || err.Error() != "bindAddress is an invalid ipv6 address." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("rejects malformed bindAddress when ipv6 is exclusive", func(t *testing.T) {
		_, err := NewServer(&ServerOptions{BindAddress: ptr("2001:0db8:85a3:0000:zz00:8a2e:0370:7334"), IPv4: ptr(false), IPv6: ptr(true)})
		if err == nil || err.Error() != "bindAddress is an invalid ipv6 address." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("accepts a valid IPv6 bindAddress when ipv6 is exclusive", func(t *testing.T) {
		if _, err := NewServer(&ServerOptions{BindAddress: ptr("2001:0db8:85a3:0000:0000:8a2e:0370:7334"), IPv4: ptr(false), IPv6: ptr(true)}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("properties exist (createInbound available)", func(t *testing.T) {
		_ = newServer(t, nil)
	})
}

func TestServerOptionNormalization(t *testing.T) {
	// Server option normalization.
	t.Run("IPv4-only is the default", func(t *testing.T) {
		s := newServer(t, nil)
		if !s.opt.IPv4 || s.opt.IPv6 || s.opt.BindAddress != "0.0.0.0" || s.opt.IPv6Only {
			t.Fatalf("opt = %+v", s.opt)
		}
	})
	t.Run("ipv4=true alone defaults bindAddress to 0.0.0.0", func(t *testing.T) {
		s := newServer(t, &ServerOptions{IPv4: ptr(true)})
		if !s.opt.IPv4 || s.opt.IPv6 || s.opt.BindAddress != "0.0.0.0" || s.opt.IPv6Only {
			t.Fatalf("opt = %+v", s.opt)
		}
	})
	t.Run("ipv6=true alone defaults bindAddress to :: with ipv6Only", func(t *testing.T) {
		s := newServer(t, &ServerOptions{IPv6: ptr(true)})
		if s.opt.IPv4 || !s.opt.IPv6 || s.opt.BindAddress != "::" || !s.opt.IPv6Only {
			t.Fatalf("opt = %+v", s.opt)
		}
	})
	t.Run("dual-stack opt-in: both true", func(t *testing.T) {
		s := newServer(t, &ServerOptions{IPv4: ptr(true), IPv6: ptr(true)})
		if !s.opt.IPv4 || !s.opt.IPv6 || s.opt.BindAddress != "::" || s.opt.IPv6Only {
			t.Fatalf("opt = %+v", s.opt)
		}
	})
	t.Run("pin specific IPv4 bindAddress", func(t *testing.T) {
		s := newServer(t, &ServerOptions{BindAddress: ptr("127.0.0.1"), IPv4: ptr(true)})
		if s.opt.BindAddress != "127.0.0.1" {
			t.Fatalf("bindAddress = %q", s.opt.BindAddress)
		}
	})
	t.Run("pin specific IPv6 bindAddress", func(t *testing.T) {
		s := newServer(t, &ServerOptions{BindAddress: ptr("::1"), IPv6: ptr(true)})
		if s.opt.BindAddress != "::1" {
			t.Fatalf("bindAddress = %q", s.opt.BindAddress)
		}
	})
	t.Run("localhost is always allowed", func(t *testing.T) {
		newServer(t, &ServerOptions{BindAddress: ptr("localhost")})
		newServer(t, &ServerOptions{BindAddress: ptr("localhost"), IPv4: ptr(true)})
		newServer(t, &ServerOptions{BindAddress: ptr("localhost"), IPv6: ptr(true)})
	})
	t.Run("ipv6-only rejects IPv4 bindAddress", func(t *testing.T) {
		_, err := NewServer(&ServerOptions{BindAddress: ptr("127.0.0.1"), IPv6: ptr(true)})
		if err == nil || err.Error() != "bindAddress is an invalid ipv6 address." {
			t.Fatalf("err = %v", err)
		}
	})
	t.Run("ipv4-only rejects IPv6 bindAddress", func(t *testing.T) {
		_, err := NewServer(&ServerOptions{BindAddress: ptr("::1"), IPv4: ptr(true)})
		if err == nil || err.Error() != "bindAddress is an invalid ipv4 address." {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestListenerClass(t *testing.T) {
	t.Run("rejects createInbound with no port", func(t *testing.T) {
		s := newServer(t, nil)
		_, err := s.CreateInbound(ListenerOptions{}, func(*InboundRequest, ResponseSender) error { return nil })
		if err == nil || err.Error() != "port is not defined." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("rejects negative port", func(t *testing.T) {
		s := newServer(t, nil)
		_, err := s.CreateInbound(ListenerOptions{Port: ptr(-1)}, func(*InboundRequest, ResponseSender) error { return nil })
		if err == nil || err.Error() != "port must be a number (0, 65353)." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("rejects port above 65353", func(t *testing.T) {
		s := newServer(t, nil)
		_, err := s.CreateInbound(ListenerOptions{Port: ptr(65_354)}, func(*InboundRequest, ResponseSender) error { return nil })
		if err == nil || err.Error() != "port must be a number (0, 65353)." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("rejects name with disallowed special characters", func(t *testing.T) {
		s := newServer(t, nil)
		_, err := s.CreateInbound(ListenerOptions{Name: "$#@!sdfe-`", Port: ptr(65_353)}, func(*InboundRequest, ResponseSender) error { return nil })
		if err == nil || !strings.Contains(err.Error(), "name must not contain certain characters") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("rejects mshOverrides with an empty (invalid) override", func(t *testing.T) {
		s := newServer(t, nil)
		_, err := s.CreateInbound(
			ListenerOptions{Name: "mshOverride", Port: ptr(4000), MSHOverrides: map[string]MSHOverride{"9.3": {}}},
			func(*InboundRequest, ResponseSender) error { return nil },
		)
		if err == nil || !strings.Contains(err.Error(), "mshOverrides override value must be a string or a function.") {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("rejects createInbound with no version", func(t *testing.T) {
		s := newServer(t, nil)
		_, err := s.CreateInbound(ListenerOptions{Port: ptr(4000)}, func(*InboundRequest, ResponseSender) error { return nil })
		if err == nil || err.Error() != "version is not defined." {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("rejects createInbound with an invalid version", func(t *testing.T) {
		s := newServer(t, nil)
		_, err := s.CreateInbound(ListenerOptions{Port: ptr(4000), Version: "9.9"}, func(*InboundRequest, ResponseSender) error { return nil })
		if err == nil || err.Error() != "version is not a valid HL7 version." {
			t.Fatalf("err = %v", err)
		}
	})
}
