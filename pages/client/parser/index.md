# go-hl7 Client :: Parser

## Introduction

This library serves as both the **builder** and **parser** for HL7 messages. There are no strict configuration requirements — the behavior is determined entirely by how it's initialized. While it is designed to be used with its sister package, [`server`](../../server/index.md), it functions independently.

> ⚠️ **Note**: The parser strictly requires valid HL7 strings. Any errors or malformed segments will cause parsing to fail. It's strongly recommended to write thorough unit tests for your application to validate HL7 message structures before consuming them.

If you're using the [client](../client/index.md) or [builder](../builder/index.md), be aware of what messages you're receiving on which ports to ensure proper parsing.

## Table of Contents

1. [Introduction](#introduction)
2. [Basic Usage](#basic-usage)
   - [Single Message](#single-message)
   - [Batch Message](#batch-message)
   - [File-Based Parsing](#file-based-parsing)
3. [Reading values](#reading-values)
4. [Recommended Use](#recommended-use)

## Basic Usage

The same `builder.Message` type powers parsing. Construct one from raw text and read fields with dotted HL7 paths.

### Single Message

Use this when you're dealing with a single HL7 message string:

```go
import (
    "strings"

    "github.com/Bugs5382/go-hl7/client/builder"
)

hl7 := strings.Join([]string{
    "MSH|^~\\&|||||20081231||ADT^A01^ADT_A01|12345||2.7",
    "EVN||20081231",
}, "\r")

message, err := builder.NewMessage(builder.MessageOptions{Text: hl7})
if err != nil { /* malformed HL7 */ }

msh_9_1 := message.Get("MSH.9.1").String() // ADT
```

### Batch Message

For HL7 batches that include multiple messages:

```go
hl7 := strings.Join([]string{
    "BHS|^~\\&|||||20231208",
    "MSH|^~\\&|||||20231208||ADT^A01^ADT_A01|CONTROL_ID||2.7",
    "EVN||20081231",
    "BTS|1",
}, "\r")

batch, _ := builder.NewBatch(builder.BatchOptions{Text: hl7})

for _, message := range batch.Messages() {
    msh_9_1 := message.Get("MSH.9.1").String() // ADT
    evn_2 := message.Get("EVN.2").String()     // 20081231

    _ = msh_9_1
    _ = evn_2
    // your logic here...
}
```

### File-Based Parsing

To parse HL7 messages from a file, give `FileOptions` a path.

**From file path:**

```go
fileBatch, _ := builder.NewFileBatch(builder.FileOptions{
    FullFilePath: "ADT.20081231.hl7",
})
```

Then extract messages just like a regular batch:

```go
for _, message := range fileBatch.Messages() {
    msh_9_1 := message.Get("MSH.9.1").String()
    evn_2 := message.Get("EVN.2").String()

    _ = msh_9_1
    _ = evn_2
    // your logic here...
}
```

## Reading values

Reads of a missing path return a shared **empty node**, so chained `Get(...).String()` is always safe (it yields `""`). Value coercions return `(T, ok)`:

```go
n, ok := message.Get("OBX.5").Int()     // (int, bool)
f, ok := message.Get("OBX.5").Float()   // (float64, bool)
b, ok := message.Get("PID.30").Bool()   // "Y"/"N" -> (bool, bool)
t, ok := message.Get("PID.7").Date()    // HL7 date -> (time.Time, bool)

message.Get("PID.3").Index(0).Index(3).String() // repetition 0, component 3
message.Exists("PV1.44")                          // false when the path resolves to the empty node
```

> ⚠️ The parser is strict — malformed HL7 yields an `error` from the constructor, and some structural read failures panic with an `HL7Error`. `NewMessage` rejects a body that doesn't begin with `MSH` or one carrying multiple MSH segments (use `Batch` for that). `NewBatch` rejects a single‑MSH body. `NewMessage`, `NewBatch`, and `NewFileBatch` all return `(T, error)`.

## Recommended Use

This parser is typically used on the **server or broker** side of your architecture. The [`server`](../../server/index.md) package depends on it: `req.GetMessage()` returns the same `*builder.Message` type documented here.

> ℹ️ On the client side, replies are parsed _before_ being handed to the `OutboundHandler` (`func(res *client.InboundResponse) error`) — so you typically won't need to call the parser directly in client response logic; just read `res.GetMessage()`.
