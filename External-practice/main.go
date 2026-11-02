package main

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	Address      = "localhost:8080"
	MaxQueueSize = 10
)

type Message struct {
	Body string `json:"body"`
}

var (
	ErrEmpty = errors.New("empty")
	ErrFull  = errors.New("full")
)

//go:generate moq -out moq_test.go . MessageBuffer
type MessageBuffer interface {
	Push(msg Message) error
	Pop() (Message, error)
}

type memoryBuffer struct {
	messages chan Message
}

func NewMemoryBuffer(maxSize int) *memoryBuffer {
	return &memoryBuffer{
		messages: make(chan Message, maxSize),
	}
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

func OpenMessageBuffer(maxSize int) (MessageBuffer, error) {
	return NewMemoryBuffer(maxSize), nil
}

type api struct {
	addr string
	buff MessageBuffer
}

func (a *api) handlePost(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(body) == 0 {
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	err = a.buff.Push(Message{Body: string(body)})
	if err != nil {
		if errors.Is(err, ErrFull) {
			http.Error(w, err.Error(), http.StatusTooManyRequests)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *api) handleGet(w http.ResponseWriter, r *http.Request) {
	msg, err := a.buff.Pop()
	if err != nil {
		if errors.Is(err, ErrEmpty) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	a := &api{addr: Address, buff: buff}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /message", a.handlePost)
	mux.HandleFunc("GET /message", a.handleGet)

	srv := &http.Server{
		Addr:         a.addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	slog.Info("starting server", "address", Address)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
