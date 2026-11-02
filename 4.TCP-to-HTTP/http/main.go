package main

import (
	"log"
	"net/http"
)

// Deep dive way
type server struct {
	addr string
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		switch r.URL.Path {
		case "/":
			w.Write([]byte("index page"))
			return
		case "/users":
			w.Write([]byte("users page"))
			return
		default:
			w.Write([]byte("404 page"))
			return
		}
	}
}

// More easy way
type api struct {
	addr string
}

func (a *api) getUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Getting users"))
}

func (a *api) createUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Created user"))
}

func main() {

	// easy way
	api := &api{addr: ":8080"}

	// Init router
	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    api.addr,
		Handler: mux,
	}

	mux.HandleFunc("GET /users", api.getUserHandler)
	mux.HandleFunc("POST /users", api.createUserHandler)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

	// Harder way
	/*s := &server{
		addr: ":8080",
	}
	if err := http.ListenAndServe(s.addr, s); err != nil {
		log.Fatal(err)
	}*/
}
