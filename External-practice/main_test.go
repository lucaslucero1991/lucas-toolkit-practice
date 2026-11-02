package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryBuffer_PushPopFIFO(t *testing.T) {
	buff := NewMemoryBuffer(3)

	require.NoError(t, buff.Push(Message{Body: "first"}))
	require.NoError(t, buff.Push(Message{Body: "second"}))
	require.NoError(t, buff.Push(Message{Body: "third"}))

	first, err := buff.Pop()
	require.NoError(t, err)
	assert.Equal(t, "first", first.Body)

	second, err := buff.Pop()
	require.NoError(t, err)
	assert.Equal(t, "second", second.Body)

	third, err := buff.Pop()
	require.NoError(t, err)
	assert.Equal(t, "third", third.Body)
}

func TestMemoryBuffer_PopEmpty(t *testing.T) {
	buff := NewMemoryBuffer(1)

	_, err := buff.Pop()
	assert.ErrorIs(t, err, ErrEmpty)
}

func TestMemoryBuffer_PushFull(t *testing.T) {
	buff := NewMemoryBuffer(2)

	require.NoError(t, buff.Push(Message{Body: "a"}))
	require.NoError(t, buff.Push(Message{Body: "b"}))

	err := buff.Push(Message{Body: "c"})
	assert.ErrorIs(t, err, ErrFull)
}

func TestMemoryBuffer_PopFreesCapacity(t *testing.T) {
	buff := NewMemoryBuffer(1)

	require.NoError(t, buff.Push(Message{Body: "a"}))
	assert.ErrorIs(t, buff.Push(Message{Body: "b"}), ErrFull)

	_, err := buff.Pop()
	require.NoError(t, err)

	assert.NoError(t, buff.Push(Message{Body: "b"}))
}

func TestHandlePost_Created(t *testing.T) {
	a := &api{addr: Address, buff: NewMemoryBuffer(MaxQueueSize)}

	req := httptest.NewRequest(http.MethodPost, "/message", strings.NewReader("hello"))
	rec := httptest.NewRecorder()

	a.handlePost(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	msg, err := a.buff.Pop()
	require.NoError(t, err)
	assert.Equal(t, "hello", msg.Body)
}

func TestHandlePost_EmptyBody(t *testing.T) {
	a := &api{addr: Address, buff: NewMemoryBuffer(MaxQueueSize)}

	req := httptest.NewRequest(http.MethodPost, "/message", strings.NewReader(""))
	rec := httptest.NewRecorder()

	a.handlePost(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlePost_Full(t *testing.T) {
	buff := NewMemoryBuffer(1)
	require.NoError(t, buff.Push(Message{Body: "Busy"}))

	a := &api{addr: Address, buff: buff}

	req := httptest.NewRequest(http.MethodPost, "/message", strings.NewReader("overflow"))
	rec := httptest.NewRecorder()

	a.handlePost(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestHandlePost_InternalError(t *testing.T) {
	a := &api{addr: Address, buff: &MessageBufferMock{
		PushFunc: func(msg Message) error {
			return fmt.Errorf("unexpected failure")
		},
	}}

	req := httptest.NewRequest(http.MethodPost, "/message", strings.NewReader("hello"))
	rec := httptest.NewRecorder()

	a.handlePost(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandleGet_OK(t *testing.T) {
	buff := NewMemoryBuffer(MaxQueueSize)
	require.NoError(t, buff.Push(Message{Body: "payload"}))

	a := &api{addr: Address, buff: buff}

	req := httptest.NewRequest(http.MethodGet, "/message", nil)
	rec := httptest.NewRecorder()

	a.handleGet(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)
	assert.Equal(t, "payload", string(body))
}

func TestHandleGet_NotFound(t *testing.T) {
	a := &api{addr: Address, buff: NewMemoryBuffer(MaxQueueSize)}

	req := httptest.NewRequest(http.MethodGet, "/message", nil)
	rec := httptest.NewRecorder()

	a.handleGet(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleGet_InternalError(t *testing.T) {
	a := &api{addr: Address, buff: &MessageBufferMock{
		PopFunc: func() (Message, error) {
			return Message{}, fmt.Errorf("unexpected failure")
		},
	}}

	req := httptest.NewRequest(http.MethodGet, "/message", nil)
	rec := httptest.NewRecorder()

	a.handleGet(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestMemoryBuffer_ConcurrentPushWithinCapacity(t *testing.T) {
	const workers = 50
	buff := NewMemoryBuffer(workers)

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(n int) {
			defer wg.Done()
			_ = buff.Push(Message{Body: fmt.Sprintf("msg-%d", n)})
		}(i)
	}

	wg.Wait()

	drained := make(map[string]bool)
	for {
		msg, err := buff.Pop()
		if err != nil {
			break
		}
		drained[msg.Body] = true
	}

	assert.Len(t, drained, workers)
	for i := 0; i < workers; i++ {
		assert.True(t, drained[fmt.Sprintf("msg-%d", i)], "missing message msg-%d", i)
	}
}

func TestMemoryBuffer_ConcurrentPushRespectsLimit(t *testing.T) {
	const capacity = 10
	const workers = 100
	buff := NewMemoryBuffer(capacity)

	var wg sync.WaitGroup
	var accepted int64
	var mu sync.Mutex
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(n int) {
			defer wg.Done()
			if err := buff.Push(Message{Body: fmt.Sprintf("msg-%d", n)}); err == nil {
				mu.Lock()
				accepted++
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int64(capacity), accepted)
}

func TestMemoryBuffer_ConcurrentPushPop(t *testing.T) {
	const capacity = 20
	const producers = 50
	buff := NewMemoryBuffer(capacity)

	var producerWg sync.WaitGroup
	producerWg.Add(producers)

	var consumerWg sync.WaitGroup
	consumerWg.Add(producers)

	var consumed int64
	var mu sync.Mutex

	for i := 0; i < producers; i++ {
		go func(n int) {
			defer producerWg.Done()
			for {
				if err := buff.Push(Message{Body: fmt.Sprintf("msg-%d", n)}); err == nil {
					return
				}
			}
		}(i)

		go func() {
			defer consumerWg.Done()
			for {
				if _, err := buff.Pop(); err == nil {
					mu.Lock()
					consumed++
					mu.Unlock()
					return
				}
			}
		}()
	}

	producerWg.Wait()
	consumerWg.Wait()

	assert.Equal(t, int64(producers), consumed)

	_, err := buff.Pop()
	assert.ErrorIs(t, err, ErrEmpty)
}
