# TCP Foundations & the `net` Package in Go

Before building HTTP servers, we should understand the layer they sit on.

## 1. The Fundamental Problem

IP alone is not enough for reliable communication:

- Packets can be lost
- Packets can arrive out of order
- There is no guarantee of delivery

This is a problem for almost any application that needs complete and ordered data (web pages, APIs, etc.).

## 2. What TCP Provides

TCP sits on top of IP and gives us two main guarantees:

- **Reliable delivery** — every packet is acknowledged. Lost packets are retried.
- **Ordered delivery** — packets are sequenced so the receiver can reassemble them correctly.

TCP is also **connection-oriented**:

- One side acts as the server (listens)
- The other acts as the client (connects)
- Communication is bidirectional once the connection is established
- Connections are identified by IP + port

## 3. How a typical web request works

1. Browser does a DNS lookup to get the IP
2. TCP handshake establishes the connection
3. HTTP runs on top of that TCP connection
4. Response comes back and the page is rendered

HTTP is an application-layer protocol that relies on TCP as its transport.

## 4. TCP in Go (`net` package)

The `net` package gives us a portable interface for TCP.

Key concepts from the documentation:

- `net.Listen("tcp", address)` → returns a `Listener`
- `Listener.Accept()` → blocks and returns a `net.Conn` for each new connection
- Both `Listener` and `Conn` implement `Close()` — this is a strong signal that you should manage their lifecycle (usually with `defer`)

Important pattern:

```go
listener, err := net.Listen("tcp", ":8080")
// handle error
defer listener.Close()

for {
    conn, err := listener.Accept()
    // handle error
    go handleConnection(conn)   // critical: do not block the accept loop
}
```

Why the goroutine?

- `Accept()` is sequential.
- Without a goroutine, you can only handle one connection at a time.
- Spawning a goroutine per connection keeps the listener free to accept new clients.

## 5. Reader and Writer interfaces

`net.Conn` implements both `io.Reader` and `io.Writer`.

This is one of the most important patterns in Go:

- Anything that implements `io.Reader` can be passed to functions that expect a reader (e.g. `bufio.NewReader`)
- The same applies to `io.Writer`

This composability is a core part of Go’s standard library design.

## Key Takeaways

- TCP solves reliability and ordering problems that raw IP has.
- HTTP runs on top of TCP.
- In Go, `net.Listen` + `Accept` + goroutines is the classic pattern for a concurrent TCP server.
- Always close listeners and connections.
- `net.Conn` implementing `io.Reader` / `io.Writer` is a fundamental pattern you will see everywhere.
