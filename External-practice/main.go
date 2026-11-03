package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	Address      = "localhost:8080"
	MaxQueueSize = 10
	MaxBodyBytes = 1 << 20
)

type Message struct {
	Body string `json:"body"`
}

var (
	ErrEmpty     = errors.New("empty")
	ErrFull      = errors.New("full")
	ErrEmptyBody = errors.New("empty body")
)

//go:generate moq -out moq_test.go . MessageBuffer
type MessageBuffer interface {
	Push(msg Message) error
	Pop() (Message, error)
}

type memoryBuffer struct {
	messages chan Message
}

func (mb *memoryBuffer) Push(msg Message) error {
	select {
	case mb.messages <- msg:
		return nil
	default:
		return ErrFull
	}
}

func (mb *memoryBuffer) Pop() (Message, error) {
	select {
	case msg := <-mb.messages:
		return msg, nil
	default:
		return Message{}, ErrEmpty
	}
}

type api struct {
	buff MessageBuffer
}

func (a *api) handlePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := readBody(w, r)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = a.buff.Push(Message{Body: body})
	if err != nil {
		if errors.Is(err, ErrFull) {
			http.Error(w, err.Error(), http.StatusTooManyRequests)
			return
		}

		slog.Error("failed to push message", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func readBody(w http.ResponseWriter, r *http.Request) (string, error) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	if len(body) == 0 {
		return "", ErrEmptyBody
	}

	return string(body), nil
}

func (a *api) handleGet(w http.ResponseWriter, r *http.Request) {
	msg, err := a.buff.Pop()
	if err != nil {
		if errors.Is(err, ErrEmpty) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		slog.Error("failed to pop message", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(msg.Body))
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	buff, err := OpenMessageBuffer(MaxQueueSize)
	if err != nil {
		slog.Error("failed to open message buffer", "error", err)
		os.Exit(1)
	}

	a := &api{buff: buff}

	srv := &http.Server{
		Addr:         Address,
		Handler:      a.routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	slog.Info("starting server", "address", Address)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func OpenMessageBuffer(maxSize int) (MessageBuffer, error) {
	if maxSize <= 0 {
		return nil, fmt.Errorf("maxSize must be greater than zero, got %d", maxSize)
	}

	return NewMemoryBuffer(maxSize), nil
}

func NewMemoryBuffer(maxSize int) *memoryBuffer {
	return &memoryBuffer{
		messages: make(chan Message, maxSize),
	}
}

func (a *api) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /message", a.handlePost)
	mux.HandleFunc("GET /message", a.handleGet)
	return mux
}
