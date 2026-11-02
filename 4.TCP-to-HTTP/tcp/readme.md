# HTTP Routing & the `net/http` Package

Before using the high-level `net/http` package, it is useful to understand the core idea behind routing.

## 1. The Fundamental Idea of Routing

A request arrives at the server and must be directed to the correct handler.

At a low level this means:

- Accepting a TCP connection
- Reading the incoming data
- Parsing the request (method + path)
- Deciding which logic should handle it
- Writing a response back to the client

This is the same conceptual model used by HTTP servers, even if the real implementation is more sophisticated.

## 2. A Minimal Routing Example

A simple mental model looks like this:

1. Client opens a TCP connection and sends a request line (e.g. `GET /index.html`)
2. Server reads the data (usually until a newline)
3. Server splits the line into parts (method and path)
4. Server looks at the method and path to decide what to do
5. Server writes a response back to the connection

Even a very basic implementation that only handles `GET` already demonstrates the core pattern of routing.

## 3. How `net/http` Builds on the Same Ideas

The standard library package `net/http` is built on top of the `net` package and TCP.

When you call:

```go
http.ListenAndServe(addr, handler)
```

under the hood it:

- Listens on a TCP network
- Accepts connections
- Serves them using the same fundamental pattern we saw at the lower level

Looking at the source of `net/http` reveals familiar concepts:

- Path cleaning and normalization
- Host and port stripping
- A multiplexer (`ServeMux`) that matches requests to handlers
- Internal switches and redirects for routing decisions

The production code is far more robust, but the underlying mental model remains the same as the simple example.

## 4. Why This Matters

Understanding the low-level flow helps you:

- See that `net/http` is not magic — it is a well-engineered abstraction over TCP and request parsing
- Recognize common patterns (listeners, connections, readers/writers, path matching)
- Feel more confident when debugging or extending HTTP servers

## Key Takeaways

- Routing is fundamentally about receiving a request and dispatching it to the right handler.
- A minimal TCP-based example already shows the core ideas of parsing and dispatching.
- `net/http.ListenAndServe` uses TCP under the hood and follows the same conceptual model.
- Looking at the standard library source reveals the same patterns (path cleaning, multiplexing, handlers) at a more mature level.
- Understanding these foundations makes the higher-level `net/http` package much clearer.
