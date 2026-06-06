# go-hl7 Client :: Client

## Introduction

To send an HL7 message, your client must connect to a server or broker that accepts messages on a specific port. Traditionally, this communication has used the TCP/MLLP protocol. UDP is not generally used because clients typically need confirmation that their message was received and accepted by the server.

While HTTPS/HTTP (via REST APIs) has become more common in modern architectures, this library currently supports only the traditional TCP/MLLP transport. MLLP does not natively provide encryption, but since HL7 messages are often exchanged within trusted networks or over secure IPSEC tunnels, additional message-level security has not always been prioritized.

> **Note:** Future versions of this library may include hooks that allow message encryption prior to transmission. The recipient would be responsible for decrypting the message before processing.

In many HL7 environments, multiple ports on the same server handle different types of messages (e.g., ADT vs. ORU), each associated with a specific workflow. While most systems use dedicated ports for specific message types, some can handle multiple types on the same port and parse accordingly.

Let's walk through how to get started using this library's `Client`.

## Table of Contents

1. [Introduction](#introduction)
2. [Basic Usage](#basic-usage)
3. [TLS](#-tls)
4. [Mutual TLS (mTLS)](#-mutual-tls-mtls)
5. [Saved Messages](#saved-messages)
6. [Running in Kubernetes](#running-in-kubernetes)

## Basic Usage

This library supports connections over IPv4, IPv6, or Fully Qualified Domain Names (FQDNs). **It runs IPv4-only by default.** Opt into dual-stack by setting both `IPv4: ptr(true)` and `IPv6: ptr(true)` — the dialer then tries both attempts so a stale or unreachable record on one family transparently falls back to the other. Passing only one is treated as exclusive (that family only).

> **Note:** IPv4 and IPv6 formats are validated for correctness when an IP literal is passed. FQDNs are not checked against DNS for resolution.

> **Note:** The IP addresses shown in this documentation follow [RFC5737](https://datatracker.ietf.org/doc/html/rfc5737) and [RFC3849](https://datatracker.ietf.org/doc/html/rfc3849) for documentation use. Replace them with actual internal or external IPs in production.

> 💡 The option structs use pointer fields (e.g. `Port *int`, `IPv4 *bool`), so define a small helper once and reuse it: `func ptr[T any](v T) *T { return &v }`.

### Step 1: Create the Client

```go
import "github.com/Bugs5382/go-hl7/client/client"

// IPv4 only (default)
c, _ := client.NewClient(client.ClientOptions{Host: "hl7.example.com"})

// Dual-stack with auto-fallback (opt-in)
dual, _ := client.NewClient(client.ClientOptions{
    Host: "hl7.example.com", IPv4: ptr(true), IPv6: ptr(true),
})

// Force IPv6 only (and validate the host against it):
v6Only, _ := client.NewClient(client.ClientOptions{Host: "2001:db8::1", IPv6: ptr(true)})
```

This initializes a client targeting the host, but does not yet establish a connection.

#### Address-family options

| Option | Meaning |
|---|---|
| (defaults) | IPv4 only — host literal must be IPv4 |
| `IPv4: ptr(true), IPv6: ptr(true)` | dual-stack with auto-fallback |
| `IPv6: ptr(true)` only | force IPv6 — host literal must be IPv6 |

If a host has multiple IPv4 / IPv6 addresses and you need to terminate on a specific one, pass the literal as `Host` — the literal pins the family for you. Setting both `IPv4` and `IPv6` to `false` returns an error from `NewClient`.

### Step 2: Create an Outbound Connection

HL7 messages are sent to specific ports. You must initiate an outbound connection to begin sending.

```go
OB_ADT, _ := c.CreateConnection(
    client.ClientListenerOptions{Port: ptr(5678)},
    func(res *client.InboundResponse) error {
        status := res.GetMessage().Get("MSA.1").String() // MSA is the Message Acknowledgment Segment

        if status == "AA" {
            // Message was accepted successfully
        } else {
            // The message may have failed to process
        }
        return nil
    },
)
```

The callback is an `OutboundHandler` — `func(res *client.InboundResponse) error`. Returning an error surfaces it as a `data.error` event.

### Step 3: Send a Message

```go
_ = OB_ADT.SendMessage(message) // `message` is a MessageItem — *builder.Message, *builder.Batch, or *builder.FileBatch.
```

Outbound connections are persistent by design. This allows multiple messages to be sent over a single TCP/MLLP socket without repeatedly re-establishing the connection.

If the connection drops, the library will attempt to reconnect up to 10 times (or a user-defined limit via `MaxConnectionAttempts`) before giving up, using exponential backoff (`RetryLow` → `RetryHigh`). If reconnection fails, your application will need to restart the connection process.

### Step 4: Close the Connection

To permanently close a connection without attempting to reconnect:

```go
_ = OB_ADT.Close()
```

Useful methods on `*Connection`: `Connect()` (when `AutoConnect` is false), `Close()`, `IsConnected()`, and `GetPort()`. On `*Client`: `CreateConnection`, `CloseAll()`, `TotalSent()`, `TotalAck()`, `TotalPending()`, and `GetHost()`.

## 🔒 TLS

If the remote HL7 server expects TLS, set `TLS` on the `ClientOptions`. A non‑nil `*client.TLSConfig` enables it. Two forms are accepted:

**Shorthand** — use the system trust store (works for certs chained to public CAs):

```go
c, _ := client.NewClient(client.ClientOptions{Host: "hl7.example.com", TLS: &client.TLSConfig{}})
```

**Full options** — use this when the server uses a private/self‑signed CA or you need to tune `ServerName`, etc.:

```go
import (
    "os"

    "github.com/Bugs5382/go-hl7/client/client"
)

ca, _ := os.ReadFile("certs/server-ca-crt.pem")

c, _ := client.NewClient(client.ClientOptions{
    Host: "hl7.example.local",
    TLS: &client.TLSConfig{
        // 🪪 Trust this CA for the server cert.
        CA: ca,
        // RejectUnauthorized: leave nil to keep Go's default verification on.
        // Set to ptr(false) to skip verification (local dev only).
    },
})
```

> 🚨 **`RejectUnauthorized: ptr(false)`** disables cert validation entirely and is meant only for local development. Anything that talks to a real hospital network should leave it `nil` (the secure default) and provide a `CA` if needed.

## 🛡️ Mutual TLS (mTLS)

When the remote server demands a **client certificate** (the typical hospital integration pattern), provide your own `Key` and `Cert` alongside the trusted CA:

```go
key, _ := os.ReadFile("certs/client-key.pem")
crt, _ := os.ReadFile("certs/client-crt.pem")
ca, _ := os.ReadFile("certs/server-ca-crt.pem")

c, _ := client.NewClient(client.ClientOptions{
    Host: "hl7.example.local",
    TLS: &client.TLSConfig{
        // 🔑 The client's own identity (this is the cert the server validates).
        Key:  key,
        Cert: crt,

        // 🪪 CA(s) you trust to issue the server's certificate.
        CA: ca,

        // RejectUnauthorized stays nil (secure default) — Go drops the
        // connection if any cert in the chain fails to validate.

        // (Optional) SNI / expected server hostname; defaults to Host.
        // ServerName: "hl7.example.local",
    },
})

OB_ADT, _ := c.CreateConnection(client.ClientListenerOptions{Port: ptr(6661)},
    func(res *client.InboundResponse) error {
        fmt.Println("✅", res.GetMessage().Get("MSA.1").String())
        return nil
    })
```

| Field | Purpose |
|---|---|
| `Key` + `Cert` | Your client identity. The server checks these against its `ca` allow-list. |
| `CA` | Trusted issuer(s) for the **server**'s certificate. |
| `RejectUnauthorized` | `*bool`; `nil`/`true` keeps verification on, `ptr(false)` skips it. Keep it on in production. |
| `ServerName` | SNI / expected server hostname when it differs from `Host`. |

> 🤝 The matching server-side mTLS configuration is documented in the [server's TLS pages](../../server/tls/index.md). The two ends MUST agree on which CA(s) issue valid certs in each direction.

## Saved Messages

This library allows you to override the default in-memory message queue behavior by supplying your own message queuing logic. You provide two hooks on `ClientListenerOptions`:

- `EnqueueMessage` – called when a message is ready to be stored.
- `FlushQueue` – called to retrieve messages and deliver them back to the connection for processing.

Their signatures are:

```go
EnqueueMessage func(message client.MessageItem, notifyPendingCount client.NotifyPendingCount) error
FlushQueue     func(callback client.FallBackHandler, notifyPendingCount client.NotifyPendingCount) error
```

where `MessageItem` is anything with a `String() string` method (so `*builder.Message`, `*builder.Batch`, and `*builder.FileBatch` all qualify), `FallBackHandler` delivers a queued message back into the connection, and `NotifyPendingCount func(count int) error` reports the depth (and feeds the `client.pending` event).

### Default Behavior (In-Memory)

If you don't supply these hooks, messages are stored in an internal in-memory slice and flushed in order. The default store appends to the slice (dropping the oldest when the cap is hit unless `ExtendMaxLimit` is set), and the default drain pops one message at a time and delivers it back through the connection.

> ⚠️ **Note:** There is a max of 10,000 messages by default (`MaxLimit`). It can be overridden to fit the amount of your choice or extended past the cap with `ExtendMaxLimit: ptr(true)`. The connection can also, optionally, emit `client.limitExceeded` (set `NotifyOnLimitExceeded: ptr(true)`) to see if a particular connection is in this state.

### Custom Behavior (Using Redis)

You can override the default queue to use Redis or any other external storage like RabbitMQ, file-based queues, etc. **This is strongly recommended.** Pass `AutoConnect: ptr(false)` when wiring a durable store so the connection flushes the queue once you `Connect()`.

**Redis Example** (using a Redis client of your choice — the queue hooks are storage-agnostic):

```go
import (
    "context"

    "github.com/Bugs5382/go-hl7/client/builder"
    "github.com/Bugs5382/go-hl7/client/client"
)

ctx := context.Background()
// rdb is your *redis.Client (or any durable store wrapper).

enqueueMessage := func(m client.MessageItem, notify client.NotifyPendingCount) error {
    if err := rdb.LPush(ctx, "hl7queue", m.String()).Err(); err != nil {
        return err
    }
    depth, _ := rdb.LLen(ctx, "hl7queue").Result()
    return notify(int(depth))
}

flushQueue := func(deliver client.FallBackHandler, notify client.NotifyPendingCount) error {
    for {
        depth, _ := rdb.LLen(ctx, "hl7queue").Result()
        if depth == 0 {
            return nil
        }
        raw, err := rdb.RPop(ctx, "hl7queue").Result()
        if err != nil {
            return err
        }
        msg, err := builder.NewMessage(builder.MessageOptions{Text: raw})
        if err != nil {
            return err
        }
        deliver(msg)
        if err := notify(int(depth - 1)); err != nil {
            return err
        }
    }
}

c, _ := client.NewClient(client.ClientOptions{Host: "0.0.0.0"})

// Create connection without auto-connecting.
outbound, _ := c.CreateConnection(
    client.ClientListenerOptions{
        Port:           ptr(5678),
        AutoConnect:    ptr(false),
        EnqueueMessage: enqueueMessage,
        FlushQueue:     flushQueue,
    },
    func(res *client.InboundResponse) error { return nil }, // simplified here
)
```

**Important:**

- `EnqueueMessage` returns an `error` — surface store failures rather than swallowing them.
- `FlushQueue` should call `deliver(msg)` rather than processing the message directly; `deliver` is the `FallBackHandler` that re-injects the message into the connection.
- The message passed to `EnqueueMessage` is always a `MessageItem` — one of `*builder.Message`, `*builder.Batch`, or `*builder.FileBatch`. A custom store can persist `m.String()` and rebuild a `Message` with `builder.NewMessage` on flush.

This flexible queuing system allows seamless integration with external storage systems—such as Redis, RabbitMQ, databases, or even flat files—enabling you to offload in-memory storage and better manage system resources.

> 🔐 **Data Safety Warning:**
> If using shared queues (like Redis), **tag or isolate messages per client instance** to prevent sending messages to the wrong downstream service. Mismatching or leaking data between client/ports can result in serious issues in production systems.

### ⚙️ Scalability & Message Reliability in Kubernetes

This library has been designed to run across **multiple pod instances** in a **Kubernetes** environment. Because this is an **outbound client connection**, the upstream HL7 server or broker returns its response to the **same client instance** that initiated the request.

> ⚠️ **Important Note on Reliability:**
> If a pod sends a message and then crashes or is terminated **before receiving the response**, that response may be **lost permanently** unless handled by an external failover or retry strategy.

### 💾 Offloading Messages with Custom Queues

Inside a Kubernetes setup you should use custom logic to store outbound messages (via `EnqueueMessage`), and you must avoid using the built-in in-memory storage within the pod. Always offload the queue to a **persistent, external system** such as:

- Redis (preferred)
- RabbitMQ
- SQL/NoSQL Databases
- Flat files or S3 buckets (need persistent storage across reboots)

This ensures your message queue is resilient to pod restarts, crashes, and horizontal scaling.

> 🔐 **Data Safety Warning:**
> If using shared queues (like Redis), **tag or isolate messages per client instance** to prevent sending messages to the wrong downstream service. Mismatching or leaking data between client/ports can result in serious issues in production systems.
