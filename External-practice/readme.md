# Webserver Message Buffer — Go

## Pre-Interview Prep
Please show up to the interview with a working, runnable solution. You can use whatever editor, tooling, or LLM assistance you like while preparing — but you should understand the code you ship, because we will read it together and ask questions.

This directory is a Go starter. If you'd rather solve the problem in a different language, ask your recruiter for a starter in your language of choice.

## The Problem
Write a simple HTTP server that brokers a message queue. The server has two endpoints:

**`GET /message`** — Retrieve a single message from the queue.
- `200 OK` on success (return the message in the response body)
- `404 Not Found` if the queue is empty
- `5XX` for any other error

**`POST /message`** — Add a new message to the end of the queue.
- `201 Created` if the message was queued
- `429 Too Many Requests` if the queue is full
- `5XX` for any other error

The server should listen on a hardcoded address and have a hardcoded max queue size.

There are implementation decisions that are intentionally abiguous. Please make your own assumption and come prepared to talk about the assumptions made.

## My Assumptions

Since the problem is intentionally ambiguous, here are the key decisions I made. Full rationale and trade-offs in [`ASSUMPTIONS.md`](./ASSUMPTIONS.md).

- **Plain-text body, not JSON** — the smoke-test posts `-d 'hello'` and expects the raw message back.
- **FIFO queue backed by a buffered channel** — a bounded, thread-safe FIFO by design, so no mutex.
- **Non-blocking Push/Pop** — empty GET returns `404` and full POST returns `429` immediately, no long-polling.
- **Error-to-status mapping** — `ErrEmpty` → `404`, `ErrFull` → `429`, empty body → `400`, anything else → `500` (real cause logged, not leaked).
- **GET pops the message** (true queue semantics); request body capped at 1 MiB.
- **Persistence is pluggable** — the `MessageBuffer` interface has two implementations: an in-memory buffer and a crash-safe file-backed buffer (see [Implementations](#implementations) below).
- **Out of scope by choice** — auth, TLS; production concerns beyond this broker exercise.

## What's Provided

`main.go` ships with scaffolding so you can focus on the interesting parts:

- `Address` and `MaxQueueSize` constants — feel free to adjust them.
- A `Message` struct with a `Body` field.
- `ErrEmpty` and `ErrFull` sentinel errors.
- A `MessageBuffer` interface (`Push`, `Pop`).
- An `OpenMessageBuffer` constructor stub that currently panics.
- A `main()` that wires the constructor and panics at the spot where the HTTP server needs to start.

Running `go run main.go` as-is will hit one of the `panic("TODO: ...")` calls.

## What You'll Build

1. A concrete implementation of `MessageBuffer`, returned from `OpenMessageBuffer`.
2. An HTTP server bound to `Address` that exposes `GET /message` and `POST /message` with the status codes described above.

## Constraints

- **Standard library only.** `net/http` is recommended. Third-party routers and frameworks (gin, chi, echo, fiber, etc.) are not permitted for this exercise — we want to see how you reason about request handling, not how well you wire a framework.
- LLM assistance is fine while preparing, but you should be able to explain every line.
- DO NOT COMMENT YOUR CODE.

## Running

```bash
go run main.go
```

Once the server is up, smoke-test it:

```bash
curl -i -X POST -d 'hello' http://localhost:8080/message    # expect 201
curl -i http://localhost:8080/message                       # expect 200 with the message
curl -i http://localhost:8080/message                       # expect 404 (queue empty)
```

## During the Interview

Be prepared to:
- Demo the server end-to-end.
- Walk through your design decisions — why this data structure, why this concurrency model, why these error mappings. 
- Be ready to extend your implementation with us. We may ask you to add behavior, handle a new edge case, or refactor part of the design.

We care more about **how** you think and code than about whether you finished every extension. Show us your process.

---

## Implementations

Both satisfy the same `MessageBuffer` interface (`Push`, `Pop`), so the HTTP layer never changes. Only the backing store differs.

### `memoryBuffer` — in-memory, buffered channel

- A bounded FIFO backed by `chan Message`. Thread-safe by construction, no mutex.
- Fastest option, zero dependencies, nothing touches disk.
- **Does not survive the process ending.** When the server stops, the queue is gone — this is the right default for an ephemeral broker.
- Constructed via `NewMemoryBuffer(maxSize)`.

### `fileBuffer` — crash-safe, write-ahead log (WAL)

The implementation wired into `main` (`OpenMessageBuffer` → `NewFileBuffer`). It keeps the queue in memory **and** mirrors every change to an append-only log file (`queue.log`) so the queue survives a restart — **including a hard crash or `kill -9`**.

How it works:

- **Every `Push` and `Pop` is written to the log and `fsync`ed to disk _before_ the operation is acknowledged.** So once a POST returns `201`, that message is durable; once a GET returns `200`, that pop is durable. No in-memory-only window.
- Each line is a JSON record: `{"op":"push","body":"..."}` or `{"op":"pop"}`.
- On startup, the log is **replayed** to rebuild the exact queue state, then **compacted** (rewritten to only the live messages via a temp file + atomic rename) so it can't grow forever.
- A **torn tail** — a half-written final record left by a crash mid-write — is detected and ignored on replay, rather than corrupting startup.
- A failed write is **rolled back** (`Truncate`) so a partial record is never left behind at runtime.

Trade-off: an `fsync` per request is a disk round-trip, so this is slower than the in-memory buffer. That latency is the price of surviving a crash. For a high-throughput broker you'd batch/group-commit syncs or move to an external durable store.

### Why not Redis / an external queue?

An external store (Redis, SQS, RabbitMQ) is the other way to survive a restart, and it also lets multiple server instances share one queue — something neither in-process option can do. It was ruled out here because it adds a second process and a third-party client, which crosses this exercise's "standard library only" line. The `fileBuffer` gives crash durability while staying 100% stdlib and single-process.

### What survives what

| Scenario                               | `memoryBuffer` | `fileBuffer` (WAL) |
|----------------------------------------|:--------------:|:------------------:|
| Normal operation                       | ✅             | ✅                 |
| Graceful shutdown (Ctrl+C / SIGTERM)   | ❌             | ✅                 |
| Hard crash / `kill -9` / power loss    | ❌             | ✅                 |
| Share one queue across instances       | ❌             | ❌ (needs external store) |

---

## Testing

### Automated

```bash
go test -race ./...
```

Covers both implementations: FIFO ordering, `ErrEmpty`/`ErrFull` mapping, capacity, handler status codes, routing, and concurrency under the race detector. The `fileBuffer` adds crash-recovery tests: messages survive a reload without `Close`, popped messages do **not** reappear, a torn tail is ignored, the log compacts on startup, and a rejected (full) push is never persisted.

### Manual — prove crash durability with curl

This is the scenario to demo. It shows new messages surviving a `kill -9` and already-consumed messages *not* coming back.

```bash
# Terminal 1 — start the server (uses the fileBuffer)
go run main.go
```

```bash
# Terminal 2
U=http://localhost:8080/message

# Fill the queue (MaxQueueSize=10), then one more POST overflows
for i in $(seq 1 10); do curl -s -o /dev/null -X POST -d "msg-$i" $U; done
curl -i -X POST -d 'overflow' $U  # 429 (full)

# Drain it all back out (FIFO: msg-1 first)
for i in $(seq 1 10); do curl -s $U; echo; done

# Now add two fresh messages
curl -i -X POST -d 'lucas'   $U   # 201
curl -i -X POST -d 'lucas_2' $U   # 201

# Simulate a hard crash — NO graceful shutdown, nothing gets a chance to flush
kill -9 $(lsof -ti :8080)

# Inspect the durable log: 10 pushes, 10 pops, 2 pushes
cat queue.log
```

```bash
# Terminal 1 — restart
go run main.go
```

```bash
# Terminal 2 — the queue came back from disk
curl -s $U                        # lucas    (survived the kill -9)
curl -s $U                        # lucas_2
curl -i $U                        # 404 empty (alpha/beta/gamma did NOT reappear — their pops were durable)
```

Note: `queue.log` is the on-disk state. Delete it (`rm -f queue.log`) to start from an empty queue. It's git-ignored (`*.log`).
