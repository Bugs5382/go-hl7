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
	"crypto/tls"
	"crypto/x509"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/modules"
	"github.com/Bugs5382/go-hl7/client/utils"
	srvutils "github.com/Bugs5382/go-hl7/server/utils"
)

// InboundHandler processes one inbound message and sends an ACK via the
// response. Returning an error signals a handler failure.
type InboundHandler func(request *InboundRequest, res ResponseSender) error

// inboundStats holds the per-listener counters.
type inboundStats struct {
	received     int
	totalMessage int
}

// Inbound is one TCP/TLS listener bound to a port. It accepts connections,
// frames inbound HL7 with a per-socket MLLP codec, parses message/batch/file
// bodies, and dispatches each to the handler. It embeds EventEmitter and
// emits: listen, client.connect, client.close, client.error, data.raw,
// data.error, error, response.sent.
type Inbound struct {
	EventEmitter
	handler InboundHandler
	main    *Server
	opt     srvutils.ValidatedListenerOptions

	mu        sync.Mutex
	listener  net.Listener
	sockets   []net.Conn
	closed    bool
	listening bool
	stats     inboundStats
}

// IsListening reports whether the listener has bound and emitted (or is about
// to emit) "listen". It lets a caller that registered its handler after the
// asynchronous emit still observe readiness, the way the once + already
// settled Promise resolves for the test helpers.
func (in *Inbound) IsListening() bool {
	in.mu.Lock()
	defer in.mu.Unlock()
	return in.listening
}

// newInbound builds, validates, and starts a listener, returning an error on
// bad options.
func newInbound(server *Server, properties ListenerOptions, handler InboundHandler) (*Inbound, error) {
	opt, err := srvutils.NormalizeListenerOptions(properties)
	if err != nil {
		return nil, err
	}
	in := &Inbound{handler: handler, main: server, opt: opt}
	if err := in.listen(); err != nil {
		return nil, err
	}
	return in, nil
}

// GetName returns the resolved listener name.
func (in *Inbound) GetName() string { return in.opt.Name }

// TotalMessage returns the per-message parse counter.
func (in *Inbound) TotalMessage() int {
	in.mu.Lock()
	defer in.mu.Unlock()
	return in.stats.totalMessage
}

// TotalReceived returns the per-frame received counter.
func (in *Inbound) TotalReceived() int {
	in.mu.Lock()
	defer in.mu.Unlock()
	return in.stats.received
}

// Close destroys all client sockets and stops the listener.
func (in *Inbound) Close() error {
	in.mu.Lock()
	if in.closed {
		in.mu.Unlock()
		return nil
	}
	in.closed = true
	sockets := in.sockets
	in.sockets = nil
	listener := in.listener
	in.mu.Unlock()

	for _, socket := range sockets {
		_ = socket.Close()
	}
	if listener != nil {
		_ = listener.Close()
	}
	return nil
}

// listen binds the listener (TCP or TLS), emits listen once bound, and accepts
// connections in a goroutine. It includes the dual-stack IPv6 -> IPv4 fallback
// (when "::" cannot bind, retry on 0.0.0.0).
func (in *Inbound) listen() error {
	port := in.opt.Port
	bindAddress := in.main.opt.BindAddress
	dualStack := in.main.opt.IPv4 && in.main.opt.IPv6

	network := "tcp"
	host := bindAddress
	if host == "localhost" {
		host = "127.0.0.1"
		if in.main.opt.IPv6Only {
			host = "::1"
		}
	}

	switch {
	case in.main.opt.IPv6Only:
		network = "tcp6"
	case !dualStack && in.main.opt.IPv4:
		network = "tcp4"
	}

	address := net.JoinHostPort(host, strconv.Itoa(port))

	listener, err := in.bind(network, address)
	if err != nil && dualStack {
		// IPv6 wildcard unavailable: retry IPv4-only on 0.0.0.0.
		fallbackHost := bindAddress
		if fallbackHost == "::" {
			fallbackHost = "0.0.0.0"
		}
		listener, err = in.bind("tcp4", net.JoinHostPort(fallbackHost, strconv.Itoa(port)))
	}
	if err != nil {
		in.emit("error", err)
		return err
	}

	in.mu.Lock()
	in.listener = listener
	in.listening = true
	in.mu.Unlock()

	// Emit "listen" from a goroutine so a handler registered right after
	// CreateInbound still catches it.
	go in.emit("listen")

	go in.acceptLoop(listener)
	return nil
}

// bind opens the net or TLS listener for the resolved network/address.
func (in *Inbound) bind(network, address string) (net.Listener, error) {
	if in.main.opt.TLS == nil {
		return net.Listen(network, address)
	}
	cfg, err := buildServerTLSConfig(in.main.opt.TLS)
	if err != nil {
		return nil, err
	}
	return tls.Listen(network, address, cfg)
}

// buildServerTLSConfig maps the server TLSConfig to a *tls.Config (ca, cert,
// key, requestCert).
func buildServerTLSConfig(cfg *srvutils.TLSConfig) (*tls.Config, error) {
	out := &tls.Config{}
	if len(cfg.Cert) > 0 && len(cfg.Key) > 0 {
		pair, err := tls.X509KeyPair(cfg.Cert, cfg.Key)
		if err != nil {
			return nil, err
		}
		out.Certificates = append(out.Certificates, pair)
	}
	if len(cfg.CA) > 0 {
		pool := x509.NewCertPool()
		if pool.AppendCertsFromPEM(cfg.CA) {
			out.ClientCAs = pool
		}
	}
	if cfg.RequestCert {
		out.ClientAuth = tls.RequestClientCert
	}
	return out, nil
}

// acceptLoop accepts connections until the listener closes, spawning a
// per-socket reader goroutine.
func (in *Inbound) acceptLoop(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		in.onClientConnected(conn)
	}
}

// onClientConnected registers a socket, sets no-delay, and starts its reader
// with a per-socket codec (so concurrent connections do not interleave their
// buffers, issue #132).
func (in *Inbound) onClientConnected(conn net.Conn) {
	in.mu.Lock()
	in.sockets = append(in.sockets, conn)
	in.mu.Unlock()

	if tcp, ok := conn.(*net.TCPConn); ok {
		_ = tcp.SetNoDelay(true)
	}

	// The charset was validated when the listener was created (issue #50), so
	// the codec cannot fail here; an empty charset keeps UTF-8.
	codec, err := modules.NewMLLPCodecWithCharset("", in.opt.Charset)
	if err != nil {
		in.closeSocket(conn)
		in.emit("error", err)
		return
	}

	in.emit("client.connect", conn)

	go in.readLoop(conn, codec)
}

// readLoop reads framed bytes for one socket, feeds the per-socket codec,
// dispatches completed messages, and handles socket errors/close.
func (in *Inbound) readLoop(conn net.Conn, codec *modules.MLLPCodec) {
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			in.onData(conn, codec, buf[:n])
		}
		if err != nil {
			in.closeSocket(conn)
			in.emit("client.close", err != nil)
			return
		}
	}
}

// onData processes one chunk: feed the codec, and on a complete frame parse it
// as file/batch/message and dispatch.
func (in *Inbound) onData(conn net.Conn, codec *modules.MLLPCodec, chunk []byte) {
	var dataResult bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				in.emit("data.error", asError(r))
			}
		}()
		dataResult = codec.ReceiveData(chunk)
	}()
	if !dataResult {
		return
	}

	in.mu.Lock()
	in.stats.received++
	in.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			in.emit("data.error", asError(r))
		}
	}()

	loaded := codec.GetLastMessage()
	if loaded == nil {
		return
	}
	completed := *loaded

	in.emit("data.raw", completed)

	switch {
	case utils.IsFile(completed):
		messages, err := parseMessages(func() ([]*builder.Message, error) {
			parser, err := builder.NewFileBatch(builder.FileOptions{Text: completed})
			if err != nil {
				return nil, err
			}
			return parser.Messages(), nil
		})
		if err != nil {
			in.rejectUnparseable(conn, completed)
			in.emit("data.error", err)
			return
		}
		in.handleMessages(conn, messages, "file")
	case utils.IsBatch(completed):
		messages, err := parseMessages(func() ([]*builder.Message, error) {
			parser, err := builder.NewBatch(builder.BatchOptions{Text: completed})
			if err != nil {
				return nil, err
			}
			return parser.Messages(), nil
		})
		if err != nil {
			in.rejectUnparseable(conn, completed)
			in.emit("data.error", err)
			return
		}
		in.handleMessages(conn, messages, "batch")
	default:
		in.dispatch(conn, completed, "message")
	}
}

// parseMessages runs a batch or file parse and turns a panic inside the parser
// (a header too short to read, for example) into an error, so the caller can
// reject the frame instead of dropping it silently (issue #49).
func parseMessages(parse func() ([]*builder.Message, error)) (messages []*builder.Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			messages, err = nil, asError(r)
		}
	}()
	return parse()
}

// handleMessages re-parses and dispatches each message from a batch/file body,
// rebuilding each Message from its text.
func (in *Inbound) handleMessages(conn net.Conn, messages []*builder.Message, fromType string) {
	for _, msg := range messages {
		in.dispatch(conn, msg.String(), fromType)
	}
}

// dispatch parses one message body, bumps totalMessage, builds the request and
// response, wires response.sent, and invokes the handler.
func (in *Inbound) dispatch(conn net.Conn, msgText, fromType string) {
	parsed, err := builder.NewMessage(builder.MessageOptions{Text: msgText})
	if err != nil {
		in.rejectUnparseable(conn, msgText)
		in.emit("data.error", err)
		return
	}

	in.mu.Lock()
	in.stats.totalMessage++
	in.mu.Unlock()

	// Enforce this listener's accepted HL7 version(s): an inbound message whose
	// MSH.12 is not accepted is rejected with an AR (Application Reject) ACK, a
	// version-mismatch event is emitted, and the handler is NOT invoked. The
	// accepted set is a single version, an allow-list, or any known version
	// (AcceptAnyVersion); the decision goes through opt.Accepts, the one shared
	// acceptance check.
	if got := parsed.Get("MSH.12").String(); !in.opt.Accepts(got) {
		res := in.newResponse(conn, parsed)
		_ = res.SendResponse("AR")
		in.emit("data.error", srvutils.NewHL7ServerError(
			"message version \""+got+"\" is not accepted by the listener (accepts "+in.opt.AcceptedDescription()+").",
		))
		return
	}

	request := NewInboundRequest(parsed, InboundRequestProps{Socket: conn, Type: fromType})
	res := in.newResponse(conn, parsed)
	res.On("response.sent", func(_ ...any) { in.emit("response.sent") })

	if err := in.handler(request, res); err != nil {
		in.emit("data.error", err)
	}
}

// rejectUnparseable replies to text that failed to parse so the sender does not
// wait out its timeout and retransmit it (issue #49). When the first segment is
// a readable MSH, the reject ACK is built from it like any other ACK (AR in
// original mode, CR in enhanced mode, echoing MSH-10). Otherwise, or when that
// ACK cannot be built, a bare AR ACK is sent that still echoes the MSH-10 of the
// first line in MSA-2 and uses the listener's version when it has exactly one.
func (in *Inbound) rejectUnparseable(conn net.Conn, msgText string) {
	first, _, _ := strings.Cut(strings.ReplaceAll(msgText, "\n", "\r"), "\r")
	if strings.HasPrefix(first, "MSH") {
		if msh, err := builder.NewMessage(builder.MessageOptions{Text: first}); err == nil {
			if _, err := in.newResponse(conn, msh).SendAck(OutcomeReject); err == nil {
				return
			}
		}
	}
	version := "2.7"
	if len(in.opt.Versions) == 1 {
		version = in.opt.Versions[0]
	}
	_ = in.newResponse(conn, nil).SendCustomResponse(bareRejectAck(first, version, time.Now()))
}

// bareRejectAck builds the AR ACK sent when no usable MSH is available. MSA-2
// echoes MSH-10 of the first line when it can be read from the raw text (MSH-1
// is the field separator, so MSH-n is the (n-1)th field after splitting).
func bareRejectAck(first, version string, now time.Time) string {
	controlID := ""
	if len(first) > 3 && strings.HasPrefix(first, "MSH") {
		if fields := strings.Split(first, string(first[3])); len(fields) > 9 {
			controlID = fields[9]
		}
	}
	return "MSH|^~\\&|||||" + utils.CreateHL7Date(now, "") + "||ACK|" + utils.RandomString(20) + "|P|" + version + "\rMSA|AR|" + controlID
}

// newResponse builds the SendResponse for a request.
func (in *Inbound) newResponse(conn net.Conn, message *builder.Message) *SendResponse {
	return NewSendResponse(conn, message, in.opt.MSHOverrides)
}

// closeSocket destroys a socket and removes it from the tracked list.
func (in *Inbound) closeSocket(conn net.Conn) {
	_ = conn.Close()
	in.mu.Lock()
	for i, s := range in.sockets {
		if s == conn {
			in.sockets = append(in.sockets[:i], in.sockets[i+1:]...)
			break
		}
	}
	in.mu.Unlock()
}

// asError coerces a recovered panic value into an error for data.error.
func asError(r any) error {
	if err, ok := r.(error); ok {
		return err
	}
	return srvutils.NewHL7ServerError("data error")
}
