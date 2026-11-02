# From TCP to an HTTP Server in Go

A first-principles walkthrough of what actually happens beneath Go's `net/http` package. The goal is not to memorize APIs, but to understand the machinery we rely on every day so we can reason about failures, performance, and design decisions.

## The Fundamental Problem

Before we can serve HTTP, we need a way to move bytes reliably between two machines.

- Machines talk over **IP (Internet Protocol)**, which moves data in packets.
- Each packet carries a header (source, destination) plus a chunk of the payload.
- IP alone is not enough for applications because it gives **no guarantees**:
  - Packets may be lost (you send a message, it may never arrive).
  - Packets may arrive out of order.

For something like loading a web page, both problems are fatal: missing or reordered data means a broken response.

## What TCP Adds

TCP runs on top of IP and turns an unreliable packet service into a reliable stream.

- **Reliable delivery**
  - Every packet must be acknowledged by the receiver.
  - Unacknowledged packets are retried until they arrive.
- **Ordered delivery**
  - Packets carry sequence information.
  - The receiver reassembles them in the correct order, waiting for missing packets and reordering late ones.

Mental model: IP gives you "best-effort postcards"; TCP gives you an "ordered, acknowledged conversation."

## TCP as a Connection

TCP is a **connection-oriented** protocol, which shapes how we design servers and clients.

- One side is the **server**, the other is the **client**.
- The channel is **bidirectional** — both sides can send once the connection exists.
- A connection must be **established** before data flows (a three-step handshake; the details matter less than the fact that it is explicit).
- Endpoints are identified by **IP + port**. Every machine exposes services on specific ports.

## Where HTTP Fits

HTTP is an application protocol that assumes a reliable transport underneath.

Example: visiting `google.com`:

- **DNS lookup** resolves the hostname to an IP address.
- A **TCP connection (handshake)** is established to that server.
- **HTTP** takes over on top of the TCP transport layer (request/response).
- The browser renders the response into a page.

Key takeaway: `net/http` is a convenience layer over `net`, which is itself a portable interface over TCP/IP.

## Building a TCP Server in Go

Go's standard library exposes this stack directly through the `net` package. `net/http` depends on `net`.

The minimal shape of a TCP server:

```go
listener, err := net.Listen("tcp", ":8080")
if err != nil {
    // handle error
}
defer listener.Close()

for {
    conn, err := listener.Accept()
    if err != nil {
        // handle error
        continue
    }
    go handleConnection(conn)
}
```

What each piece tells us:

- `net.Listen("tcp", addr)` returns a **`Listener`**.
  - The `Listener` interface exposes a `Close` method — a strong Go convention signaling a long-lived resource you are responsible for closing. Hence `defer listener.Close()`.
  - `Accept()` blocks until a client connects, then returns a connection.
- Each `Accept()` returns a **`net.Conn`**.
  - `Conn` exposes `Read`, `Write`, and `Close`.
  - Again, the presence of `Close` is the hint: close the connection when done.

## The Goroutine Is Not Optional

The `go handleConnection(conn)` line is the critical design decision, not a detail.

- `Accept()` runs inside the `for` loop; handling a connection inline would **block the loop**.
- Blocking means only **one connection is processed at a time** — effectively a serial server.
- Spawning a goroutine per connection lets the accept loop immediately return to waiting for the next client.
- This is the idiomatic Go concurrency model: cheap goroutines handle independent work, the loop stays responsive.

## Why `Read` and `Write` Matter: the io Interfaces

The real power of `net.Conn` is that it satisfies Go's core I/O interfaces.

- `Conn` implements `io.Reader` and `io.Writer`.
- Anything that implements `io.Reader` can be passed anywhere an `io.Reader` is expected.

Example — wrapping the connection to read it line by line:

```go
func handleConnection(conn net.Conn) {
    defer conn.Close()

    reader := bufio.NewReader(conn) // bufio.NewReader takes an io.Reader
    line, err := reader.ReadString('\n')
    // parse `line` (e.g. a command from the client)
}
```

The mental model worth internalizing:

- `bufio.NewReader` does not know or care that it was handed a TCP connection — it only knows `io.Reader`.
- A connection, a file, an in-memory buffer, and an HTTP body are **interchangeable** at this boundary.
- This composability is why Go's I/O ecosystem is so uniform: small interfaces (`Reader`, `Writer`) compose into everything else.

## Resource Ownership

A recurring theme ties the whole example together.

- Every long-lived resource (`Listener`, `Conn`) exposes `Close`.
- Treat `Close` as a contract: pair each acquisition with a `defer ...Close()`.
- Forgetting this leaks file descriptors and connections — a common source of production failures.

## Key Takeaways

- IP moves packets but guarantees nothing; **TCP adds reliable, ordered delivery** on top of it.
- TCP is **connection-oriented and bidirectional**, with endpoints identified by **IP + port**.
- HTTP is an application protocol layered over a TCP transport; `net/http` sits on top of `net`.
- A Go TCP server is a loop: **`Listen` → `Accept` → handle**, with `defer Close()` on every resource.
- **One goroutine per connection** keeps the accept loop non-blocking and enables concurrency.
- `net.Conn` satisfies **`io.Reader`/`io.Writer`**, making connections composable with the entire standard-library I/O ecosystem.
- The `Close` method is a convention: it signals ownership of a long-lived resource you must release.
