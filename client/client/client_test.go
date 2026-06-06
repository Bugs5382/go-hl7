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

package client

import (
	"errors"
	"testing"

	"github.com/Bugs5382/go-hl7/client/helpers"
)

// ptr is a small generic helper for the *bool/*int option pointers (the Go
// modeling of the optional option keys).
func ptr[T any](v T) *T { return &v }

// expectHL7FatalError mirrors the test util expectHL7FatalError: the error must
// be an *HL7FatalError with the exact message.
func expectHL7FatalError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	var fatal *helpers.HL7FatalError
	if !errors.As(err, &fatal) {
		t.Fatalf("expected *HL7FatalError, got %T (%v)", err, err)
	}
	if err.Error() != message {
		t.Fatalf("error message = %q, want %q", err.Error(), message)
	}
}

func TestClientValid(t *testing.T) {
	t.Run("valid - properties exist (createConnection available)", func(t *testing.T) {
		client, err := NewClient(ClientOptions{Host: "hl7.server.local"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client == nil {
			t.Fatalf("expected a client")
		}
	})

	t.Run("getHost returns the configured host", func(t *testing.T) {
		client, err := NewClient(ClientOptions{Host: "hl7.server.local"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := client.GetHost(); got != "hl7.server.local" {
			t.Fatalf("GetHost() = %q, want %q", got, "hl7.server.local")
		}
	})

	t.Run("port is set on outbound connection", func(t *testing.T) {
		client, err := NewClient(ClientOptions{Host: "hl7.server.local"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		outbound, err := client.CreateConnection(
			ClientListenerOptions{AutoConnect: ptr(false), Port: ptr(12_345)},
			func(*InboundResponse) error { return nil },
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := outbound.GetPort(); got != 12_345 {
			t.Fatalf("GetPort() = %d, want %d", got, 12_345)
		}
	})
}

func TestClientErrors(t *testing.T) {
	t.Run("accepts ipv4 and ipv6 both true (dual-stack)", func(t *testing.T) {
		if _, err := NewClient(ClientOptions{Host: "5.8.6.1", IPv4: ptr(true), IPv6: ptr(true)}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects malformed IPv4 host when ipv4 is exclusive", func(t *testing.T) {
		_, err := NewClient(ClientOptions{Host: "123.34.52.455", IPv4: ptr(true)})
		if err == nil || err.Error() != "host is not a valid IPv4 address." {
			t.Fatalf("err = %v, want host is not a valid IPv4 address.", err)
		}
	})

	t.Run("accepts valid IPv4 host when ipv4 is exclusive", func(t *testing.T) {
		if _, err := NewClient(ClientOptions{Host: "123.34.52.45", IPv4: ptr(true)}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects malformed IPv6 host when ipv6 is exclusive", func(t *testing.T) {
		_, err := NewClient(ClientOptions{Host: "2001:0db8:85a3:0000:zz00:8a2e:0370:7334", IPv6: ptr(true)})
		if err == nil || err.Error() != "host is not a valid IPv6 address." {
			t.Fatalf("err = %v, want host is not a valid IPv6 address.", err)
		}
	})

	t.Run("accepts valid IPv6 host when ipv6 is exclusive", func(t *testing.T) {
		if _, err := NewClient(ClientOptions{Host: "2001:0db8:85a3:0000:0000:8a2e:0370:7334", IPv6: ptr(true)}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("rejects ipv4 and ipv6 both false", func(t *testing.T) {
		_, err := NewClient(ClientOptions{Host: "192.0.2.1", IPv4: ptr(false), IPv6: ptr(false)})
		want := "ipv4 and ipv6 cannot both be disabled — at least one address family must be enabled."
		if err == nil || err.Error() != want {
			t.Fatalf("err = %v, want %q", err, want)
		}
	})

	t.Run("rejects empty host", func(t *testing.T) {
		_, err := NewClient(ClientOptions{})
		expectHL7FatalError(t, err, "host is not defined or the length is less than 0.")
	})
}

func TestOutboundConnectionOptions(t *testing.T) {
	newClient := func(t *testing.T) *Client {
		t.Helper()
		client, err := NewClient(ClientOptions{Host: "localhost"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		return client
	}

	t.Run("rejects createConnection with no port", func(t *testing.T) {
		client := newClient(t)
		_, err := client.CreateConnection(ClientListenerOptions{}, func(*InboundResponse) error { return nil })
		expectHL7FatalError(t, err, "port is not defined.")
	})

	t.Run("rejects negative port", func(t *testing.T) {
		client := newClient(t)
		_, err := client.CreateConnection(ClientListenerOptions{Port: ptr(-1)}, func(*InboundResponse) error { return nil })
		if err == nil || err.Error() != "port must be a number (1, 65353)." {
			t.Fatalf("err = %v, want port must be a number (1, 65353).", err)
		}
	})

	t.Run("rejects port above 65353", func(t *testing.T) {
		client := newClient(t)
		_, err := client.CreateConnection(ClientListenerOptions{Port: ptr(65_354)}, func(*InboundResponse) error { return nil })
		if err == nil || err.Error() != "port must be a number (1, 65353)." {
			t.Fatalf("err = %v, want port must be a number (1, 65353).", err)
		}
	})

	t.Run("rejects enqueueMessage without flushQueue", func(t *testing.T) {
		client := newClient(t)
		_, err := client.CreateConnection(
			ClientListenerOptions{Port: ptr(12_345), EnqueueMessage: func(MessageItem, NotifyPendingCount) error { return nil }},
			func(*InboundResponse) error { return nil },
		)
		expectHL7FatalError(t, err, "flushQueue is not set.")
	})

	t.Run("rejects flushQueue without enqueueMessage", func(t *testing.T) {
		client := newClient(t)
		_, err := client.CreateConnection(
			ClientListenerOptions{Port: ptr(12_345), FlushQueue: func(FallBackHandler, NotifyPendingCount) error { return nil }},
			func(*InboundResponse) error { return nil },
		)
		expectHL7FatalError(t, err, "enqueueMessage is not set.")
	})
}
