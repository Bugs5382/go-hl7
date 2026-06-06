# AGENTS.md — go-hl7 for AI coding agents

> Dense, example-first reference for using `go-hl7` **as a dependency**. Every code block compiles against the shipping API (Go ≥ 1.26). If you are an agent generating code that imports this library, read this file first and copy the patterns verbatim — the non-obvious rules in [Hard Rules](#hard-rules-read-before-writing-code) are where naive code breaks.

Module path: `github.com/Bugs5382/go-hl7` (single module, no external runtime deps).

```sh
go get github.com/Bugs5382/go-hl7
```

A helper used throughout (option structs use pointer fields to tell "unset" from "zero"):

```go
func ptr[T any](v T) *T { return &v }
```

---

## Package map

| Import path | What you use it for | Key exports |
|---|---|---|
| `github.com/Bugs5382/go-hl7/client/hl7` | Typed, version-aware **message builders**. | `NewHL7_2_1` … `NewHL7_2_8`, `Props` (`= map[string]any`), `Options`, `BuildMSH`, `Build<SEG>` (`BuildPID`, `BuildEVN`, `BuildOBX`, `BuildORC`, `BuildOBR`, `BuildPV1`, `BuildNTE`, `BuildMSA`, `BuildERR`, …), `BuildSegment(name, props)`, `ToMessage() *builder.Message`, `String()`. |
| `github.com/Bugs5382/go-hl7/client/builder` | The **wire message model** + parser. | `Message`, `Batch`, `FileBatch`; `NewMessage`, `NewBatch`, `NewFileBatch`; `MessageOptions`, `BatchOptions`, `FileOptions`; node API `Get(path)`, `Index(i)`, `Set(path, v)`, `SetIndex(i, v)`, `Exists(path)`, coercions `String/Int/Float/Bool/Date`. |
| `github.com/Bugs5382/go-hl7/client/client` | The **outbound** TCP/MLLP client. | `NewClient`, `Client`, `Connection`, `ClientOptions`, `ClientListenerOptions`, `TLSConfig`, `InboundResponse`, `MessageItem`, `OutboundHandler`, `FallBackHandler`, `NotifyPendingCount`. |
| `github.com/Bugs5382/go-hl7/client/modules` | The MLLP codec (rarely touched directly). | `MLLPCodec`, `NewMLLPCodec(returnChar string)`, `(*MLLPCodec).SendMessage(w io.Writer, msg string)`. |
| `github.com/Bugs5382/go-hl7/server` | The **inbound** TCP/TLS listener. | `NewServer`, `Server`, `ServerOptions`, `ListenerOptions`, `TLSConfig`, `CreateInbound`, `Inbound`, `InboundRequest`, `ResponseSender`, `InboundHandler`, `MSHOverride`, `StringOverride`, `FuncOverride`. |
| `github.com/Bugs5382/go-hl7/client/helpers` | Error sentinels for `errors.Is`. | `ErrFatal` (`HL7FatalError`, 500), `ErrParser` (`HL7ParserError`, 404), `ErrValidation` (`HL7ValidationError`, 404). |
| `github.com/Bugs5382/go-hl7/client/hl7/metadata` | Generated spec metadata. | `SEGMENT_SPECS map[string]SegmentSpec`, `DATATYPE_SPECS`, `SegmentSpec`, `FieldSpec`, `IsKnownVersion(v string) bool`. |
| `github.com/Bugs5382/go-hl7/client/hl7/tables` | Generated HL7 value tables. | `TABLES map[string]map[string][]string` (version → tableID → allowed codes). |
| `github.com/Bugs5382/go-hl7/client/utils` | Small helpers. | `CreateHL7Date(t time.Time, length string) string` (length `""`/`"8"`/`"12"`/`"14"`). |

Version constructors (no implicit default — the constructor *is* the version selector):
`NewHL7_2_1`, `NewHL7_2_2`, `NewHL7_2_3`, `NewHL7_2_3_1`, `NewHL7_2_4`, `NewHL7_2_5`, `NewHL7_2_5_1`, `NewHL7_2_6`, `NewHL7_2_7`, `NewHL7_2_7_1`, `NewHL7_2_8`. Each takes an optional `hl7.Options` (`hl7.NewHL7_2_5()` is valid).

---

## Hard Rules (read before writing code)

1. **HL7 version is REQUIRED per connection — and mismatches are rejected.**
   - **Client:** `ClientOptions.Version` is mandatory and single-set. It must be one of `2.1, 2.2, 2.3, 2.3.1, 2.4, 2.5, 2.5.1, 2.6, 2.7, 2.7.1, 2.8`, else `NewClient` returns `version is not defined.` / `version is not a valid HL7 version.`. Before any send, the message's `MSH.12` **must equal** the client version; if it differs, `conn.SendMessage` returns an error and **does not transmit** (for a batch/file, *every* inner message must match). Error text: `message version "2.5" does not match the connection version "2.7".`
   - **Server:** `ListenerOptions.Version` is mandatory **per listener** (same valid set / errors from `CreateInbound`). When an inbound `MSH.12` differs, the server replies with an **`AR`** (Application Reject) ACK, emits a version-mismatch `data.error` event, and **does not call your handler**. Dedicate one port per version.
   - This is intentional and diverges from node-hl7 (which is version-agnostic on the transport). Build the message with the **matching version constructor** for the connection.

2. **`BuildMSH` must run first.** Calling any other `Build*` (or `BuildSegment`) before `BuildMSH` **panics** with `HL7FatalError("MSH Header must be built first.")`. Calling `BuildMSH` twice **panics** with `HL7FatalError("You can only have one MSH Header per HL7 Message.")`. MSH must go through `BuildMSH`, not `BuildSegment("MSH", …)`.

3. **Value-table enforcement is a hard, version-aware error.** Table-bound fields/components (e.g. `MSA.1`→`0008`, `PID.8`→`0001`, `OBX.11`→`0085`) are validated against the value set for the active version. An out-of-table code raises `HL7ValidationError` — and a code valid in one version can be rejected in another. Per-version usage codes also apply: withdrawn (`W`/`X`) fields error, backward-compat (`B`) fields warn, and segments that didn't exist in the version are refused.

4. **Panics vs errors — know which is which.**
   - **Return `error`:** all constructors — `NewClient`, `CreateConnection`, `NewServer`, `CreateInbound`, `builder.NewMessage`/`NewBatch`/`NewFileBatch` — and senders `conn.SendMessage`, `res.SendResponse`, `res.SendCustomResponse`.
   - **Panic with a typed `HL7Error`:** the spec-driven builders (`BuildMSH`/`Build<SEG>`/`BuildSegment`) on hard validation/usage failures, and some structural reads. Set `hl7.Options{HardError: true}` to make *soft* findings panic too; otherwise they surface via `b.On("error"/"warning", func(string))`. If you need errors at a boundary, wrap the builder run in `recover`.
   - Match returned errors with `errors.Is(err, helpers.ErrParser | helpers.ErrValidation | helpers.ErrFatal)`.

5. **Reads never panic on a missing path.** `Get(path)` of a missing node returns a shared **empty node**, so `Get(...).String()` always yields `""`. Use `Exists(path)` to distinguish. Coercions return `(T, ok)` — never a bare value.

6. **No method overloading.** Dotted HL7 path → `Get(path string)` / `Set(path, v)` (1-based, e.g. `"PID.5.1"`). 0-based child position → `Index(i int)` / `SetIndex(i, v)`. Don't mix them up.

---

## Build a message (MSH first → segments → ToMessage)

```go
package main

import (
	"time"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/hl7"
)

func buildADT() *builder.Message {
	b := hl7.NewHL7_2_7(). // version pinned by the constructor
		BuildMSH(hl7.Props{ // MUST be first
			"msh_3":  "MY_APP",
			"msh_4":  "MY_FAC",
			"msh_5":  "EPIC",
			"msh_6":  "HOSP",
			"msh_9":  "ADT^A01",  // composite OK as a literal string
			"msh_10": "MSG00001", // control id; auto-randomized if omitted
			"msh_11": "P",        // P = production, T = test
		}).
		BuildEVN(hl7.Props{"evn_1": "A01", "evn_2": time.Now()}).
		BuildPID(hl7.Props{
			"pid_3":  "MRN12345",
			"pid_5":  "DOE^JANE^A",                       // last^first^middle
			"pid_7":  time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC), // time.Time or string
			"pid_8":  "F",                                 // validated against table 0001
			"pid_11": "123 ELM ST^^SPRINGFIELD^IL^62701",  // ^^ = empty component
		}).
		BuildSegment("DG1", hl7.Props{ // universal builder for the long tail
			"dg1_1": "1",
			"dg1_3": "I10^Diagnosis^I10",
		})

	msg := b.ToMessage() // *builder.Message — keep mutating or send it
	msg.Set("PID.13", "555-0100")       // dotted path, 1-based
	seg, _ := msg.AddSegment("NTE")     // returns (*Segment, error)
	_ = seg
	return msg
}
```

`Props` values may be `string`, `int`, `time.Time`, or a `map[string]any` composite. Delimiters inside a string literal: `^` component, `&` sub-component, `~` repetition, `^^` empty component. A composite may also be passed as a typed object — components keyed by number (`"1"`), trailing `_<num>`, or camelCased spec label:

```go
b.BuildPID(hl7.Props{
	"pid_11": map[string]any{
		"streetAddress":   "123 ELM ST",
		"city":            "SPRINGFIELD",
		"stateOrProvince": "IL",
		"zipOrPostalCode": "62701",
	},
})
```

Non-standard encoding chars (embedded in `MSH.1`/`MSH.2`, immutable after construction):

```go
b := hl7.NewHL7_2_5(hl7.Options{
	SeparatorField: "!", SeparatorComponent: "+", SeparatorSubComponent: "]",
	SeparatorRepetition: "?", SeparatorEscape: "#",
})
```

---

## Parse a message / reply

```go
import "github.com/Bugs5382/go-hl7/client/builder"

msg, err := builder.NewMessage(builder.MessageOptions{Text: hl7String})
if err != nil { /* malformed HL7 (errors.Is(err, helpers.ErrParser)) */ }

_ = msg.Get("MSH.9.1").String()                 // ADT
_ = msg.Get("PID.5.1").String()                 // DOE
_ = msg.Get("PID.3").Index(0).Index(3).String() // repetition 0, component 3
_ = msg.Exists("PV1.44")                         // false ⇒ resolves to empty node

n, ok := msg.Get("OBX.5").Int()    // (int, bool)
f, ok := msg.Get("OBX.5").Float()  // (float64, bool)
yn, ok := msg.Get("PID.30").Bool() // "Y"/"N" → (bool, bool)
ts, ok := msg.Get("PID.7").Date()  // HL7 date → (time.Time, bool)
_, _, _, _, _, _, _, _ = n, f, yn, ts, ok, ok, ok, ok
```

`NewMessage` rejects a body not starting with `MSH` or carrying multiple MSH segments (use `NewBatch` for that). `NewBatch` rejects a single-MSH body. All three constructors return `(T, error)`.

---

## Send as a client (Version is required)

```go
package main

import (
	"fmt"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/client"
)

func ptr[T any](v T) *T { return &v }

func send(msg *builder.Message) {
	c, _ := client.NewClient(client.ClientOptions{Version: "2.7", Host: "127.0.0.1"}) // Version REQUIRED
	conn, _ := c.CreateConnection(
		client.ClientListenerOptions{
			Port:                  ptr(3000), // remote port, required
			WaitAck:               ptr(true), // serialize sends on the previous ACK (default true)
			MaxConnectionAttempts: ptr(10),
		},
		func(res *client.InboundResponse) error { // OutboundHandler: receives the ACK
			fmt.Println("ACK:", res.GetMessage().Get("MSA.1").String()) // AA / AR / AE
			return nil
		},
	)
	defer conn.Close()

	if err := conn.SendMessage(msg); err != nil { // msg.MSH.12 must == "2.7" or this errors and does NOT send
		fmt.Println("rejected:", err)
	}
}
```

`SendMessage` accepts any `MessageItem` (`*builder.Message`, `*builder.Batch`, `*builder.FileBatch`). Useful methods: `*Connection`: `Connect()`, `Close()`, `IsConnected()`, `GetPort()`. `*Client`: `CreateConnection`, `CloseAll()`, `TotalSent()`, `TotalAck()`, `TotalPending()`, `GetHost()`. The connection is persistent (many sends over one socket); on drop it reconnects with exponential backoff (`RetryLow`→`RetryHigh`, capped by `MaxConnectionAttempts`).

**Dual-stack:** IPv4-only by default. Opt into dual-stack with both `IPv4: ptr(true), IPv6: ptr(true)`; one alone is exclusive; both `false` errors from `NewClient`.

**TLS / mTLS:** a non-nil `*client.TLSConfig` enables TLS (`&client.TLSConfig{}` trusts system roots):

```go
c, _ := client.NewClient(client.ClientOptions{
	Version: "2.7", Host: "hl7.example.local",
	TLS: &client.TLSConfig{
		Cert: crt, Key: key, // client identity (mTLS)
		CA:   ca,            // trusted issuer(s) for the server cert
		ServerName: "hl7.example.local",
		// RejectUnauthorized: nil keeps Go verification ON (keep it on in prod).
	},
})
```

**Pluggable queue** (offload to Redis/etc. for multi-pod; default is in-memory capped at `MaxLimit`=10000):

```go
c.CreateConnection(
	client.ClientListenerOptions{
		Port: ptr(3000), AutoConnect: ptr(false),
		EnqueueMessage: func(m client.MessageItem, notify client.NotifyPendingCount) error {
			// e.g. redis.LPush(ctx, "hl7queue", m.String())
			return notify(queueDepth())
		},
		FlushQueue: func(deliver client.FallBackHandler, notify client.NotifyPendingCount) error {
			for queueDepth() > 0 {
				msg, err := builder.NewMessage(builder.MessageOptions{Text: pop()})
				if err != nil { return err }
				deliver(msg)
				if err := notify(queueDepth()); err != nil { return err }
			}
			return nil
		},
	},
	func(res *client.InboundResponse) error { return nil },
)
```

`MessageItem` is `interface { String() string }`; persist `m.String()` and rebuild on flush.

---

## Serve as a server (per-listener Version, auto + custom ACK)

```go
package main

import (
	"fmt"

	"github.com/Bugs5382/go-hl7/server"
)

func ptr[T any](v T) *T { return &v }

func main() {
	srv, _ := server.NewServer(nil) // nil ⇒ defaults: IPv4-only on 0.0.0.0

	in, _ := srv.CreateInbound(
		server.ListenerOptions{Version: "2.7", Port: ptr(3000)}, // Version REQUIRED per listener
		func(req *server.InboundRequest, res server.ResponseSender) error { // InboundHandler
			msg := req.GetMessage()  // *builder.Message
			_ = req.GetType()        // "message" | "batch" | "file"
			_ = req.GetSocket()      // net.Conn
			fmt.Println("recv", msg.Get("MSH.10").String())
			return res.SendResponse("AA") // auto ACK: AA/AR/AE (+ CA/CR/CE on ≥2.2)
		},
	)
	defer in.Close()

	in.On("listen", func(_ ...any) { fmt.Println("listening :3000") })
	select {} // keep alive
}
```

`InboundHandler` = `func(req *InboundRequest, res ResponseSender) error`. The handler is invoked **once per parsed message**, even inside a BHS batch or FHS file. Counters: `in.TotalReceived()` (frames), `in.TotalMessage()` (messages), `in.IsListening()`.

`ResponseSender` interface:

```go
type ResponseSender interface {
	GetAckMessage() *builder.Message // the ACK that was sent (nil before send)
	GetCodec() *modules.MLLPCodec
	GetSocket() net.Conn
	SendResponse(ackType string) error       // auto ACK (sender/receiver swapped, MSH.10→MSA.2)
	SendCustomResponse(message any) error     // verbatim *builder.Message OR raw HL7 string
}
```

**Version gate on auto ACKs:** `CA`/`CR`/`CE` require inbound `MSH.12` ≥ 2.2; on 2.1 the library falls back to `AE`. `AA`/`AR`/`AE` work on every version.

**Custom ACK** — full control of the wire bytes (no MSA validation, no MSH overrides, no auto-swap):

```go
import (
	"strings"
	"time"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/utils"
	"github.com/Bugs5382/go-hl7/server"
)

func handler(req *server.InboundRequest, res server.ResponseSender) error {
	ctrlID := req.GetMessage().Get("MSH.10").String()
	text := strings.Join([]string{
		"MSH|^~\\&|MY_APP|MY_FAC|EPIC|HOSP|" + utils.CreateHL7Date(time.Now(), "") + "||ACK^A01|RESP_" + ctrlID + "|P|2.7",
		"MSA|AA|" + ctrlID + "|All good|||MY_VENDOR_OK",
		"ERR|||0^Message accepted^HL70357|I",
	}, "\r")
	ack, err := builder.NewMessage(builder.MessageOptions{Text: text})
	if err != nil { return err }
	return res.SendCustomResponse(ack) // or: res.SendCustomResponse(text)
}
```

**MSH overrides** on the auto ACK (literal `StringOverride` or callback `FuncOverride`; applied only to `SendResponse`, skipped by `SendCustomResponse`):

```go
srv.CreateInbound(
	server.ListenerOptions{
		Version: "2.7", Port: ptr(3000),
		MSHOverrides: map[string]server.MSHOverride{
			"3":   server.StringOverride("MY_APP"),
			"9.3": server.StringOverride("ACK"),
			"12":  server.FuncOverride(func(m *builder.Message) string { return m.Get("MSH.12").String() }),
		},
	},
	func(req *server.InboundRequest, res server.ResponseSender) error { return res.SendResponse("AA") },
)
```

**TLS / mTLS (server side):**

```go
srv, _ := server.NewServer(&server.ServerOptions{
	TLS: &server.TLSConfig{
		Key: key, Cert: crt, // server identity
		CA: ca, RequestCert: true, // mTLS: require + verify client certs
	},
})
```

**Dual-stack:** IPv4-only by default (`0.0.0.0`). `IPv4: ptr(true), IPv6: ptr(true)` ⇒ binds `::`; IPv6-only ⇒ defaults to `::`; both `false` errors.

---

## Batches & file batches

A batch = BHS + N messages + BTS. A file batch = FHS + content + FTS. Both satisfy `MessageItem`, so either goes straight to `conn.SendMessage`. The receiver fans out to your handler once per inner message.

```go
import "github.com/Bugs5382/go-hl7/client/builder"

// Batch
batch, _ := builder.NewBatch(builder.BatchOptions{})
batch.Start("")                 // (re)stamp BHS.7; "" = default 14-char date
batch.Add(mk("MSG00001"), -1)   // index -1 appends
batch.Add(mk("MSG00002"), -1)
batch.End()                     // append BTS with the count
_ = conn.SendMessage(batch)

// File batch
file, _ := builder.NewFileBatch(builder.FileOptions{Location: "./out", Extension: "hl7"})
file.Start()
file.AddMessage(mk("MSG00001"))
file.End()
_ = file.CreateFile("ADT")      // writes hl7.ADT.<date>.hl7 under ./out
_ = file.FileName()

// Parse a batch
parsed, _ := builder.NewBatch(builder.BatchOptions{Text: hl7BatchString})
for _, m := range parsed.Messages() {
	_ = m.Get("MSH.10").String()
}
```

`Message.ToFile(...)` / `Batch.ToFile(...)` are one-shot convenience wrappers.

---

## Errors

```go
import (
	"errors"

	"github.com/Bugs5382/go-hl7/client/builder"
	"github.com/Bugs5382/go-hl7/client/helpers"
)

_, err := builder.NewMessage(builder.MessageOptions{Text: "not hl7"})
switch {
case errors.Is(err, helpers.ErrParser):     // HL7ParserError (404) — malformed body
case errors.Is(err, helpers.ErrValidation): // HL7ValidationError (404) — out-of-table / usage
case errors.Is(err, helpers.ErrFatal):      // HL7FatalError (500) — fatal usage/connection
}
```

Builders **panic** typed `HL7Error`s on hard validation; wrap in `recover` to convert at a boundary.

---

## Deeper docs

- `client/README.md` — full builder API, batches, queues, parsing, TLS, events.
- `server/README.md` — listener options, ACKs, MSH overrides, TLS/mTLS, performance, events.
- `pages/` — deep-dive walkthroughs (`pages/client/...`, `pages/server/...`).
- API reference: <https://pkg.go.dev/github.com/Bugs5382/go-hl7>.
