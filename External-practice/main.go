package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

const (
	Address      = "localhost:8080"
	MaxQueueSize = 3
	MaxBodyBytes = 1 << 20
	QueuePath    = "queue.log"
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

const (
	opPush = "push"
	opPop  = "pop"
)

type record struct {
	Op   string `json:"op"`
	Body string `json:"body,omitempty"`
}

type fileBuffer struct {
	mu      sync.Mutex
	items   []Message
	maxSize int
	file    *os.File
	size    int64
}

func (fb *fileBuffer) Push(msg Message) error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if len(fb.items) >= fb.maxSize {
		return ErrFull
	}

	if err := fb.writeRecord(record{Op: opPush, Body: msg.Body}); err != nil {
		return err
	}

	fb.items = append(fb.items, msg)
	return nil
}

func (fb *fileBuffer) Pop() (Message, error) {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	if len(fb.items) == 0 {
		return Message{}, ErrEmpty
	}

	if err := fb.writeRecord(record{Op: opPop}); err != nil {
		return Message{}, err
	}

	msg := fb.items[0]
	fb.items = fb.items[1:]
	return msg, nil
}

func (fb *fileBuffer) Close() error {
	fb.mu.Lock()
	defer fb.mu.Unlock()

	return fb.file.Close()
}

func (fb *fileBuffer) writeRecord(rec record) error {
	line, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	line = append(line, '\n')

	if _, err := fb.file.Write(line); err != nil {
		fb.rollback()
		return err
	}

	if err := fb.file.Sync(); err != nil {
		fb.rollback()
		return err
	}

	fb.size += int64(len(line))
	return nil
}

func (fb *fileBuffer) rollback() {
	if err := fb.file.Truncate(fb.size); err != nil {
		slog.Error("failed to roll back partial queue log write", "error", err)
	}
}

func replayLog(path string, maxSize int) ([]Message, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var items []Message
	decoder := json.NewDecoder(file)
	for {
		var rec record
		if err := decoder.Decode(&rec); err != nil {
			var syntaxErr *json.SyntaxError
			switch {
			case errors.Is(err, io.EOF):
				return items, nil
			case errors.Is(err, io.ErrUnexpectedEOF), errors.As(err, &syntaxErr):
				slog.Warn("ignoring corrupt tail of queue log", "error", err)
				return items, nil
			default:
				return nil, err
			}
		}

		switch rec.Op {
		case opPush:
			if len(items) >= maxSize {
				return nil, ErrFull
			}
			items = append(items, Message{Body: rec.Body})
		case opPop:
			if len(items) > 0 {
				items = items[1:]
			}
		default:
			return nil, fmt.Errorf("unknown record op %q", rec.Op)
		}
	}
}

func compactLog(path string, items []Message) error {
	tmpPath := path + ".tmp"

	tmp, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(tmp)
	for _, item := range items {
		if err := encoder.Encode(record{Op: opPush, Body: item.Body}); err != nil {
			tmp.Close()
			return err
		}
	}

	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	dir, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer dir.Close()

	return dir.Sync()
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
		slog.Error("some error happens:", "error", err.Error())
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

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("starting server", "address", Address)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
	}

	if err := buff.Close(); err != nil {
		slog.Error("failed to close buffer", "error", err)
		os.Exit(1)
	}

	slog.Info("buffer closed, exiting")
}

func OpenMessageBuffer(maxSize int) (*fileBuffer, error) {
	if maxSize <= 0 {
		return nil, fmt.Errorf("maxSize must be greater than zero, got %d", maxSize)
	}

	return NewFileBuffer(maxSize, QueuePath)
}

func NewMemoryBuffer(maxSize int) *memoryBuffer {
	return &memoryBuffer{
		messages: make(chan Message, maxSize),
	}
}

func NewFileBuffer(maxSize int, path string) (*fileBuffer, error) {
	items, err := replayLog(path, maxSize)
	if err != nil {
		return nil, err
	}

	if err := compactLog(path, items); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	return &fileBuffer{
		items:   items,
		maxSize: maxSize,
		file:    file,
		size:    info.Size(),
	}, nil
}

func (a *api) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /message", a.handlePost)
	mux.HandleFunc("GET /message", a.handleGet)
	return mux
}
