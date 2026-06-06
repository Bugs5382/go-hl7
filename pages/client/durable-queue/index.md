# 🗄️ go-hl7 Client :: Durable Queue

## Introduction

The client's outbound queue is in-memory by default: messages waiting to be sent (while reconnecting, or while a previous ACK is still outstanding) live in a slice on the `*Connection`. That is fast and dependency-free, but it does **not** survive a process restart. If the pod crashes or is rolled, every message still sitting in that slice is lost.

For any deployment where losing a queued message is unacceptable, back the queue with a database. The client exposes two hooks on `client.ClientListenerOptions` for exactly this — `EnqueueMessage` and `FlushQueue` — and they are storage-agnostic. This page walks through a **durable, database-backed** implementation using PostgreSQL as the worked example. MongoDB and MySQL use the same two hooks with the storage calls swapped; see the [closing note](#-mongodb-and-mysql).

> 💡 This builds directly on the [Saved Messages](../client/index.md#saved-messages) section of the Client page. If you have not read how the hooks fit together, start there — this page extends that same example to a persistent store.

## 🧾 Table of Contents

1. [Introduction](#introduction)
2. [The hooks](#-the-hooks)
3. [Schema](#-schema)
4. [EnqueueMessage — persist on the way in](#-enqueuemessage--persist-on-the-way-in)
5. [FlushQueue — redeliver and delete on success](#-flushqueue--redeliver-and-delete-on-success)
6. [Wiring it into a connection](#-wiring-it-into-a-connection)
7. [MongoDB and MySQL](#-mongodb-and-mysql)
8. [Operational notes](#-operational-notes)

---

## 🪝 The hooks

Both hooks live on `client.ClientListenerOptions`. Their exact signatures are:

```go
EnqueueMessage func(message client.MessageItem, notifyPendingCount client.NotifyPendingCount) error
FlushQueue     func(callback client.FallBackHandler, notifyPendingCount client.NotifyPendingCount) error
```

- **`MessageItem`** is anything with a `String() string` method, so `*builder.Message`, `*builder.Batch`, and `*builder.FileBatch` all qualify. Persist `message.String()` and rebuild a message with `builder.NewMessage` on the way out.
- **`FallBackHandler`** is `func(message client.MessageItem)` — it redelivers a queued message back into the connection. In `FlushQueue` you call it `callback(msg)` for each row you read; you do **not** send the message yourself.
- **`NotifyPendingCount`** is `func(count int) error` — report the current queue depth so the connection can surface it on the `client.pending` event.

If you leave both hooks `nil`, the client uses its default in-memory queue. Supply both to take over storage.

---

## 🧱 Schema

A single table is enough. Persist the serialized body and enough metadata to drain in order and to isolate per client instance.

```sql
CREATE TABLE pending_messages (
    id          BIGSERIAL    PRIMARY KEY,
    -- The serialized HL7 body — message.String(). One column holds a single
    -- Message, a Batch, or a FileBatch; they all serialize to text.
    payload     TEXT         NOT NULL,
    -- Which connection/client instance owns this row. Tag every row so a
    -- shared table never hands one pod's messages to another.
    instance_id TEXT         NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Drain oldest-first, scoped to this instance.
CREATE INDEX pending_messages_drain_idx
    ON pending_messages (instance_id, created_at, id);
```

> 🔐 **Tag every row with an `instance_id`.** If several pods or connections share one table, draining without a scope can send one instance's messages to the wrong downstream service. This is the database equivalent of the shared-Redis warning on the Client page.

---

## 📥 `EnqueueMessage` — persist on the way in

`EnqueueMessage` is called when a message is ready to be stored. Insert the serialized body, then report the new depth through `notifyPendingCount`. Return the store error rather than swallowing it — a failed enqueue should be visible.

PostgreSQL here uses the standard `database/sql` package with a driver of your choice (`github.com/lib/pq` or `github.com/jackc/pgx/v5/stdlib`). The hook interface does not care which — it only sees a `*sql.DB`.

```go
import (
    "context"
    "database/sql"

    "github.com/Bugs5382/go-hl7/client/client"
)

// db is your *sql.DB, opened with the Postgres driver of your choice.
// instanceID isolates this connection's rows in a shared table.
func enqueueMessage(ctx context.Context, db *sql.DB, instanceID string) func(client.MessageItem, client.NotifyPendingCount) error {
    return func(m client.MessageItem, notify client.NotifyPendingCount) error {
        // m.String() works for *builder.Message, *builder.Batch, and
        // *builder.FileBatch alike.
        if _, err := db.ExecContext(ctx,
            `INSERT INTO pending_messages (payload, instance_id) VALUES ($1, $2)`,
            m.String(), instanceID,
        ); err != nil {
            return err
        }

        var depth int
        if err := db.QueryRowContext(ctx,
            `SELECT count(*) FROM pending_messages WHERE instance_id = $1`,
            instanceID,
        ).Scan(&depth); err != nil {
            return err
        }
        return notify(depth)
    }
}
```

---

## 📤 `FlushQueue` — redeliver and delete on success

`FlushQueue` drains the store back into the connection. Read pending rows oldest-first, hand each body to the `callback` (the `FallBackHandler`) so it is re-injected into the connection, then **delete the row** once it has been handed back. After each row, report the remaining depth.

```go
import (
    "context"
    "database/sql"

    "github.com/Bugs5382/go-hl7/client/builder"
    "github.com/Bugs5382/go-hl7/client/client"
)

func flushQueue(ctx context.Context, db *sql.DB, instanceID string) func(client.FallBackHandler, client.NotifyPendingCount) error {
    return func(deliver client.FallBackHandler, notify client.NotifyPendingCount) error {
        for {
            var (
                id      int64
                payload string
            )
            // Take the oldest row for this instance.
            err := db.QueryRowContext(ctx,
                `SELECT id, payload FROM pending_messages
                   WHERE instance_id = $1
                   ORDER BY created_at, id
                   LIMIT 1`,
                instanceID,
            ).Scan(&id, &payload)
            if err == sql.ErrNoRows {
                return nil // queue drained
            }
            if err != nil {
                return err
            }

            // Rebuild a sendable message from the stored body.
            msg, err := builder.NewMessage(builder.MessageOptions{Text: payload})
            if err != nil {
                return err
            }

            // Re-inject it into the connection. Do NOT send it yourself.
            deliver(msg)

            // Delete on success so it is not redelivered after a restart.
            if _, err := db.ExecContext(ctx,
                `DELETE FROM pending_messages WHERE id = $1`, id,
            ); err != nil {
                return err
            }

            var depth int
            if err := db.QueryRowContext(ctx,
                `SELECT count(*) FROM pending_messages WHERE instance_id = $1`,
                instanceID,
            ).Scan(&depth); err != nil {
                return err
            }
            if err := notify(depth); err != nil {
                return err
            }
        }
    }
}
```

> 💡 Deleting **after** `deliver(msg)` is the durability guarantee: a row only leaves the table once it has been handed back into the connection. If the process dies mid-flush, the undeleted rows are still in the table and are picked up on the next flush. For an at-least-once guarantee against rebuilds that themselves restart, you can instead mark a row in-flight (e.g. a `claimed_at` column) and delete it only after the ACK is observed in your `OutboundHandler`.

---

## 🔌 Wiring it into a connection

Pass both hooks on `client.ClientListenerOptions`. Note that **`Version` is required per client** — every connection it opens inherits that one HL7 version, and `SendMessage` rejects any body whose `MSH.12` differs. Set `AutoConnect: ptr(false)` so you control when the first flush runs; the queue drains when you `Connect()`.

```go
import "github.com/Bugs5382/go-hl7/client/client"

func ptr[T any](v T) *T { return &v }

// Version is REQUIRED and pins the client to a single HL7 version.
c, _ := client.NewClient(client.ClientOptions{Version: "2.7", Host: "hl7.example.com"})

const instanceID = "adt-pod-1" // unique per client instance / pod

outbound, _ := c.CreateConnection(
    client.ClientListenerOptions{
        Port:           ptr(5678),
        AutoConnect:    ptr(false),
        EnqueueMessage: enqueueMessage(ctx, db, instanceID),
        FlushQueue:     flushQueue(ctx, db, instanceID),
    },
    func(res *client.InboundResponse) error {
        // Inspect MSA.1 ("AA" = accepted) as usual.
        return nil
    },
)

_ = outbound.Connect() // flushes anything left in pending_messages from a prior run
```

Because `pending_messages` is persistent, a connection that comes up after a crash flushes whatever the previous process left behind — that is the whole point of backing the queue with a database.

---

## 🍃 MongoDB and MySQL

The hook contract is identical; only the storage calls change. The shape is always the same three moves:

1. **`EnqueueMessage`** — store `m.String()`, then `notify(depth)`.
2. **`FlushQueue`** — read oldest-first, `deliver(msg)` each rebuilt body, delete on success, `notify(remaining)`.
3. **Tag every record** with an instance identifier.

| Store | Persist (`EnqueueMessage`) | Drain (`FlushQueue`) |
|---|---|---|
| 🐘 PostgreSQL | `INSERT INTO pending_messages …` | `SELECT … ORDER BY created_at LIMIT 1`, then `DELETE` |
| 🐬 MySQL | same `database/sql` API, `?` placeholders instead of `$1` | same query shape; `AUTO_INCREMENT` id |
| 🍃 MongoDB | `InsertOne` a `{payload, instanceID, createdAt}` document | `FindOneAndDelete` sorted by `createdAt` |

For MySQL the code above is almost unchanged — swap the `$1`/`$2` placeholders for `?` and use a MySQL driver (`github.com/go-sql-driver/mysql`). For MongoDB, replace the SQL with the driver's `InsertOne` / `FindOneAndDelete` and count with `CountDocuments`. The `MessageItem` → `String()` → `builder.NewMessage` round-trip is the same everywhere.

---

## 🛟 Operational notes

- **Return store errors.** Both hooks return `error`; surface a failed insert or read rather than dropping a message silently.
- **Order matters.** Drain oldest-first (`created_at, id`) so messages leave in the order they arrived.
- **One row per instance.** Always scope queries by `instance_id`; never let a shared table cross-deliver between pods.
- **Pick your delivery guarantee.** Delete-after-deliver is at-least-once against process restarts. If you need to survive a rebuild that fails mid-send, claim rows and delete them only after the ACK lands in your `OutboundHandler`.
- **`Version` is per client.** Every connection inherits the client's single HL7 version; a rebuilt message whose `MSH.12` differs is rejected by `SendMessage` before it goes out.

> 📖 See also the [Client walkthrough](../client/index.md) for the in-memory default and the Redis variant, and the [Parser page](../parser/index.md) for how `builder.NewMessage` rebuilds a message from stored text.
