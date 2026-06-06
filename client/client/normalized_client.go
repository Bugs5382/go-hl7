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
	"regexp"
	"strings"

	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/hl7/metadata"
	"github.com/Bugs5382/go-hl7/client/utils"
)

// Default client option values, mirroring the DEFAULT_CLIENT_OPTS.
const (
	defaultAutoSelectFamily               = true
	defaultAutoSelectFamilyAttemptTimeout = 250
	defaultConnectionTimeout              = 0
	defaultClientIPv4                     = true
	defaultClientIPv6                     = false
	defaultMaxAttempts                    = 10
	defaultMaxConnectionAttempts          = 10
	defaultMaxTimeout                     = 10
	defaultRetryHigh                      = 30_000
	defaultRetryLow                       = 1000
)

// Default per-port listener option values, mirroring the
// DEFAULT_LISTEN_CLIENT_OPTS.
const (
	defaultAutoConnect = true
	defaultMaxLimit    = 10_000
	defaultWaitAck     = true
	defaultEncoding    = "utf8"
)

var (
	looksLikeIPv4RE = regexp.MustCompile(`^[0-9.]+$`)
)

// boolOr returns *p when non-nil, else def, standing in for the
// `{ ...DEFAULTS, ...raw }` merge for a single bool field.
func boolOr(p *bool, def bool) bool {
	if p != nil {
		return *p
	}
	return def
}

// intOr returns *p when non-nil, else def.
func intOr(p *int, def int) int {
	if p != nil {
		return *p
	}
	return def
}

// normalizeClientListenerOptions validates and fills the per-port options,
// mirroring the normalizeClientListenerOptions. It returns an error where
// the spec throws (HL7FatalError) or assertNumber throws (plain error).
func normalizeClientListenerOptions(client validatedClientOptions, raw ClientListenerOptions) (validatedClientListenerOptions, error) {
	out := validatedClientListenerOptions{
		autoConnect:           boolOr(raw.AutoConnect, defaultAutoConnect),
		encoding:              raw.Encoding,
		enqueueMessage:        raw.EnqueueMessage,
		extendMaxLimit:        boolOr(raw.ExtendMaxLimit, false),
		flushQueue:            raw.FlushQueue,
		maxAttempts:           intOr(raw.MaxAttempts, defaultMaxAttempts),
		maxConnectionAttempts: intOr(raw.MaxConnectionAttempts, defaultMaxConnectionAttempts),
		maxLimit:              intOr(raw.MaxLimit, defaultMaxLimit),
		notifyOnLimitExceeded: boolOr(raw.NotifyOnLimitExceeded, false),
		retryHigh:             intOr(raw.RetryHigh, client.retryHigh),
		retryLow:              intOr(raw.RetryLow, client.retryLow),
		waitAck:               boolOr(raw.WaitAck, defaultWaitAck),
	}
	if out.encoding == "" {
		out.encoding = defaultEncoding
	}

	// Reject a missing port: `if (port === undefined) throw "port is not
	// defined."`. Go can't tell a string-typed port from a number at the type
	// level (the "port is not valid number." path is a TypeScript-only check),
	// so a nil pointer is the "not defined" case.
	if raw.Port == nil {
		return validatedClientListenerOptions{}, helpers.NewHL7FatalError("port is not defined.")
	}
	out.port = *raw.Port

	// enqueueMessage / flushQueue must be set together (the paired check).
	if raw.EnqueueMessage != nil && raw.FlushQueue == nil {
		return validatedClientListenerOptions{}, helpers.NewHL7FatalError("flushQueue is not set.")
	}
	if raw.EnqueueMessage == nil && raw.FlushQueue != nil {
		return validatedClientListenerOptions{}, helpers.NewHL7FatalError("enqueueMessage is not set.")
	}

	if err := utils.AssertNumber(float64(out.maxLimit), "maxLimit", 1); err != nil {
		return validatedClientListenerOptions{}, err
	}
	if err := utils.AssertNumber(float64(out.maxAttempts), "maxAttempts", 1, 50); err != nil {
		return validatedClientListenerOptions{}, err
	}
	if err := utils.AssertNumber(float64(out.maxConnectionAttempts), "maxConnectionAttempts", 1, 50); err != nil {
		return validatedClientListenerOptions{}, err
	}
	if err := utils.AssertNumber(float64(out.port), "port", 1, 65_353); err != nil {
		return validatedClientListenerOptions{}, err
	}

	return out, nil
}

// normalizeClientOptions validates and fills the client options, mirroring
// the normalizeClientOptions including the ipv4/ipv6 dual-stack semantics
// and IP-literal family validation.
func normalizeClientOptions(raw ClientOptions) (validatedClientOptions, error) {
	out := validatedClientOptions{
		autoSelectFamily:               boolOr(raw.AutoSelectFamily, defaultAutoSelectFamily),
		autoSelectFamilyAttemptTimeout: intOr(raw.AutoSelectFamilyAttemptTimeout, defaultAutoSelectFamilyAttemptTimeout),
		connectionTimeout:              intOr(raw.ConnectionTimeout, defaultConnectionTimeout),
		host:                           raw.Host,
		ipv4:                           boolOr(raw.IPv4, defaultClientIPv4),
		ipv6:                           boolOr(raw.IPv6, defaultClientIPv6),
		maxTimeout:                     intOr(raw.MaxTimeout, defaultMaxTimeout),
		retryHigh:                      intOr(raw.RetryHigh, defaultRetryHigh),
		retryLow:                       intOr(raw.RetryLow, defaultRetryLow),
		tls:                            raw.TLS,
		version:                        raw.Version,
	}

	// Backward-compatible semantics: passing only one of ipv4/ipv6 explicitly
	// (true) is "that family only". The spec checks hasOwnProperty; the pointer's
	// non-nil-ness is the Go equivalent of the key being present.
	rawIPv4 := raw.IPv4 != nil
	rawIPv6 := raw.IPv6 != nil
	if rawIPv4 && *raw.IPv4 && !rawIPv6 {
		out.ipv6 = false
	}
	if rawIPv6 && *raw.IPv6 && !rawIPv4 {
		out.ipv4 = false
	}

	if out.host == "" {
		return validatedClientOptions{}, helpers.NewHL7FatalError("host is not defined or the length is less than 0.")
	}

	if !out.ipv4 && !out.ipv6 {
		return validatedClientOptions{}, helpers.NewHL7FatalError("ipv4 and ipv6 cannot both be disabled — at least one address family must be enabled.")
	}

	// Detect whether host is an IP literal or an FQDN. Literals validate against
	// the requested family; FQDNs defer to DNS at connect time.
	literalFamily := utils.DetectIPFamily(out.host)
	looksLikeIPv4 := looksLikeIPv4RE.MatchString(out.host)
	looksLikeIPv6 := strings.Contains(out.host, ":")

	switch {
	case out.ipv4 && !out.ipv6:
		if looksLikeIPv4 && !utils.ValidIPv4(out.host) {
			return validatedClientOptions{}, helpers.NewHL7FatalError("host is not a valid IPv4 address.")
		}
		if literalFamily == 6 {
			return validatedClientOptions{}, helpers.NewHL7FatalError("host is not a valid IPv4 address.")
		}
		out.family = 4
	case !out.ipv4 && out.ipv6:
		if looksLikeIPv6 && !utils.ValidIPv6(out.host) {
			return validatedClientOptions{}, helpers.NewHL7FatalError("host is not a valid IPv6 address.")
		}
		if literalFamily == 4 {
			return validatedClientOptions{}, helpers.NewHL7FatalError("host is not a valid IPv6 address.")
		}
		out.family = 6
	default:
		// dual-stack: family resolved at connect time (0 = unspecified)
		out.family = literalFamily
	}

	if err := utils.AssertNumber(float64(out.connectionTimeout), "connectionTimeout", 0, 60_000); err != nil {
		return validatedClientOptions{}, err
	}
	if err := utils.AssertNumber(float64(out.maxTimeout), "maxTimeout", 1, 50); err != nil {
		return validatedClientOptions{}, err
	}
	if err := utils.AssertNumber(float64(out.autoSelectFamilyAttemptTimeout), "autoSelectFamilyAttemptTimeout", 10, 60_000); err != nil {
		return validatedClientOptions{}, err
	}

	// Version is required and must be one of the known HL7 versions. This pins
	// the client to a single HL7 version; SendMessage rejects any message whose
	// MSH.12 differs (an intentional divergence from node-hl7).
	if out.version == "" {
		return validatedClientOptions{}, helpers.NewHL7FatalError("version is not defined.")
	}
	if !metadata.IsKnownVersion(out.version) {
		return validatedClientOptions{}, helpers.NewHL7FatalError("version is not a valid HL7 version.")
	}

	return out, nil
}
