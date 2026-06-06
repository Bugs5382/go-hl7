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

import "github.com/Bugs5382/go-hl7/client/builder"

// TLSConfig configures client-side TLS. A non-nil *TLSConfig (including the
// empty &TLSConfig{}) enables TLS; the empty value uses Go's defaults.
type TLSConfig struct {
	// Cert is the PEM-encoded certificate.
	Cert []byte
	// Key is the PEM-encoded private key.
	Key []byte
	// CA is an optional PEM-encoded certificate authority bundle to trust
	// instead of the system roots, used for self-signed certs.
	CA []byte
	// RejectUnauthorized, when set to a false value, skips peer-certificate
	// verification. nil leaves Go's default verification on.
	RejectUnauthorized *bool
	// ServerName overrides the SNI/verification host name.
	ServerName string
}

// ClientOptions configures a Client. Pointer fields distinguish "not provided"
// (nil) from an explicit value, which the IPv4/IPv6 dual-stack defaults rely
// on.
type ClientOptions struct {
	// AutoSelectFamily enables Happy-Eyeballs dual-stack racing when neither
	// family is exclusive (default true).
	AutoSelectFamily *bool
	// AutoSelectFamilyAttemptTimeout is the per-family attempt timeout in ms
	// (default 250).
	AutoSelectFamilyAttemptTimeout *int
	// ConnectionTimeout, in ms, ends and retries a stalled connection; 0 stays
	// connected (default 0).
	ConnectionTimeout *int
	// Host is the FQDN or IPv4/IPv6 address to connect to.
	Host string
	// IPv4 enables the IPv4 family (default true).
	IPv4 *bool
	// IPv6 enables the IPv6 family (default false).
	IPv6 *bool
	// MaxAttempts caps message-send retries while reconnecting (default 10,
	// range 1..50).
	MaxAttempts *int
	// MaxConnectionAttempts caps initial connection attempts (default 10,
	// range 1..50).
	MaxConnectionAttempts *int
	// MaxTimeout caps connection-timeout occurrences before giving up
	// (default 10).
	MaxTimeout *int
	// RetryHigh is the max backoff delay in ms (default 30000).
	RetryHigh *int
	// RetryLow is the backoff step in ms (default 1000).
	RetryLow *int
	// TLS enables and configures TLS; non-nil turns it on.
	TLS *TLSConfig
	// Version is the REQUIRED HL7 version every message sent over this client's
	// connections must declare in MSH.12. It must be one of the known HL7
	// versions (2.1, 2.2, 2.3, 2.3.1, 2.4, 2.5, 2.5.1, 2.6, 2.7, 2.7.1, 2.8).
	// Each client is pinned to a single version; outgoing messages whose
	// MSH.12 differs are rejected before they are sent.
	Version string
}

// ClientListenerOptions configures a single connection (port) and may override
// the client defaults.
type ClientListenerOptions struct {
	// AutoConnect connects immediately on creation when true; otherwise the
	// caller must call Connect (default true).
	AutoConnect *bool
	// Encoding is retained for API parity; HL7 bodies are UTF-8 byte slices so
	// it is informational only (default "utf8").
	Encoding string
	// EnqueueMessage is the custom queue-store hook; pairs with FlushQueue.
	// nil uses the default in-memory queue.
	EnqueueMessage func(message MessageItem, notifyPendingCount NotifyPendingCount) error
	// ExtendMaxLimit, when true, lets the in-memory queue grow past MaxLimit
	// instead of dropping the oldest.
	ExtendMaxLimit *bool
	// FlushQueue is the custom queue-drain hook; pairs with EnqueueMessage.
	// nil uses the default in-memory queue.
	FlushQueue func(callback FallBackHandler, notifyPendingCount NotifyPendingCount) error
	// MaxAttempts overrides the client MaxAttempts for this port.
	MaxAttempts *int
	// MaxConnectionAttempts overrides the client value for this port.
	MaxConnectionAttempts *int
	// MaxLimit caps the in-memory pending queue (default 10000).
	MaxLimit *int
	// NotifyOnLimitExceeded emits client.limitExceeded when the queue overflows.
	NotifyOnLimitExceeded *bool
	// Port is the remote port to connect to (required).
	Port *int
	// RetryHigh overrides the client backoff ceiling for this port.
	RetryHigh *int
	// RetryLow overrides the client backoff step for this port.
	RetryLow *int
	// WaitAck serializes sends on the previous ACK when true (default true).
	WaitAck *bool
}

// MessageItem is anything the connection can send: any value with a String()
// method that returns the message body the codec frames. *Message, *Batch, and
// *FileBatch all satisfy it.
type MessageItem interface {
	String() string
}

// FallBackHandler delivers a queued message back into the connection on flush.
type FallBackHandler func(message MessageItem)

// NotifyPendingCount reports the current pending-queue depth.
type NotifyPendingCount func(count int) error

// OutboundHandler receives a parsed ACK/response.
type OutboundHandler func(res *InboundResponse) error

// validatedClientOptions is the fully-resolved client option set.
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
	// version is the validated, required HL7 version pinned to this client; every
	// message sent over its connections must declare it in MSH.12.
	version string
}

// validatedClientListenerOptions is the fully-resolved per-port option set.
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
