# Working with `net/http`: Handlers and Routing

Once we move beyond raw TCP, we start consuming the `net/http` package. This is the foundation we will use throughout the course.

## 1. The Core Interface: `http.Handler`

The most important interface in `net/http` is:

```go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

Any type that implements this method can be used as an HTTP handler.

This interface appears everywhere:
- In the standard library
- In third-party routers
- In almost every Go web package

Understanding and implementing `http.Handler` is essential.

## 2. Booting a Server

The simplest way to start an HTTP server is:

```go
http.ListenAndServe(":8080", handler)
```

- The first argument is the address
- The second argument is anything that implements `http.Handler`

Under the hood this still uses TCP, but the package handles the connection management for us.

## 3. ResponseWriter and Request

Two key types you will use constantly:

- `http.ResponseWriter` — used to write the response (headers + body). It implements `io.Writer`.
- `*http.Request` — contains all information about the incoming request (method, URL, headers, body, etc.)

These two types abstract away most of the HTTP protocol details so we can focus on application logic.

## 4. Minimal Routing with Switch

A basic way to do routing is by switching on the request method and path:

```go
func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        switch r.URL.Path {
        case "/":
            // index
        case "/users":
            // users
        default:
            http.NotFound(w, r)
        }
    }
}
```

This works, but it quickly becomes messy as the number of routes grows.

Prefer using the constants from the `http` package (`http.MethodGet`, etc.) instead of hard-coded strings.

## 5. Moving to `http.ServeMux`

A cleaner and more idiomatic approach is to use `http.ServeMux` (the standard library multiplexer / router):

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users", getUsersHandler)
mux.HandleFunc("GET /", indexHandler)

server := &http.Server{
    Addr:    ":8080",
    Handler: mux,
}
server.ListenAndServe()
```

Key points:

- `ServeMux` itself implements `http.Handler`
- `HandleFunc` registers a path (and optionally a method, in Go 1.22+) to a handler function
- This keeps routing declarative and separated from business logic
- Most third-party routers (chi, gorilla/mux, etc.) build additional features on top of the same ideas

## 6. Using `http.Server` for Configuration

Instead of calling `http.ListenAndServe` directly, it is common to create an `http.Server` value:

```go
server := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  5 * time.Second,
    WriteTimeout: 10 * time.Second,
}
```

This gives you control over timeouts, TLS configuration, and other production concerns.

## Key Takeaways

- `http.Handler` is the central interface of Go web programming.
- `ResponseWriter` and `Request` abstract the HTTP protocol for us.
- Simple switch-based routing works for tiny examples but does not scale well.
- `http.ServeMux` is the standard, clean way to register routes.
- Prefer `http.Server` when you need timeouts or other configuration.
- Third-party routers add convenience on top of the same foundational concepts.
