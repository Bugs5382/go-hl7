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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/declaration"
	"github.com/Bugs5382/go-hl7/client/helpers"
	"github.com/Bugs5382/go-hl7/client/modules"
	"github.com/Bugs5382/go-hl7/client/utils"
)

// connectionStats mirrors the Connection.stats per-connection counters.
type connectionStats struct {
	acknowledged int
	pending      int
	sent         int
}

// Connection is one TCP/MLLP (optionally TLS) connection to a remote port. It
// auto-reconnects with exponential backoff, serializes sends on the ACK when
// waitAck is set, and queues messages while disconnected. It mirrors the reference's
// Connection (an EventEmitter) over the same event names via the embedded
// eventEmitter: connect, close, connection, open, connecting, client.sent,
// client.acknowledged, client.error, client.timeout, client.pending,
// client.limitExceeded, data.raw, data.error.
type Connection struct {
	eventEmitter

	handler OutboundHandler
	main    *Client
	opt     validatedClientListenerOptions

	mu               sync.Mutex
	readyState       declaration.ReadyState
	awaitingResponse bool
	codec            *modules.MLLPCodec
	socket           net.Conn
	// socketGen increments each dial so a stale socket's read/close goroutine
	// can detect it is no longer the live socket (relies on closure
	// identity; Go uses a generation counter).
	socketGen         int
	pendingMessages   []MessageItem
	pendingSetup      bool
	retryCount        int
	retryTimeoutCount int
	retryTimer        *time.Timer
	connectionTimer   *time.Timer
	onConnect         *declaration.Deferred[struct{}]
	// lastError holds the most recent socket error for the close handler to
	// surface (the connectionError local).
	lastError error

	maxLimit              int
	extendMaxLimit        bool
	notifyOnLimitExceeded bool
	enqueueMessageFn      func(message MessageItem, notifyPendingCount NotifyPendingCount) error
	flushQueueFn          func(callback FallBackHandler, notifyPendingCount NotifyPendingCount) error

	stats connectionStats
}

// newConnection constructs and (when autoConnect) starts a connection,
// mirroring the Connection constructor.
func newConnection(client *Client, properties ClientListenerOptions, handler OutboundHandler) (*Connection, error) {
	opt, err := normalizeClientListenerOptions(client.opt, properties)
	if err != nil {
		return nil, err
	}

	c := &Connection{
		handler:               handler,
		main:                  client,
		opt:                   opt,
		maxLimit:              opt.maxLimit,
		extendMaxLimit:        opt.extendMaxLimit,
		notifyOnLimitExceeded: opt.notifyOnLimitExceeded,
		onConnect:             declaration.NewDeferred[struct{}](),
	}

	c.enqueueMessageFn = opt.enqueueMessage
	if c.enqueueMessageFn == nil {
		c.enqueueMessageFn = c.defaultEnqueueMessage
	}
	c.flushQueueFn = opt.flushQueue
	if c.flushQueueFn == nil {
		c.flushQueueFn = c.defaultFlushQueue
	}

	if opt.autoConnect {
		c.mu.Lock()
		c.readyState = declaration.Connecting
		c.mu.Unlock()
		c.emit("connecting")
		c.connect()
	} else {
		c.mu.Lock()
		c.readyState = declaration.Open
		c.mu.Unlock()
		c.emit("open")
	}

	return c, nil
}

// Close force-closes the connection, stopping reconnection timers. It mirrors
// the close(): a CLOSING connection waits for the socket close, a CONNECTING
// one clears its retry timer first. Restarting requires a fresh Connect.
func (c *Connection) Close() error {
	c.mu.Lock()
	switch c.readyState {
	case declaration.Closed:
		c.mu.Unlock()
		return nil
	case declaration.Closing:
		sock := c.socket
		c.mu.Unlock()
		if sock != nil {
			_ = sock.Close()
		}
		return nil
	case declaration.Connecting:
		if c.retryTimer != nil {
			c.retryTimer.Stop()
		}
		c.retryTimer = nil
	}

	c.readyState = declaration.Closing
	sock := c.socket
	if c.connectionTimer != nil {
		c.connectionTimer.Stop()
	}
	c.mu.Unlock()

	if sock != nil {
		_ = sock.Close()
	}

	c.emit("close")

	c.mu.Lock()
	c.readyState = declaration.Closed
	c.mu.Unlock()
	return nil
}

// Connect starts the connection if it was not auto-started, mirroring the
// connect(). It is a no-op while already connecting/connected/open.
func (c *Connection) Connect() error {
	c.mu.Lock()
	state := c.readyState
	c.mu.Unlock()
	switch state {
	case declaration.Connecting, declaration.Connected, declaration.Open:
		return nil
	case declaration.Closing:
		c.mu.Lock()
		sock := c.socket
		c.mu.Unlock()
		if sock != nil {
			// wait for the socket to close
			_ = sock.Close()
		}
		return nil
	}

	c.emit("connecting")
	c.connect()
	return nil
}

// GetPort returns the port this connection dials, mirroring the getPort.
func (c *Connection) GetPort() int { return c.opt.port }

// IsConnected reports whether the connection has reached the CONNECTED state.
// It lets a caller that registered its "connect" handler after the asynchronous
// emit still observe readiness (the Go analog of the already-resolved
// connect Promise in the test helpers).
func (c *Connection) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.readyState == declaration.Connected
}

// SendMessage frames and sends a message to the remote side. When not ready
// (disconnected, or waitAck with an outstanding ACK) it queues the message and
// kicks off a connection attempt. It mirrors the sendMessage including the
// maxAttempts retry loop and the waitAck serialization gate.
func (c *Connection) SendMessage(message MessageItem) error {
	theMessage := message.String()
	// The codec's only argument is the return character (the \r default); the
	// encoding the source threaded through is dropped since Go bodies are UTF-8
	// byte slices. Passing encoding here would wrongly set the join character.
	codec := modules.NewMLLPCodec("")
	maxAttempts := c.opt.maxAttempts
	attempts := 0

	c.mu.Lock()
	shouldQueue := c.readyState != declaration.Connected || (c.opt.waitAck && c.awaitingResponse)
	c.mu.Unlock()

	if shouldQueue {
		if err := c.enqueueMessageFn(message, c.handlePendingUpdate); err != nil {
			return err
		}

		c.mu.Lock()
		startSetup := !c.pendingSetup
		if startSetup {
			c.pendingSetup = true
		}
		c.mu.Unlock()

		if startSetup {
			c.connect()
		}
		return nil
	}

	for {
		c.mu.Lock()
		ready := c.readyState == declaration.Connected
		c.mu.Unlock()
		if ready {
			break
		}
		attempts++
		if attempts >= maxAttempts {
			return helpers.NewHL7FatalError("In an invalid state to send message.")
		}
	}

	c.mu.Lock()
	if c.opt.waitAck {
		c.awaitingResponse = true
	}
	sock := c.socket
	c.mu.Unlock()

	if err := codec.SendMessage(sock, theMessage); err != nil {
		return err
	}

	c.mu.Lock()
	c.stats.sent++
	sent := c.stats.sent
	c.mu.Unlock()
	c.emit("client.sent", sent)
	return nil
}

// connect dials the socket (TCP or TLS) and wires the connect/data/close/error
// handling onto a read goroutine, mirroring the _connect. The dial runs in a
// goroutine so the events fire asynchronously the way the libuv connect does.
func (c *Connection) connect() {
	host := c.main.opt.host
	port := c.opt.port
	family := c.main.opt.family

	// Strip surrounding brackets from IPv6 literals (accepts both forms).
	dialHost := host
	if strings.HasPrefix(dialHost, "[") && strings.HasSuffix(dialHost, "]") {
		dialHost = dialHost[1 : len(dialHost)-1]
	}

	c.mu.Lock()
	c.retryTimer = nil
	c.codec = modules.NewMLLPCodec("")
	gen := c.socketGen + 1
	c.socketGen = gen
	connTimeout := c.main.opt.connectionTimeout
	state := c.readyState
	c.mu.Unlock()

	// the spec arms a connection timeout when timeout > 0 and we are
	// connecting/connected; it destroys the socket on fire and stops after
	// maxTimeout occurrences.
	if connTimeout > 0 && (state == declaration.Connected || state == declaration.Connecting) {
		c.mu.Lock()
		if c.retryTimeoutCount < c.main.opt.maxTimeout {
			c.connectionTimer = time.AfterFunc(time.Duration(connTimeout)*time.Millisecond, func() {
				c.mu.Lock()
				c.retryTimeoutCount++
				sock := c.socket
				c.mu.Unlock()
				c.emit("client.timeout")
				if sock != nil {
					_ = sock.Close()
				}
			})
			c.mu.Unlock()
		} else {
			c.mu.Unlock()
			_ = c.Close()
			return
		}
	}

	network := familyNetwork(family)
	address := net.JoinHostPort(dialHost, strconv.Itoa(port))

	go func() {
		var (
			conn net.Conn
			err  error
		)
		dialer := &net.Dialer{}
		if c.main.opt.tls == nil {
			conn, err = dialer.Dial(network, address)
		} else {
			tlsCfg := buildTLSConfig(c.main.opt.tls, dialHost)
			conn, err = tls.DialWithDialer(dialer, network, address, tlsCfg)
		}
		if err != nil {
			c.handleSocketError(err)
			c.handleSocketClose(gen, err)
			return
		}

		if tcp, ok := conn.(*net.TCPConn); ok {
			_ = tcp.SetNoDelay(true)
		}

		c.mu.Lock()
		c.socket = conn
		c.mu.Unlock()

		c.handleConnect(conn)
		c.readLoop(conn, gen)
	}()
}

// familyNetwork maps the resolved family (0/4/6) to a Go dial network.
func familyNetwork(family int) string {
	switch family {
	case 4:
		return "tcp4"
	case 6:
		return "tcp6"
	default:
		return "tcp"
	}
}

// buildTLSConfig maps the TLSConfig to a *tls.Config, mirroring the
// tls.connect option pass-through (rejectUnauthorized, ca, servername, cert).
func buildTLSConfig(cfg *TLSConfig, dialHost string) *tls.Config {
	out := &tls.Config{ServerName: dialHost}
	if cfg.ServerName != "" {
		out.ServerName = cfg.ServerName
	}
	if cfg.RejectUnauthorized != nil && !*cfg.RejectUnauthorized {
		out.InsecureSkipVerify = true
	}
	if len(cfg.CA) > 0 {
		pool := x509.NewCertPool()
		if pool.AppendCertsFromPEM(cfg.CA) {
			out.RootCAs = pool
		}
	}
	if len(cfg.Cert) > 0 && len(cfg.Key) > 0 {
		if pair, err := tls.X509KeyPair(cfg.Cert, cfg.Key); err == nil {
			out.Certificates = append(out.Certificates, pair)
		}
	}
	return out
}

// handleConnect runs the socket "connect" handler: mark CONNECTED, reset the
// retry count, flush the pending queue, then negotiate (open/connection) and
// emit connect.
func (c *Connection) handleConnect(conn net.Conn) {
	c.mu.Lock()
	c.readyState = declaration.Connected
	c.retryCount = 1
	c.pendingSetup = false
	c.mu.Unlock()

	// flush queue: deliver pending messages back into sendMessage.
	_ = c.flushQueueFn(func(message MessageItem) {
		go func() { _ = c.SendMessage(message) }()
	}, c.handlePendingUpdate)

	c.emit("connect")

	// negotiate: once writable we are OPEN, resolve onConnect, emit connection.
	c.mu.Lock()
	c.readyState = declaration.Connected
	c.mu.Unlock()
	c.onConnect.Resolve(struct{}{})
	c.emit("connection")
}

// readLoop reads framed bytes off the socket, feeds the codec, and dispatches
// completed messages to the handler. It mirrors the socket "data" handler.
// It returns (and triggers the close handling) when the socket ends/errors.
func (c *Connection) readLoop(conn net.Conn, gen int) {
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			c.onData(buf[:n])
		}
		if err != nil {
			c.handleSocketClose(gen, err)
			return
		}
	}
}

// onData feeds a chunk into the codec and, once a full frame is buffered,
// parses the response(s) and invokes the handler. It mirrors the data
// handler (receiveData -> getLastMessage -> Batch|Message -> InboundResponse).
func (c *Connection) onData(chunk []byte) {
	c.mu.Lock()
	codec := c.codec
	c.mu.Unlock()
	if codec == nil {
		return
	}

	var dataResult bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				c.emit("data.error", asError(r))
			}
		}()
		dataResult = codec.ReceiveData(chunk)
	}()

	if !dataResult {
		return
	}

	// We got a response (good/bad/error); clear the awaiting gate.
	c.mu.Lock()
	c.awaitingResponse = false
	c.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			c.emit("data.error", asError(r))
		}
	}()

	loaded := codec.GetLastMessage()
	if loaded == nil {
		return
	}
	completed := *loaded

	c.emit("data.raw", completed)

	if c.handler == nil {
		return
	}

	if utils.IsBatch(completed) {
		// A batched response is parsed and each message delivered individually so
		// every ACK reaches the handler, mirroring Batch.messages().
		parser, err := builder.NewBatch(builder.BatchOptions{Text: completed})
		if err != nil {
			c.emit("data.error", err)
			return
		}
		for _, msg := range parser.Messages() {
			c.deliver(msg.String())
		}
		return
	}
	c.deliver(completed)
}

// deliver parses one message body into an InboundResponse and hands it to the
// outbound handler, bumping the acknowledged counter. It mirrors the per-message
// tail of the data handler.
func (c *Connection) deliver(msgText string) {
	parsed, err := builder.NewMessage(builder.MessageOptions{Text: msgText})
	if err != nil {
		c.emit("data.error", err)
		return
	}

	c.mu.Lock()
	c.stats.acknowledged++
	ack := c.stats.acknowledged
	c.mu.Unlock()

	response, err := NewInboundResponse(parsed.String())
	if err != nil {
		c.emit("data.error", err)
		return
	}
	c.emit("client.acknowledged", ack)
	_ = c.handler(response)
}

// handleSocketError records the first socket error, mirroring the socket
// "error" handler which keeps only the first error and surfaces the code as
// client.error from the close handler. Here we emit client.error directly with
// the underlying error so callers can inspect its code (ECONNREFUSED maps to a
// connection-refused error string).
func (c *Connection) handleSocketError(err error) {
	c.mu.Lock()
	c.lastError = err
	c.mu.Unlock()
}

// handleSocketClose runs the socket "close" handler: depending on the ready
// state it either finalizes CLOSED or schedules an exponential-backoff
// reconnect (up to maxConnectionAttempts), emitting client.error.
func (c *Connection) handleSocketClose(gen int, cause error) {
	c.mu.Lock()
	// Ignore a close from a stale (superseded) socket.
	if gen != 0 && gen != c.socketGen {
		c.mu.Unlock()
		return
	}

	if c.readyState == declaration.Closing {
		c.readyState = declaration.Closed
		c.mu.Unlock()
		return
	}

	if c.readyState == declaration.Closed {
		c.mu.Unlock()
		return
	}

	connTimeout := c.main.opt.connectionTimeout
	// the spec only reconnects when connectionTimeout > 0; with timeout 0 it stays
	// closed after the socket drops (but still surfaces the error once).
	connErr := c.lastError
	if connErr == nil {
		connErr = cause
	}
	if connErr == nil {
		connErr = helpers.NewHL7FatalError("Socket closed unexpectedly by server.")
	}
	c.lastError = nil

	if connTimeout <= 0 {
		// Surface the connection error so callers (and the failure tests) see
		// it, then stay where we are without reconnecting.
		wasConnecting := c.readyState == declaration.Connecting
		c.mu.Unlock()
		if wasConnecting {
			c.emit("client.error", normalizeDialError(connErr))
		}
		return
	}

	if c.readyState == declaration.Open {
		c.onConnect = declaration.NewDeferred[struct{}]()
	}
	c.readyState = declaration.Connecting
	retryCount := c.retryCount
	c.retryCount++
	maxConn := c.opt.maxConnectionAttempts
	delay := utils.ExpBackoff(c.opt.retryLow, c.opt.retryHigh, retryCount, 2)

	if retryCount < maxConn {
		c.retryTimer = time.AfterFunc(time.Duration(delay)*time.Millisecond, c.connect)
		c.mu.Unlock()
		c.emit("client.error", normalizeDialError(connErr))
	} else if retryCount > maxConn {
		c.mu.Unlock()
		_ = c.Close()
	} else {
		c.mu.Unlock()
	}
}

// normalizeDialError surfaces a connection-refused error with a stable
// ECONNREFUSED marker so callers can match it the way the spec exposes error.code.
func normalizeDialError(err error) error {
	if err == nil {
		return err
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if strings.Contains(strings.ToLower(opErr.Err.Error()), "refused") {
			return &dialError{err: err, code: "ECONNREFUSED"}
		}
	}
	if strings.Contains(strings.ToLower(err.Error()), "refused") {
		return &dialError{err: err, code: "ECONNREFUSED"}
	}
	return err
}

// dialError wraps a dial failure with an errno-style code string so the
// failure tests can assert error.code == "ECONNREFUSED".
type dialError struct {
	err  error
	code string
}

func (e *dialError) Error() string { return e.err.Error() }
func (e *dialError) Unwrap() error { return e.err }

// Code returns the errno-style code (e.g. "ECONNREFUSED").
func (e *dialError) Code() string { return e.code }

// defaultEnqueueMessage is the in-memory queue store, mirroring the
// defaultEnqueueMessage including the overflow handling at maxLimit.
func (c *Connection) defaultEnqueueMessage(message MessageItem, notifyPendingCount NotifyPendingCount) error {
	c.mu.Lock()
	if len(c.pendingMessages) == c.maxLimit {
		c.handleQueueOverflow()
	}
	c.pendingMessages = append(c.pendingMessages, message)
	count := len(c.pendingMessages)
	c.mu.Unlock()
	return notifyPendingCount(count)
}

// defaultFlushQueue drains the in-memory queue back into the connection,
// mirroring the defaultFlushQueue.
func (c *Connection) defaultFlushQueue(callback FallBackHandler, notifyPendingCount NotifyPendingCount) error {
	for {
		c.mu.Lock()
		if len(c.pendingMessages) == 0 {
			c.mu.Unlock()
			return nil
		}
		message := c.pendingMessages[0]
		c.pendingMessages = c.pendingMessages[1:]
		count := len(c.pendingMessages)
		c.mu.Unlock()
		callback(message)
		if err := notifyPendingCount(count); err != nil {
			return err
		}
	}
}

// handleQueueOverflow drops the oldest message (unless extendMaxLimit) and
// optionally emits client.limitExceeded. Caller holds c.mu. Mirrors the
// handleQueueOverflow.
func (c *Connection) handleQueueOverflow() {
	if !c.extendMaxLimit {
		c.pendingMessages = c.pendingMessages[1:]
	}
	if c.notifyOnLimitExceeded {
		// emit without the lock held to avoid re-entrancy; capture the port.
		port := c.opt.port
		go c.emit("client.limitExceeded", port)
	}
}

// handlePendingUpdate records the pending depth and emits client.pending,
// mirroring the _handlePendingUpdate.
func (c *Connection) handlePendingUpdate(count int) error {
	c.mu.Lock()
	c.stats.pending = count
	c.mu.Unlock()
	c.emit("client.pending", count)
	return nil
}

// asError coerces a recovered panic value into an error for data.error.
func asError(r any) error {
	if err, ok := r.(error); ok {
		return err
	}
	return helpers.NewHL7FatalErrorf("%v", r)
}
