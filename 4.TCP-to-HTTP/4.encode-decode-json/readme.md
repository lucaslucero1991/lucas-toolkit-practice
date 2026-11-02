# JSON, Headers, Status Codes and Basic Validation

This is the last foundational layer before building a production-style API. We now focus on how to properly interact with the client: JSON encoding/decoding, headers, status codes, and simple validation.

## 1. Setting Response Headers

When returning JSON we must tell the client what content type to expect:

```go
w.Header().Set("Content-Type", "application/json")
```

This is a fundamental part of the HTTP contract. Clients rely on headers to know how to interpret the response body.

## 2. Encoding JSON Responses

The standard library `encoding/json` package works directly with `io.Writer`:

```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)

if err := json.NewEncoder(w).Encode(users); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
```

Key points:

- `json.NewEncoder` accepts any `io.Writer` (and `http.ResponseWriter` implements it)
- This is another example of Go’s powerful interface composition
- Always handle the error from `Encode`

## 3. Decoding JSON Request Bodies

For incoming JSON we use a decoder:

```go
var user User
if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
}
```

- `400 Bad Request` is the appropriate status when the client sends malformed JSON
- `500 Internal Server Error` is reserved for problems on the server side

## 4. Struct Tags for JSON

Go uses struct tags to control JSON field names:

```go
type User struct {
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
}
```

Because exported fields start with an uppercase letter, tags are necessary to produce the conventional lowercase/snake_case JSON that clients expect.

## 5. Returning Proper Status Codes

The `net/http` package provides constants for all standard status codes:

- `http.StatusOK` (200)
- `http.StatusCreated` (201)
- `http.StatusBadRequest` (400)
- `http.StatusInternalServerError` (500)
- etc.

Use `http.Error` for simple error responses:

```go
http.Error(w, "user already exists", http.StatusBadRequest)
```

Or set the status manually when you need more control:

```go
w.WriteHeader(http.StatusCreated)
```

## 6. Simple Validation Without External Packages

You can perform basic validation directly in the handler or in a small helper function:

```go
func insertUser(users []User, u User) error {
    if u.FirstName == "" {
        return errors.New("first name is required")
    }
    if u.LastName == "" {
        return errors.New("last name is required")
    }

    for _, existing := range users {
        if existing.FirstName == u.FirstName && existing.LastName == u.LastName {
            return errors.New("user already exists")
        }
    }

    // append and return nil
    return nil
}
```

This keeps the example free of external dependencies while still showing the important idea: validate early and return clear errors to the client.

## Key Takeaways

- Always set the `Content-Type` header when returning JSON.
- `json.NewEncoder` / `json.NewDecoder` work with `io.Writer` / `io.Reader` — a core Go pattern.
- Use struct tags to control the JSON shape.
- Choose status codes carefully: 400 for client mistakes, 500 for server problems, 201 for successful creation.
- Simple validation can be done without third-party libraries; keep it clear and explicit.
- These building blocks (headers, JSON, status codes, validation) are used constantly in real APIs.
