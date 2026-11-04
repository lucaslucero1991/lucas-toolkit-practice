# Assumptions & Interview Notes

Study notes for the message-buffer exercise. The prompt says the design is
"intentionally ambiguous" and to "come prepared to talk about the assumptions made."
This file is my prep: the decisions I made, why, and how to walk through the demo.

---

## Assumptions

### 1. Message body is plain text, not JSON
- **Ambiguity:** the prompt never specifies the payload format.
- **Evidence:** the smoke-test uses `curl -d 'hello'` and expects the GET to return
  the raw message, not a JSON object.
- **Decision:** read and write the raw body with `io.ReadAll` / `w.Write`. The
  `json:"body"` tag stays on the struct as future-proofing but is not exercised today.
- **Pivot:** switching to JSON is ~3 lines (`json.NewDecoder`/`Encoder`) plus a
  `Content-Type: application/json`.

### 2. FIFO queue backed by a buffered channel
- **Ambiguity:** data structure and concurrency model are explicitly left to me.
- **Decision:** `chan Message` with a fixed capacity. A buffered channel *is* a
  bounded FIFO queue, and it is thread-safe by design, so there is no mutex.
- **Alternative I can discuss:** slice + `sync.Mutex`, or a ring buffer with head/tail
  indices. (I also have the slice+mutex version on another branch.)

### 3. Non-blocking Push/Pop (`select` with `default`)
- **Ambiguity:** the prompt does not say whether a GET on an empty queue should block
  and wait, or respond immediately.
- **Decision:** do not block. `Pop` on empty returns `ErrEmpty` -> 404 immediately;
  `Push` on full returns `ErrFull` -> 429 immediately. This matches the required
  status codes, which imply an immediate response rather than a long poll.

### 4. Hardcoded capacity (10) and address (localhost:8080)
- **Ambiguity:** the prompt asks for a hardcoded address and max size and says
  "feel free to adjust them."
- **Decision:** `MaxQueueSize = 10`, `Address = "localhost:8080"` as constants.

### 5. Error-to-status-code mapping
- **Decision:**
  - `ErrFull` -> `429 Too Many Requests`
  - `ErrEmpty` -> `404 Not Found`
  - any other error -> `500 Internal Server Error` (satisfies "5XX for any other error")
- **Extra assumption:** an empty POST body -> `400 Bad Request`. The prompt does not
  mention this; I treat a contentless POST as a client error.

### 6. Request body size capped at 1 MiB
- **Ambiguity:** the prompt says nothing about payload size.
- **Decision:** `http.MaxBytesReader` at 1 MiB so a large body cannot exhaust memory.
  A body over the limit -> `400` (treated as a client error).

### 7. Internal errors are not leaked to the client
- **Decision:** on a 500, log the real error with `slog` and return a generic
  `"internal server error"` message. Maturity choice, not required by the prompt.

### 8. GET removes the message (pop, not peek)
- **Ambiguity:** "retrieve a single message" does not say whether the message stays.
- **Decision:** GET pops the message (true queue semantics). Two GETs against a single
  message return `200` then `404`.

### 9. Out of scope (deliberately)
- No auth, no TLS, no graceful shutdown, no persistence.
- These are production concerns, not part of an in-memory broker exercise. I can talk
  through how I'd add each one, but I did not implement them to keep the solution
  focused and within the "standard library only" constraint.

---

## End-to-End Demo Script

Run the server:

```bash
go run main.go
```

Then, in another terminal, walk through the happy path and each error case:

```bash
# 1. POST a message -> 201 Created
curl -i -X POST -d 'hello' http://localhost:8080/message

# 2. GET it back -> 200 OK, body is "hello"
curl -i http://localhost:8080/message

# 3. GET again on an empty queue -> 404 Not Found
curl -i http://localhost:8080/message

# 4. POST an empty body -> 400 Bad Request
curl -i -X POST -d '' http://localhost:8080/message

# 5. Fill the queue (capacity 10), then one more POST -> 429 Too Many Requests
for i in $(seq 1 10); do curl -s -X POST -d "msg-$i" http://localhost:8080/message; done
curl -i -X POST -d 'overflow' http://localhost:8080/message

# 6. A method that is not registered -> 405 Method Not Allowed
curl -i -X DELETE http://localhost:8080/message
```

What to say while demoing:
- Point out that the status code for each case is driven by a sentinel error
  (`ErrFull`, `ErrEmpty`) mapped in the handler, not by ad-hoc logic.
- Note that the 405 is free: Go's `net/http` method-based routing
  (`"POST /message"`, `"GET /message"`) rejects unregistered methods automatically.

---

## Tests to Highlight

```bash
go test -race ./...
```

- **Buffer unit tests:** FIFO ordering, empty -> `ErrEmpty`, full -> `ErrFull`,
  capacity frees up after a Pop.
- **Handler tests:** every status code (201, 400, 429, 200, 404) plus the 500 path,
  which is driven through the generated `MessageBufferMock` because the real buffer
  never returns an unexpected error.
- **Routing tests:** full request path via `httptest.NewServer(a.routes())`, including
  the 405 for an unregistered method.
- **Concurrency tests (run with `-race`):**
  - no message lost or duplicated under 50 concurrent pushes,
  - exactly `capacity` pushes accepted when 100 goroutines race a capacity-10 queue,
  - producers and consumers in parallel leave the queue empty at the end.
- **Why `-race` matters:** data races are non-deterministic; without the race detector
  a concurrency bug can pass by luck. `-race` instruments memory access and reports the
  race even when that run did not produce a wrong value.

---

## Likely Interview Questions (and my answers)

- **"Why a channel instead of a slice?"** A buffered channel is a bounded FIFO that is
  thread-safe by construction, so it removes the mutex and the whole class of
  check-then-act races. Trade-off: the non-blocking semantics need `select`/`default`.

- **"Why `errors.Is` instead of `==`?"** Both work today because the errors are not
  wrapped, but `errors.Is` keeps working if someone later wraps them with `%w`. It is
  the robust default.

- **"Why does `OpenMessageBuffer` return an error if it never fails?"** The signature
  came from the scaffolding and anticipates implementations that can fail to open
  (disk, network). I gave it a real use by validating `maxSize > 0` there.

- **"Why is the 500 path only tested with a mock?"** The in-memory buffer only returns
  `ErrFull`, `ErrEmpty`, or `nil`. The mock lets me force an unexpected error so I can
  exercise the defensive 500 branch, which exists because I code against the interface,
  not the concrete type.

- **"How would you make this production-ready?"** Graceful shutdown with
  `signal.NotifyContext` + `srv.Shutdown`, `ReadHeaderTimeout`/`IdleTimeout` against
  Slowloris, auth/TLS in front, per-client rate limiting, and persistence if messages
  must survive a restart.

- **"What happens under high concurrency?"** The channel serializes access; no message
  is lost or duplicated and the capacity is never exceeded. This is verified by the
  concurrency tests under `-race`.
