package utils

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
	"regexp"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/hl7/metadata"
	clientutils "github.com/Bugs5382/go-hl7/client/utils"
)

// TLSConfig mirrors the subset of the tls.ConnectionOptions the server uses
// (ca, cert, key, requestCert). A non-nil *TLSConfig enables TLS.
type TLSConfig struct {
	// CA is the PEM-encoded CA bundle (tls `ca`).
	CA []byte
	// Cert is the PEM-encoded server certificate (tls `cert`).
	Cert []byte
	// Key is the PEM-encoded private key (tls `key`).
	Key []byte
	// RequestCert requests a client certificate (tls `requestCert`).
	RequestCert bool
}

// MSHOverride is a single MSH-field override value: either a literal string or
// a function computing the value from the inbound message. the
// `Record<string, ((message) => string) | string>` becomes a map of these;
// exactly one of String/Func is meaningful (Func wins when non-nil).
type MSHOverride struct {
	// String is the literal override value.
	String string
	// Func computes the override from the inbound message.
	Func func(message *builder.Message) string
	// isString records that this override was built from a string (so an empty
	// literal still validates), mirroring the typeof === "string" check.
	isString bool
}

// StringOverride builds a literal MSH override (the string form).
func StringOverride(value string) MSHOverride { return MSHOverride{String: value, isString: true} }

// FuncOverride builds a computed MSH override (the function form).
func FuncOverride(fn func(message *builder.Message) string) MSHOverride {
	return MSHOverride{Func: fn}
}

// valid reports whether the override carries a usable value, mirroring the
// "must be a string or a function" guard.
func (o MSHOverride) valid() bool { return o.isString || o.Func != nil }

// ServerOptions mirrors the ServerOptions. Pointer fields
// distinguish "not provided" from explicit, matching the hasOwnProperty
// checks for the dual-stack ipv4/ipv6 semantics.
type ServerOptions struct {
	// BindAddress is the network address to listen on; defaults depend on
	// ipv4/ipv6 (bindAddress).
	BindAddress *string
	// Encoding is retained for parity; Go bodies are UTF-8 (encoding).
	Encoding string
	// IPv4 accepts IPv4 connections (ipv4, default true).
	IPv4 *bool
	// IPv6 accepts IPv6 connections (ipv6, default false).
	IPv6 *bool
	// TLS enables TLS when non-nil (tls).
	TLS *TLSConfig
}

// ListenerOptions mirrors the ListenerOptions (per-inbound
// createInbound options).
type ListenerOptions struct {
	// Encoding is retained for parity (encoding).
	Encoding string
	// MSHOverrides optionally overrides ACK MSH fields keyed by path (
	// mshOverrides).
	MSHOverrides map[string]MSHOverride
	// Name names the listener; defaults to a random string (name).
	Name string
	// Port is the network address to listen on, 0..65353 (port, required).
	Port *int
	// Version is the REQUIRED HL7 version this listener accepts. It must be one
	// of the known HL7 versions (2.1, 2.2, 2.3, 2.3.1, 2.4, 2.5, 2.5.1, 2.6,
	// 2.7, 2.7.1, 2.8). This is an intentional divergence from node-hl7, which
	// leaves the transport version-agnostic: here each port enforces its own
	// version and an inbound message whose MSH.12 differs is rejected with an
	// AR (Application Reject) ACK before the handler runs.
	Version string
}

// NormalizedServerOptions is the fully-resolved server option set, mirroring
// the NormalizedServerOptions.
type NormalizedServerOptions struct {
	// BindAddress is the resolved listen address.
	BindAddress string
	// Encoding is the resolved encoding.
	Encoding string
	// IPv4 is the resolved IPv4 flag.
	IPv4 bool
	// IPv6 is the resolved IPv6 flag.
	IPv6 bool
	// IPv6Only is forwarded to the listener (true for IPv6-only).
	IPv6Only bool
	// TLS is the resolved TLS config (nil when plain TCP).
	TLS *TLSConfig
}

// ValidatedListenerOptions is the fully-resolved per-inbound option set,
// mirroring the ValidatedOptions.
type ValidatedListenerOptions struct {
	// Encoding is the resolved encoding.
	Encoding string
	// MSHOverrides is the validated override map.
	MSHOverrides map[string]MSHOverride
	// Name is the resolved (or generated) listener name.
	Name string
	// Port is the validated listen port.
	Port int
	// Version is the validated, required HL7 version this listener enforces;
	// inbound messages whose MSH.12 differs are rejected with an AR ACK.
	Version string
}

var nameFormatRE = regexp.MustCompile("[ `!@#$%^&*()+\\-=\\[\\]{};':\"\\\\|,.<>/?~]")

// NormalizeListenerOptions validates and fills the per-inbound options,
// mirroring the normalizeListenerOptions. It returns an error where the spec
// throws (HL7ListenerError) or assertNumber throws (plain error).
func NormalizeListenerOptions(properties ListenerOptions) (ValidatedListenerOptions, error) {
	out := ValidatedListenerOptions{
		Encoding:     properties.Encoding,
		MSHOverrides: properties.MSHOverrides,
		Name:         properties.Name,
		Version:      properties.Version,
	}
	if out.Encoding == "" {
		out.Encoding = "utf8"
	}

	if out.Name == "" {
		out.Name = clientutils.RandomString(20)
	} else if nameFormatRE.MatchString(out.Name) {
		return ValidatedListenerOptions{}, NewHL7ListenerError("name must not contain certain characters: `!@#$%^&*()+\\-=\\[\\]{};':\"\\\\|,.<>\\/?~.")
	}

	for _, override := range out.MSHOverrides {
		if !override.valid() {
			return ValidatedListenerOptions{}, NewHL7ListenerError("mshOverrides override value must be a string or a function.")
		}
	}

	if properties.Port == nil {
		return ValidatedListenerOptions{}, NewHL7ListenerError("port is not defined.")
	}
	out.Port = *properties.Port

	if err := clientutils.AssertNumber(float64(out.Port), "port", 0, 65_353); err != nil {
		return ValidatedListenerOptions{}, err
	}

	// Version is required and must be one of the known HL7 versions. Each port
	// enforces its own version; an inbound message whose MSH.12 differs is
	// rejected with an AR ACK (an intentional divergence from node-hl7).
	if out.Version == "" {
		return ValidatedListenerOptions{}, NewHL7ListenerError("version is not defined.")
	}
	if !metadata.IsKnownVersion(out.Version) {
		return ValidatedListenerOptions{}, NewHL7ListenerError("version is not a valid HL7 version.")
	}

	return out, nil
}

// NormalizeServerOptions validates and fills the server options, mirroring
// the normalizeServerOptions including the dual-stack ipv4/ipv6 semantics
// and bindAddress family validation.
func NormalizeServerOptions(properties *ServerOptions) (NormalizedServerOptions, error) {
	if properties == nil {
		properties = &ServerOptions{}
	}

	ipv4 := true
	ipv6 := false
	if properties.IPv4 != nil {
		ipv4 = *properties.IPv4
	}
	if properties.IPv6 != nil {
		ipv6 = *properties.IPv6
	}

	// Backward-compatible: only one of ipv4/ipv6 set true is "that family only".
	rawIPv4 := properties.IPv4 != nil
	rawIPv6 := properties.IPv6 != nil
	if rawIPv4 && *properties.IPv4 && !rawIPv6 {
		ipv6 = false
	}
	if rawIPv6 && *properties.IPv6 && !rawIPv4 {
		ipv4 = false
	}

	if !ipv4 && !ipv6 {
		return NormalizedServerOptions{}, NewHL7ServerError("ipv4 and ipv6 cannot both be disabled — at least one address family must be enabled.")
	}

	dualStack := ipv4 && ipv6
	ipv6Only := !ipv4 && ipv6

	bindAddress := ""
	if properties.BindAddress == nil {
		if dualStack || ipv6Only {
			bindAddress = "::"
		} else {
			bindAddress = "0.0.0.0"
		}
	} else {
		bindAddress = *properties.BindAddress
	}

	if bindAddress != "localhost" {
		switch {
		case dualStack:
			if len(bindAddress) > 0 && !clientutils.ValidIPv4(bindAddress) && !clientutils.ValidIPv6(bindAddress) {
				return NormalizedServerOptions{}, NewHL7ServerError("bindAddress is not a valid IPv4 or IPv6 address.")
			}
		case ipv6Only:
			if !clientutils.ValidIPv6(bindAddress) {
				return NormalizedServerOptions{}, NewHL7ServerError("bindAddress is an invalid ipv6 address.")
			}
		default:
			if !clientutils.ValidIPv4(bindAddress) {
				return NormalizedServerOptions{}, NewHL7ServerError("bindAddress is an invalid ipv4 address.")
			}
		}
	}

	encoding := properties.Encoding
	if encoding == "" {
		encoding = "utf8"
	}

	return NormalizedServerOptions{
		BindAddress: bindAddress,
		Encoding:    encoding,
		IPv4:        ipv4,
		IPv6:        ipv6,
		IPv6Only:    ipv6Only,
		TLS:         properties.TLS,
	}, nil
}
