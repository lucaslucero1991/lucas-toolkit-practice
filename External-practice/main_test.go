package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
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
	a := &api{buff: NewMemoryBuffer(MaxQueueSize)}

	req := httptest.NewRequest(http.MethodPost, "/message", strings.NewReader("hello"))
	rec := httptest.NewRecorder()

	a.handlePost(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	msg, err := a.buff.Pop()
	require.NoError(t, err)
	assert.Equal(t, "hello", msg.Body)
}

func TestHandlePost_EmptyBody(t *testing.T) {
	a := &api{buff: NewMemoryBuffer(MaxQueueSize)}

	req := httptest.NewRequest(http.MethodPost, "/message", strings.NewReader(""))
	rec := httptest.NewRecorder()

	a.handlePost(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlePost_Full(t *testing.T) {
	buff := NewMemoryBuffer(1)
	require.NoError(t, buff.Push(Message{Body: "Busy"}))

	a := &api{buff: buff}

	req := httptest.NewRequest(http.MethodPost, "/message", strings.NewReader("overflow"))
	rec := httptest.NewRecorder()

	a.handlePost(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
}

func TestHandlePost_InternalError(t *testing.T) {
	a := &api{buff: &MessageBufferMock{
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

	a := &api{buff: buff}

	req := httptest.NewRequest(http.MethodGet, "/message", nil)
	rec := httptest.NewRecorder()

	a.handleGet(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	body, err := io.ReadAll(rec.Body)
	require.NoError(t, err)
	assert.Equal(t, "payload", string(body))
}

func TestHandleGet_NotFound(t *testing.T) {
	a := &api{buff: NewMemoryBuffer(MaxQueueSize)}

	req := httptest.NewRequest(http.MethodGet, "/message", nil)
	rec := httptest.NewRecorder()

	a.handleGet(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestHandleGet_InternalError(t *testing.T) {
	a := &api{buff: &MessageBufferMock{
		PopFunc: func() (Message, error) {
			return Message{}, fmt.Errorf("unexpected failure")
		},
	}}

	req := httptest.NewRequest(http.MethodGet, "/message", nil)
	rec := httptest.NewRecorder()

	a.handleGet(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestMemoryBuffer_ConcurrentPushNoLoss(t *testing.T) {
	const workers = 50
	buff := NewMemoryBuffer(workers)

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			_ = buff.Push(Message{Body: fmt.Sprintf("msg-%d", i)})
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

func TestMemoryBuffer_ConcurrentPushRespectsCapacity(t *testing.T) {
	const capacity = 10
	const workers = 100
	buff := NewMemoryBuffer(capacity)

	var wg sync.WaitGroup
	var accepted atomic.Int64
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			if err := buff.Push(Message{Body: fmt.Sprintf("msg-%d", i)}); err == nil {
				accepted.Add(1)
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, int64(capacity), accepted.Load())
}

func TestMemoryBuffer_ConcurrentPushPop(t *testing.T) {
	const capacity = 20
	const producers = 50
	buff := NewMemoryBuffer(capacity)

	var wg sync.WaitGroup
	var consumed atomic.Int64
	wg.Add(producers * 2)

	for i := 0; i < producers; i++ {
		go func(i int) {
			defer wg.Done()
			for buff.Push(Message{Body: fmt.Sprintf("msg-%d", i)}) != nil {
			}
		}(i)

		go func() {
			defer wg.Done()
			for {
				if _, err := buff.Pop(); err == nil {
					consumed.Add(1)
					return
				}
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, int64(producers), consumed.Load())

	_, err := buff.Pop()
	assert.ErrorIs(t, err, ErrEmpty)
}

func TestReadBody_EmptyReturnsSentinel(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/message", strings.NewReader(""))
	rec := httptest.NewRecorder()

	_, err := readBody(rec, req)

	assert.ErrorIs(t, err, ErrEmptyBody)
}

func TestReadBody_OK(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/message", strings.NewReader("payload"))
	rec := httptest.NewRecorder()

	body, err := readBody(rec, req)

	require.NoError(t, err)
	assert.Equal(t, "payload", body)
}

func TestRoutes_PostThenGet(t *testing.T) {
	a := &api{buff: NewMemoryBuffer(MaxQueueSize)}
	srv := httptest.NewServer(a.routes())
	defer srv.Close()

	postResp, err := http.Post(srv.URL+"/message", "text/plain", strings.NewReader("routed"))
	require.NoError(t, err)
	defer postResp.Body.Close()
	assert.Equal(t, http.StatusCreated, postResp.StatusCode)

	getResp, err := http.Get(srv.URL + "/message")
	require.NoError(t, err)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusOK, getResp.StatusCode)

	body, err := io.ReadAll(getResp.Body)
	require.NoError(t, err)
	assert.Equal(t, "routed", string(body))
}

func TestRoutes_MethodNotAllowed(t *testing.T) {
	a := &api{buff: NewMemoryBuffer(MaxQueueSize)}
	srv := httptest.NewServer(a.routes())
	defer srv.Close()

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/message", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestFileBuffer_PushPopFIFO(t *testing.T) {
	buff, err := NewFileBuffer(3, filepath.Join(t.TempDir(), "queue.json"))
	require.NoError(t, err)

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

func TestFileBuffer_PopEmpty(t *testing.T) {
	buff, err := NewFileBuffer(1, filepath.Join(t.TempDir(), "queue.json"))
	require.NoError(t, err)

	_, err = buff.Pop()
	assert.ErrorIs(t, err, ErrEmpty)
}

func TestFileBuffer_PushFull(t *testing.T) {
	buff, err := NewFileBuffer(2, filepath.Join(t.TempDir(), "queue.json"))
	require.NoError(t, err)

	require.NoError(t, buff.Push(Message{Body: "a"}))
	require.NoError(t, buff.Push(Message{Body: "b"}))

	assert.ErrorIs(t, buff.Push(Message{Body: "c"}), ErrFull)
}

func TestFileBuffer_PopFreesCapacity(t *testing.T) {
	buff, err := NewFileBuffer(1, filepath.Join(t.TempDir(), "queue.json"))
	require.NoError(t, err)

	require.NoError(t, buff.Push(Message{Body: "a"}))
	assert.ErrorIs(t, buff.Push(Message{Body: "b"}), ErrFull)

	_, err = buff.Pop()
	require.NoError(t, err)

	assert.NoError(t, buff.Push(Message{Body: "b"}))
}

func TestFileBuffer_SurvivesCrashWithoutClose(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.log")

	original, err := NewFileBuffer(3, path)
	require.NoError(t, err)

	require.NoError(t, original.Push(Message{Body: "first"}))
	require.NoError(t, original.Push(Message{Body: "second"}))

	reloaded, err := NewFileBuffer(3, path)
	require.NoError(t, err)

	first, err := reloaded.Pop()
	require.NoError(t, err)
	assert.Equal(t, "first", first.Body)

	second, err := reloaded.Pop()
	require.NoError(t, err)
	assert.Equal(t, "second", second.Body)

	_, err = reloaded.Pop()
	assert.ErrorIs(t, err, ErrEmpty)
}

func TestFileBuffer_PoppedMessagesDoNotReappearAfterCrash(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.log")

	original, err := NewFileBuffer(3, path)
	require.NoError(t, err)

	require.NoError(t, original.Push(Message{Body: "first"}))
	require.NoError(t, original.Push(Message{Body: "second"}))
	require.NoError(t, original.Push(Message{Body: "third"}))

	popped, err := original.Pop()
	require.NoError(t, err)
	assert.Equal(t, "first", popped.Body)

	reloaded, err := NewFileBuffer(3, path)
	require.NoError(t, err)

	second, err := reloaded.Pop()
	require.NoError(t, err)
	assert.Equal(t, "second", second.Body)

	third, err := reloaded.Pop()
	require.NoError(t, err)
	assert.Equal(t, "third", third.Body)

	_, err = reloaded.Pop()
	assert.ErrorIs(t, err, ErrEmpty)
}

func TestFileBuffer_NewMessagesAfterReloadSurviveSecondCrash(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.log")

	first, err := NewFileBuffer(3, path)
	require.NoError(t, err)
	require.NoError(t, first.Push(Message{Body: "beta"}))
	require.NoError(t, first.Push(Message{Body: "gamma"}))

	second, err := NewFileBuffer(3, path)
	require.NoError(t, err)
	_, err = second.Pop()
	require.NoError(t, err)
	_, err = second.Pop()
	require.NoError(t, err)
	require.NoError(t, second.Push(Message{Body: "lucas"}))
	require.NoError(t, second.Push(Message{Body: "lucas_2"}))

	third, err := NewFileBuffer(3, path)
	require.NoError(t, err)

	lucas, err := third.Pop()
	require.NoError(t, err)
	assert.Equal(t, "lucas", lucas.Body)

	lucas2, err := third.Pop()
	require.NoError(t, err)
	assert.Equal(t, "lucas_2", lucas2.Body)

	_, err = third.Pop()
	assert.ErrorIs(t, err, ErrEmpty)
}

func TestFileBuffer_RejectedPushIsNotPersisted(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.log")

	original, err := NewFileBuffer(1, path)
	require.NoError(t, err)

	require.NoError(t, original.Push(Message{Body: "kept"}))
	require.ErrorIs(t, original.Push(Message{Body: "rejected"}), ErrFull)

	reloaded, err := NewFileBuffer(1, path)
	require.NoError(t, err)

	kept, err := reloaded.Pop()
	require.NoError(t, err)
	assert.Equal(t, "kept", kept.Body)

	_, err = reloaded.Pop()
	assert.ErrorIs(t, err, ErrEmpty)
}

func TestFileBuffer_IgnoresTornTailRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.log")

	original, err := NewFileBuffer(3, path)
	require.NoError(t, err)
	require.NoError(t, original.Push(Message{Body: "complete"}))

	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	_, err = file.WriteString(`{"op":"push","bo`)
	require.NoError(t, err)
	require.NoError(t, file.Close())

	reloaded, err := NewFileBuffer(3, path)
	require.NoError(t, err)

	complete, err := reloaded.Pop()
	require.NoError(t, err)
	assert.Equal(t, "complete", complete.Body)

	_, err = reloaded.Pop()
	assert.ErrorIs(t, err, ErrEmpty)
}

func TestFileBuffer_CompactsLogOnStartup(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.log")

	original, err := NewFileBuffer(2, path)
	require.NoError(t, err)

	for i := 0; i < 20; i++ {
		require.NoError(t, original.Push(Message{Body: fmt.Sprintf("msg-%d", i)}))
		_, err := original.Pop()
		require.NoError(t, err)
	}
	require.NoError(t, original.Push(Message{Body: "survivor"}))
	require.NoError(t, original.Close())

	before, err := os.ReadFile(path)
	require.NoError(t, err)

	reloaded, err := NewFileBuffer(2, path)
	require.NoError(t, err)

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Less(t, len(after), len(before))

	survivor, err := reloaded.Pop()
	require.NoError(t, err)
	assert.Equal(t, "survivor", survivor.Body)
}

func TestFileBuffer_ReloadWithSmallerCapacityFails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "queue.log")

	original, err := NewFileBuffer(3, path)
	require.NoError(t, err)
	require.NoError(t, original.Push(Message{Body: "a"}))
	require.NoError(t, original.Push(Message{Body: "b"}))
	require.NoError(t, original.Close())

	_, err = NewFileBuffer(1, path)
	assert.ErrorIs(t, err, ErrFull)
}

func TestNewFileBuffer_MissingFileStartsEmpty(t *testing.T) {
	buff, err := NewFileBuffer(3, filepath.Join(t.TempDir(), "does-not-exist.log"))
	require.NoError(t, err)

	_, err = buff.Pop()
	assert.ErrorIs(t, err, ErrEmpty)
}

func TestFileBuffer_PushAfterCloseFails(t *testing.T) {
	buff, err := NewFileBuffer(3, filepath.Join(t.TempDir(), "queue.log"))
	require.NoError(t, err)
	require.NoError(t, buff.Close())

	err = buff.Push(Message{Body: "late"})
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrFull)
}

func TestFileBuffer_ConcurrentPushPopNoLoss(t *testing.T) {
	const capacity = 5
	const producers = 20
	path := filepath.Join(t.TempDir(), "queue.log")
	buff, err := NewFileBuffer(capacity, path)
	require.NoError(t, err)

	var wg sync.WaitGroup
	var consumed atomic.Int64
	wg.Add(producers * 2)

	for i := 0; i < producers; i++ {
		go func(i int) {
			defer wg.Done()
			for buff.Push(Message{Body: fmt.Sprintf("msg-%d", i)}) != nil {
			}
		}(i)

		go func() {
			defer wg.Done()
			for {
				if _, err := buff.Pop(); err == nil {
					consumed.Add(1)
					return
				}
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, int64(producers), consumed.Load())

	reloaded, err := NewFileBuffer(capacity, path)
	require.NoError(t, err)

	_, err = reloaded.Pop()
	assert.ErrorIs(t, err, ErrEmpty)
}
