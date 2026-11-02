package main

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
)

const (
	Address      = "localhost:8080"
	MaxQueueSize = 10
)

type Message struct {
	Body string `json:"body"`
}

// Sentinel errors the MessageBuffer methods may return.
var (
	ErrEmpty = errors.New("empty")
	ErrFull  = errors.New("full")
)

// MessageBuffer is a bounded FIFO queue of messages.
//
//go:generate moq -out moq_test.go . MessageBuffer
type MessageBuffer interface {
	Push(msg Message) error
	Pop() (Message, error)
}

type memoryBuffer struct {
	mu       sync.Mutex
	messages []Message
	maxSize  int
}

func (mb *memoryBuffer) Push(msg Message) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if len(mb.messages) >= mb.maxSize {
		return ErrFull
	}

	mb.messages = append(mb.messages, msg)

	return nil
}

func (mb *memoryBuffer) Pop() (Message, error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if len(mb.messages) == 0 {
		return Message{}, ErrEmpty
	}

	msg := mb.messages[0]
	mb.messages = mb.messages[1:]
	return msg, nil
}

// OpenMessageBuffer constructs a MessageBuffer with the given max size.
func OpenMessageBuffer(maxSize int) (MessageBuffer, error) {
	return &memoryBuffer{
		messages: make([]Message, 0, maxSize),
		maxSize:  maxSize,
	}, nil
	//panic("TODO: implement the MessageBuffer interface and return it here")
}

func handlePost(buff MessageBuffer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		err = buff.Push(Message{
			Body: string(body),
		})

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
}

func handleGet(buff MessageBuffer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		msg, err := buff.Pop()
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

	mux := http.NewServeMux()
	mux.HandleFunc("POST /message", handlePost(buff))
	mux.HandleFunc("GET /message", handleGet(buff))

	slog.Info("starting server", "address", Address)
	if err := http.ListenAndServe(Address, mux); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}

	// TODO: build an HTTP server that listens on Address and exposes:
	//
	//   GET  /message
	//     - 200 with the message body if a message was popped from the queue
	//     - 404 if the queue is empty
	//     - 5XX for any other error
	//
	//   POST /message
	//     - 201 if the message was pushed onto the queue
	//     - 429 if the queue is full
	//     - 5XX for any other error
	//
	// You may use anything in the standard library (net/http is recommended).
	// Third-party routers and frameworks are NOT permitted for this exercise.
	//panic("TODO: start the HTTP server")
}
