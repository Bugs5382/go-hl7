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

import "github.com/Bugs5382/go-hl7/client/builder"

// TLSConfig mirrors the subset of the tls.ConnectionOptions the client and
// its tests use. A non-nil *TLSConfig (including the empty &TLSConfig{}) enables
// TLS, standing in for the `tls: true | TLSOptions` union: passing the
// boolean `true` becomes &TLSConfig{}.
type TLSConfig struct {
	// Cert is the PEM-encoded certificate (the tls `cert`).
	Cert []byte
	// Key is the PEM-encoded private key (the tls `key`).
	Key []byte
	// CA is an optional PEM-encoded certificate authority bundle to trust
	// instead of the system roots (the tls `ca`), used for self-signed certs.
	CA []byte
	// RejectUnauthorized, when set to a false value, skips peer-certificate
	// verification (the tls `rejectUnauthorized: false`). nil leaves Go's
	// default verification on.
	RejectUnauthorized *bool
	// ServerName overrides the SNI/verification host name (the tls
	// `servername`).
	ServerName string
}

// ClientOptions mirrors the ClientOptions. Pointer fields distinguish
// "not provided" (nil) from an explicit value, matching the
// hasOwnProperty checks for the ipv4/ipv6 dual-stack semantics.
type ClientOptions struct {
	// AutoSelectFamily enables Happy-Eyeballs dual-stack racing when neither
	// family is exclusive (autoSelectFamily, default true).
	AutoSelectFamily *bool
	// AutoSelectFamilyAttemptTimeout is the per-family attempt timeout in ms
	// (autoSelectFamilyAttemptTimeout, default 250).
	AutoSelectFamilyAttemptTimeout *int
	// ConnectionTimeout, in ms, ends and retries a stalled connection; 0 stays
	// connected (connectionTimeout, default 0).
	ConnectionTimeout *int
	// Host is the FQDN or IPv4/IPv6 address to connect to (host).
	Host string
	// IPv4 enables the IPv4 family (ipv4, default true).
	IPv4 *bool
	// IPv6 enables the IPv6 family (ipv6, default false).
	IPv6 *bool
	// MaxAttempts caps message-send retries while reconnecting (
	// maxAttempts, default 10, range 1..50).
	MaxAttempts *int
	// MaxConnectionAttempts caps initial connection attempts (
	// maxConnectionAttempts, default 10, range 1..50).
	MaxConnectionAttempts *int
	// MaxTimeout caps connection-timeout occurrences before giving up (
	// maxTimeout, default 10).
	MaxTimeout *int
	// RetryHigh is the max backoff delay in ms (retryHigh, default 30000).
	RetryHigh *int
	// RetryLow is the backoff step in ms (retryLow, default 1000).
	RetryLow *int
	// TLS enables and configures TLS; non-nil turns it on (tls).
	TLS *TLSConfig
}

// ClientListenerOptions mirrors the ClientListenerOptions (the per-port
// createConnection options that may override the client defaults).
type ClientListenerOptions struct {
	// AutoConnect connects immediately on creation when true; otherwise the
	// caller must call Connect (autoConnect, default true).
	AutoConnect *bool
	// Encoding is retained for API parity; Go HL7 bodies are UTF-8 byte
	// slices so it is informational only (encoding, default "utf8").
	Encoding string
	// EnqueueMessage is the custom queue-store hook; pairs with FlushQueue
	// (enqueueMessage). nil uses the default in-memory queue.
	EnqueueMessage func(message MessageItem, notifyPendingCount NotifyPendingCount) error
	// ExtendMaxLimit, when true, lets the in-memory queue grow past MaxLimit
	// instead of dropping the oldest (extendMaxLimit).
	ExtendMaxLimit *bool
	// FlushQueue is the custom queue-drain hook; pairs with EnqueueMessage
	// (flushQueue). nil uses the default in-memory queue.
	FlushQueue func(callback FallBackHandler, notifyPendingCount NotifyPendingCount) error
	// MaxAttempts overrides the client MaxAttempts for this port (
	// maxAttempts).
	MaxAttempts *int
	// MaxConnectionAttempts overrides the client value for this port (
	// maxConnectionAttempts).
	MaxConnectionAttempts *int
	// MaxLimit caps the in-memory pending queue (maxLimit, default 10000).
	MaxLimit *int
	// NotifyOnLimitExceeded emits client.limitExceeded when the queue overflows
	// (notifyOnLimitExceeded).
	NotifyOnLimitExceeded *bool
	// Port is the remote port to connect to (port, required).
	Port *int
	// RetryHigh overrides the client backoff ceiling for this port.
	RetryHigh *int
	// RetryLow overrides the client backoff step for this port.
	RetryLow *int
	// WaitAck serializes sends on the previous ACK when true (waitAck,
	// default true).
	WaitAck *bool
}

// MessageItem is anything the connection can send. the MessageItem is
// `Batch | FileBatch | Message`; the contract is the String()-able message body
// the codec frames, which Message, Batch, and FileBatch all satisfy.
type MessageItem interface {
	String() string
}

// FallBackHandler delivers a queued message back into the connection on
// flush, mirroring the FallBackHandler.
type FallBackHandler func(message MessageItem)

// NotifyPendingCount reports the current pending-queue depth, mirroring the
// NotifyPendingCount.
type NotifyPendingCount func(count int) error

// OutboundHandler receives a parsed ACK/response, mirroring the
// OutboundHandler ((res) => Promise<void> | void).
type OutboundHandler func(res *InboundResponse) error

// validatedClientOptions is the fully-resolved client option set, mirroring
// the ValidatedClientOptions.
type validatedClientOptions struct {
	autoSelectFamily               bool
	autoSelectFamilyAttemptTimeout int
	connectionTimeout              int
	// family resolves the connection family: 0 dual-stack (let the resolver
	// decide), 4 or 6 to force that family.
	family     int
	host       string
	ipv4       bool
	ipv6       bool
	maxTimeout int
	retryHigh  int
	retryLow   int
	tls        *TLSConfig
}

// validatedClientListenerOptions is the fully-resolved per-port option set,
// mirroring the ValidatedClientListenerOptions.
type validatedClientListenerOptions struct {
	autoConnect           bool
	encoding              string
	enqueueMessage        func(message MessageItem, notifyPendingCount NotifyPendingCount) error
	extendMaxLimit        bool
	flushQueue            func(callback FallBackHandler, notifyPendingCount NotifyPendingCount) error
	maxAttempts           int
	maxConnectionAttempts int
	maxLimit              int
	notifyOnLimitExceeded bool
	port                  int
	retryHigh             int
	retryLow              int
	waitAck               bool
}

// compile-time guards that Message, Batch, and FileBatch satisfy MessageItem
// (each has String()).
var (
	_ MessageItem = (*builder.Message)(nil)
	_ MessageItem = (*builder.Batch)(nil)
	_ MessageItem = (*builder.FileBatch)(nil)
)
