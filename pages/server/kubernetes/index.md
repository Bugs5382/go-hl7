# ☸️ Running the go-hl7 `server` on Kubernetes

> A practical reference for scaling the HL7 listener horizontally across multiple pods, decoupling work with Redis or RabbitMQ, and choosing where TLS terminates.

```mermaid
flowchart LR
    A[👩‍⚕️ EMR / Sender] -- TCP/MLLP TLS --> B[(NLB / Service<br/>type=LoadBalancer)]
    B -- "TCP (plain or TLS)" --> C{HL7 listener pods}
    C -->|push| Q[(Redis / RabbitMQ)]
    Q -->|consume| W[Worker Deployment]
    W --> D[(EMR / FHIR / DB)]
```

The pattern is always the same:

1. **Listeners** terminate the MLLP frame, parse it, ACK fast, and shove the message onto a queue.
2. **Workers** pull from the queue and do the real work (transform, route to FHIR, write to DB).
3. **Queue** is durable and shared, so any worker pod can pick up any message and pods can come and go without losing data.

## 🧾 Table of Contents

1. [Why split listener and worker?](#-why-split-listener-and-worker)
2. [TLS termination — where?](#-tls-termination--where)
3. [The HL7 listener Deployment + Service](#-the-hl7-listener-deployment--service)
4. [Pattern A — TLS terminated at the Service / NLB](#-pattern-a--tls-terminated-at-the-service--nlb)
5. [Pattern B — TLS terminated at the Go app (mTLS, hospital networks)](#-pattern-b--tls-terminated-at-the-go-app-mtls-hospital-networks)
6. [Wiring up Redis as the durable queue](#-wiring-up-redis-as-the-durable-queue)
7. [Wiring up RabbitMQ as the durable queue](#-wiring-up-rabbitmq-as-the-durable-queue)
8. [Worker deployment](#-worker-deployment)
9. [Health checks, sticky sessions, scaling](#-health-checks-sticky-sessions-scaling)
10. [Sizing & limits](#-sizing--limits)

---

## 🧩 Why split listener and worker?

| Concern | Listener pod | Worker pod |
|---|---|---|
| Goal | ACK fast (sub‑second). | Do the actual work (transform, persist, forward). |
| State | Stateless except for in-flight TCP buffers. | Stateless; idempotent over the queue. |
| Scaling trigger | Inbound connection count / CPU. | Queue depth. |
| Restart cost | Drops in-flight TCP frames (sender retries). | Drops nothing — message stays on the queue. |
| TLS | Yes (or terminated at Service). | No (talks Redis/RabbitMQ over the cluster network). |

Putting heavy work in the listener handler means a downstream FHIR slowdown back-pressures the sender — and an HL7 sender that gets stuck retrying for 30 seconds tends to flood you with duplicates. The "ACK first, queue, work later" pattern is the most important architectural decision in this whole document.

---

## 🔐 TLS termination — where?

Two valid patterns. Pick **one** per environment.

```mermaid
flowchart LR
    subgraph patternA["Pattern A · Terminate at the Service / NLB"]
      a1[Sender 🔒] -- TLS --> a2[(LB / Ingress<br/>terminates TLS)]
      a2 -- plain TCP --> a3[Listener pod]
    end

    subgraph patternB["Pattern B · Terminate inside the Go app"]
      b1[Sender 🔒] -- TLS --> b2[(LB passthrough)]
      b2 -- TLS --> b3[Listener pod<br/>go-hl7 server with TLS]
    end
```

| | **A · Terminate at LB / Service** | **B · Terminate in the Go app** |
|---|---|---|
| Cert lives on | The load balancer (cloud LB, envoy, nginx, etc.) | Each listener pod (mounted secret) |
| mTLS supported | Often, depending on LB | Yes, natively via the `TLS` option |
| Easier ops | ✅ One cert renewal point | ❌ Cert distributed to every pod |
| Sees plaintext on the wire inside the cluster | ✅ Yes | ❌ No (encrypted right up to the Go process) |
| Good fit for | Public-facing, simple TLS | Hospital integrations requiring mTLS / strict cert pinning |

> 🛡️ **mTLS in hospital networks.** Many hospital integration teams insist on **client-cert auth all the way to the application** (no TLS-stripping load balancers in between). That points you at Pattern B even if Pattern A would otherwise be operationally easier.

Both patterns are below.

---

## 🚀 The HL7 listener Deployment + Service

Common to both TLS patterns. We show ADT on port 6661 and ORU on port 6662 — adjust to your traffic shape. The container image is your compiled Go binary (`CGO_ENABLED=0 go build` into a minimal `scratch`/`distroless` image works well — the library is standard-library only).

### `Deployment`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hl7-listener
  labels: { app: hl7-listener }
spec:
  replicas: 3
  selector:
    matchLabels: { app: hl7-listener }
  template:
    metadata:
      labels: { app: hl7-listener }
    spec:
      containers:
        - name: hl7-listener
          image: ghcr.io/your-org/hl7-listener:1.0.0
          ports:
            - { name: adt, containerPort: 6661 }
            - { name: oru, containerPort: 6662 }
          env:
            - { name: APP_ENV, value: "production" }
            - { name: REDIS_URL, valueFrom: { secretKeyRef: { name: hl7-secrets, key: redis-url } } }
            # For Pattern B (mTLS in the Go app), also mount certs.
          readinessProbe:
            tcpSocket: { port: 6661 }
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            tcpSocket: { port: 6661 }
            initialDelaySeconds: 30
            periodSeconds: 30
          resources:
            requests: { cpu: "200m", memory: "128Mi" }
            limits:   { cpu: "1000m", memory: "256Mi" }
```

> ✅ Use **`tcpSocket`** probes, not `httpGet` — the listener doesn't speak HTTP. The MLLP framing makes a real "is the app accepting connections?" probe trivial: if the TCP handshake completes, the pod is healthy.

### `Service`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: hl7-listener
  labels: { app: hl7-listener }
spec:
  type: LoadBalancer            # or ClusterIP behind an Ingress / Gateway API.
  selector: { app: hl7-listener }
  ports:
    - { name: adt, port: 6661, targetPort: adt, protocol: TCP }
    - { name: oru, port: 6662, targetPort: oru, protocol: TCP }
  # 🪝 Sticky sessions: keep one TCP connection on one pod for its lifetime.
  sessionAffinity: ClientIP
  externalTrafficPolicy: Local  # preserves the sender's IP for logging / mTLS CN checks
```

> 🪝 **`sessionAffinity: ClientIP`** matters more than you'd think: HL7 senders typically open one TCP connection per shift and reuse it for thousands of messages. Affinity keeps that connection on a single pod, which means the per-socket `MLLPCodec` buffer state stays consistent.

---

## 🅰️ Pattern A — TLS terminated at the Service / NLB

The simplest setup: the cloud LB (or an ingress controller / service mesh sidecar) handles TLS, and your Go pods speak plain TCP inside the cluster.

```yaml
# AWS NLB example: TLS listener that decrypts and forwards plain TCP.
apiVersion: v1
kind: Service
metadata:
  name: hl7-listener
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert: "arn:aws:acm:us-east-1:123:certificate/abc"
    service.beta.kubernetes.io/aws-load-balancer-ssl-ports: "6661,6662"
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: "tcp"
spec:
  type: LoadBalancer
  selector: { app: hl7-listener }
  ports:
    - { name: adt-tls, port: 6661, targetPort: 6661, protocol: TCP }
    - { name: oru-tls, port: 6662, targetPort: 6662, protocol: TCP }
  externalTrafficPolicy: Local
  sessionAffinity: ClientIP
```

The Go app:

```go
import "github.com/Bugs5382/go-hl7/server"

func ptr[T any](v T) *T { return &v }

srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("0.0.0.0")}) // ⬅️ no TLS
srv.CreateInbound(server.ListenerOptions{Version: "2.7", Port: ptr(6661)}, handleADT)
srv.CreateInbound(server.ListenerOptions{Version: "2.7", Port: ptr(6662)}, handleORU)
select {} // keep the process alive
```

✅ **Pros**: cert lives in ACM (or cert-manager); rotation is automatic; pods are simpler.
❌ **Cons**: Inside-cluster traffic is plaintext (mitigate with a service mesh if needed); mTLS support depends on the LB.

---

## 🅱️ Pattern B — TLS terminated at the Go app (mTLS, hospital networks)

The LB passes encrypted bytes through to the pod, and the go-hl7 `server` itself terminates the TLS handshake — including verifying the client cert.

### Mount the certs as a Kubernetes `Secret`

```bash
kubectl create secret generic hl7-tls \
  --from-file=tls.key=server-key.pem \
  --from-file=tls.crt=server-crt.pem \
  --from-file=ca.crt=trusted-client-ca.pem
```

### Reference the secret in the `Deployment`

```yaml
spec:
  template:
    spec:
      containers:
        - name: hl7-listener
          # ...
          volumeMounts:
            - { name: tls, mountPath: /etc/hl7/tls, readOnly: true }
      volumes:
        - name: tls
          secret:
            secretName: hl7-tls
            defaultMode: 0400
```

### And configure `Server` to use them

```go
import (
    "os"

    "github.com/Bugs5382/go-hl7/server"
)

key, _ := os.ReadFile("/etc/hl7/tls/tls.key")
crt, _ := os.ReadFile("/etc/hl7/tls/tls.crt")
ca, _ := os.ReadFile("/etc/hl7/tls/ca.crt")

srv, _ := server.NewServer(&server.ServerOptions{
    BindAddress: ptr("0.0.0.0"),
    TLS: &server.TLSConfig{
        Key:  key,
        Cert: crt,

        // 🤝 mTLS: demand a client cert from every sender.
        RequestCert: true,
        CA:          ca,
    },
})

srv.CreateInbound(server.ListenerOptions{Version: "2.7", Port: ptr(6661)}, handleADT)
srv.CreateInbound(server.ListenerOptions{Version: "2.7", Port: ptr(6662)}, handleORU)
```

### And the Service / LB needs to **passthrough**

```yaml
apiVersion: v1
kind: Service
metadata:
  name: hl7-listener
  annotations:
    # AWS NLB: no ssl-cert annotation -> raw TCP passthrough.
    service.beta.kubernetes.io/aws-load-balancer-type: "nlb"
    service.beta.kubernetes.io/aws-load-balancer-backend-protocol: "tcp"
spec:
  type: LoadBalancer
  selector: { app: hl7-listener }
  ports:
    - { name: adt, port: 6661, targetPort: 6661, protocol: TCP }
    - { name: oru, port: 6662, targetPort: 6662, protocol: TCP }
  externalTrafficPolicy: Local
  sessionAffinity: ClientIP
```

✅ **Pros**: end-to-end TLS / mTLS; you can read peer certificate details inside the handler via `req.GetSocket()` (cast to `*tls.Conn`); no third-party LB sees plaintext.
❌ **Cons**: cert rotation across pods (use [`cert-manager`](https://cert-manager.io/) + a rolling restart, or [Reloader](https://github.com/stakater/Reloader) to auto-restart on Secret change).

---

## 🟥 Wiring up Redis as the durable queue

Run Redis as a separate workload (Bitnami / managed Redis / Elasticache, etc.). The listener handler **acknowledges first**, then publishes. The `server` package has no queue of its own — your handler is the boundary, so push to the store inside it:

```go
import (
    "context"
    "encoding/json"
    "time"

    "github.com/Bugs5382/go-hl7/server"
)

ctx := context.Background()
// rdb is your *redis.Client built from os.Getenv("REDIS_URL").

srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("0.0.0.0")})

srv.CreateInbound(server.ListenerOptions{Version: "2.7", Port: ptr(6661), Name: "IB_ADT"}, func(req *server.InboundRequest, res server.ResponseSender) error {
    msg := req.GetMessage()

    // 1️⃣  Push the parsed message onto the queue. Sub-millisecond.
    env, _ := json.Marshal(map[string]string{
        "receivedAt": time.Now().UTC().Format(time.RFC3339),
        "sourceIp":   req.GetSocket().RemoteAddr().String(),
        "controlId":  msg.Get("MSH.10").String(),
        "raw":        msg.String(),
    })
    if err := rdb.LPush(ctx, "hl7:adt", env).Err(); err != nil {
        return err
    }

    // 2️⃣  ACK the sender. They unblock immediately.
    return res.SendResponse("AA")
})
```

> 💡 **Use a separate list per workflow** (`hl7:adt`, `hl7:oru`, `hl7:siu`) so workers can scale independently.

A minimal Redis deployment for dev/staging:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata: { name: redis }
spec:
  serviceName: redis
  replicas: 1
  selector: { matchLabels: { app: redis } }
  template:
    metadata: { labels: { app: redis } }
    spec:
      containers:
        - name: redis
          image: redis:7-alpine
          args: ["--appendonly", "yes"]   # 🪛 enable AOF for durability
          ports: [{ containerPort: 6379 }]
          volumeMounts: [{ name: data, mountPath: /data }]
  volumeClaimTemplates:
    - metadata: { name: data }
      spec: { accessModes: [ReadWriteOnce], resources: { requests: { storage: 5Gi } } }
---
apiVersion: v1
kind: Service
metadata: { name: redis }
spec:
  selector: { app: redis }
  ports: [{ port: 6379, targetPort: 6379 }]
```

> 🚨 **Production**: use a managed Redis (Elasticache, Memorystore, Upstash) or run Sentinel/Cluster. A single-pod Redis can lose un-flushed messages on a node failure.

---

## 🟧 Wiring up RabbitMQ as the durable queue

When you need topic routing (one HL7 message → multiple consumers) or persistent durable queues with confirms, RabbitMQ is the better fit:

```go
import (
    amqp "github.com/rabbitmq/amqp091-go"
    "github.com/Bugs5382/go-hl7/server"
)

conn, _ := amqp.Dial(os.Getenv("RABBITMQ_URL"))
ch, _ := conn.Channel()
_ = ch.Confirm(false) // put the channel in confirm mode
_, _ = ch.QueueDeclare("hl7.adt", true /* durable */, false, false, false, nil)
confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))

srv, _ := server.NewServer(&server.ServerOptions{BindAddress: ptr("0.0.0.0")})

srv.CreateInbound(server.ListenerOptions{Version: "2.7", Port: ptr(6661)}, func(req *server.InboundRequest, res server.ResponseSender) error {
    body := []byte(req.GetMessage().String())

    // Publish with a confirm so we know the broker has it before we ACK.
    if err := ch.Publish("", "hl7.adt", false, false, amqp.Publishing{
        DeliveryMode: amqp.Persistent,
        ContentType:  "text/plain",
        Body:         body,
    }); err != nil {
        return err
    }
    if c := <-confirms; !c.Ack {
        return fmt.Errorf("broker did not confirm publish")
    }

    return res.SendResponse("AA")
})
```

> ⚠️ Wait for the **confirm** before sending the ACK — otherwise you can ACK a sender, lose the broker connection, and silently drop the message.

---

## 🟦 Worker deployment

Workers are completely separate from the listener — they don't bind any TCP ports, just consume the queue:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata: { name: hl7-worker }
spec:
  replicas: 2
  selector: { matchLabels: { app: hl7-worker } }
  template:
    metadata: { labels: { app: hl7-worker } }
    spec:
      containers:
        - name: worker
          image: ghcr.io/your-org/hl7-worker:1.0.0
          env:
            - { name: REDIS_URL, valueFrom: { secretKeyRef: { name: hl7-secrets, key: redis-url } } }
          resources:
            requests: { cpu: "200m", memory: "128Mi" }
            limits:   { cpu: "1000m", memory: "512Mi" }
```

…and a minimal Redis worker — note it uses the same `builder` parser as the listener to rebuild the `Message`:

```go
import (
    "context"
    "encoding/json"

    "github.com/Bugs5382/go-hl7/client/builder"
)

ctx := context.Background()
// rdb is your *redis.Client built from os.Getenv("REDIS_URL").

for {
    // BLPOP blocks until a message is available — no busy-loop.
    popped, err := rdb.BLPop(ctx, 5*time.Second, "hl7:adt").Result()
    if err != nil || len(popped) < 2 {
        continue
    }

    var env struct {
        Raw       string `json:"raw"`
        ControlId string `json:"controlId"`
    }
    _ = json.Unmarshal([]byte(popped[1]), &env)

    msg, err := builder.NewMessage(builder.MessageOptions{Text: env.Raw})
    if err != nil {
        continue
    }

    _ = persistAdmission(msg, env)  // 🩺 the actual work
    _ = forwardToFhir(msg)          // 🌐 downstream system
}
```

Scale workers based on **queue depth** (e.g. with KEDA) rather than CPU — if FHIR slows down, you want more workers, not more listener pods.

---

## ❤️ Health checks, sticky sessions, scaling

| Concern | Recipe |
|---|---|
| Probes | `tcpSocket` on the HL7 port. The MLLP listener doesn't speak HTTP. |
| Sticky sessions | `sessionAffinity: ClientIP` on the `Service`. Keeps a sender's connection on one pod (and so its MLLPCodec buffer) for the connection's lifetime. |
| Pod restarts | Use `terminationGracePeriodSeconds: 60` and a `preStop` that calls your shutdown hook (`IB.Close()`) so in-flight ACKs finish before the pod dies. |
| HPA on the listener | Scale on **connection count** if you can; CPU is a poor proxy for "I'm overloaded with senders". |
| HPA on the worker | Scale on **queue depth** (KEDA Redis or RabbitMQ scaler). |
| TLS rotation (Pattern B) | `cert-manager` + [Reloader](https://github.com/stakater/Reloader) to roll pods on Secret change. |

---

## 📏 Sizing & limits

For the typical hospital workload (~60K ADT/day with bursts to a few hundred per minute) a starter shape is:

| Workload | Replicas | CPU | Memory |
|---|---|---|---|
| `hl7-listener` | 2–3 | 200m / 1 CPU | 128Mi / 256Mi |
| `hl7-worker` | 2–4 | 200m / 1 CPU | 128Mi / 512Mi |
| Redis (single, AOF) | 1 (managed/HA in prod) | 100m / 500m | 256Mi / 1 GiB |

A Go HL7 listener has a notably small footprint — no runtime, no GC pauses of note at this volume, and zero third-party dependencies in the hot path. If you exceed those numbers comfortably, the bottleneck is almost always downstream (FHIR / DB) rather than the HL7 listener — the `server` itself moves messages at hundreds-per-second on a single pod. See [`pages/server/performance/index.md`](../performance/index.md) for the full throughput notes.

---

## 🔗 See also

- [Performance & throughput notes](../performance/index.md) — what the `server` measures (`TotalReceived` / `TotalMessage`) and how to read it.
- [TLS & mTLS](../tls/index.md) — the underlying Go-side TLS / mTLS configuration that Pattern B uses.
- [Custom queues on the client side](../../client/client/index.md#custom-behavior-using-redis) — the symmetric pattern when **sending** HL7 from a Kubernetes pod.
