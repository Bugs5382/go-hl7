# httpadapter

An optional, dependency-free bridge between a standard `net/http` application
and a go-hl7 outbound MLLP connection. It is the Go analog of the node
[`fastify-hl7`](https://github.com/Bugs5382/fastify-hl7) plugin: where that
plugin decorates a Fastify instance with an HL7 client, this package hands a
`net/http` app a small set of `http.Handler` values wired to a `client.Connection`.

The package uses only the standard library plus go-hl7 itself. Nothing else in
the module imports it, so it adds no cost to consumers that do not want it.

## What it provides

- `SendHandler()` — a `POST` handler that reads an HL7 v2 (ER7) message from the
  request body, sends it over MLLP, and writes the ACK back as the response body.
- `HealthHandler()` — a `GET` handler that reports connection status as JSON
  (`{"connected":bool,"port":int}`), `200` when up and `503` otherwise.

Sends are serialized so a single in-flight message is correlated with the next
ACK, honoring the request context and a configurable timeout (`WithTimeout`,
default 10s).

## Usage

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Bugs5382/go-hl7/client/client"
	"github.com/Bugs5382/go-hl7/client/httpadapter"
)

func ptr[T any](v T) *T { return &v }

func main() {
	c, err := client.NewClient(client.ClientOptions{Version: "2.7", Host: "127.0.0.1"})
	if err != nil {
		log.Fatal(err)
	}

	// Create the adapter first, then register its ack handler on the
	// connection so ACKs are routed back to it.
	a := httpadapter.New(httpadapter.WithTimeout(5 * time.Second))
	conn, err := c.CreateConnection(
		client.ClientListenerOptions{Port: ptr(3000)},
		a.AckHandler(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	a.Bind(conn)

	mux := http.NewServeMux()
	mux.Handle("POST /hl7", a.SendHandler())
	mux.Handle("GET /healthz", a.HealthHandler())

	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

Send a message and read the ACK:

```sh
curl -X POST --data-binary @message.hl7 \
  -H 'Content-Type: application/hl7-v2' \
  http://localhost:8080/hl7
```

## Responses

`SendHandler`:

| Status | Meaning |
|---|---|
| `200` | ACK returned in the body (`Content-Type: application/hl7-v2`) |
| `400` | Request body is not a parseable HL7 message |
| `405` | Method was not `POST` |
| `502` | The send failed (for example a version mismatch or transport error) |
| `503` | The adapter has not been bound to a connection |
| `504` | No ACK arrived before the timeout |

## Scope

This adapter intentionally covers the common case: forward a single HL7 message
and return its ACK, plus a health probe. It does not re-export message/batch
builders or manage inbound listeners the way `fastify-hl7` does — in Go those
needs are met by importing `client/hl7`, `client/builder`, and `server`
directly. The request body is parsed as a single message; batches and file
batches are out of scope.
