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
- **Out of scope by choice** — auth, TLS, graceful shutdown, persistence; production concerns beyond an in-memory broker.

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
