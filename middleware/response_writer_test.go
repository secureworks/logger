package middleware_test

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/http"
	"testing"

	"github.com/secureworks/logger/internal/testutils"
	"github.com/secureworks/logger/middleware"
)

type mockRW struct {
	StatusCode int
	Buffer     *bytes.Buffer
	Err        error
	Hdr        http.Header
}

func (w *mockRW) Write(byt []byte) (int, error) {
	if w.Err != nil {
		return 1, w.Err
	}
	return w.Buffer.Write(byt)
}
func (w *mockRW) WriteHeader(statusCode int) { w.StatusCode = statusCode }

func (w *mockRW) Header() http.Header { return w.Hdr }

type mockHijackerFlusher struct {
	CalledFlush  bool
	CalledHijack bool
}

func (w *mockHijackerFlusher) Write(_ []byte) (int, error) { return 0, nil }

func (w *mockHijackerFlusher) WriteHeader(_ int) {}

func (w *mockHijackerFlusher) Header() http.Header { return http.Header{} }

func (w *mockHijackerFlusher) Flush() { w.CalledFlush = true }

func (w *mockHijackerFlusher) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.CalledHijack = true
	return nil, nil, nil
}

func TestResponseWriter(t *testing.T) {
	t.Run("passes Write through to underlying response writer", func(t *testing.T) {
		mock := &mockRW{Buffer: new(bytes.Buffer)}
		w := middleware.NewResponseWriter(mock)

		var n int
		var err error

		n, err = w.Write([]byte(`{"example":`))
		testutils.AssertNil(t, err)
		testutils.AssertEqual(t, 11, n)
		n, err = w.Write([]byte(`true}`))
		testutils.AssertNil(t, err)
		testutils.AssertEqual(t, 5, n)

		testutils.AssertEqual(t, `{"example":true}`, mock.Buffer.String())
		testutils.AssertEqual(t, 200, mock.StatusCode)
	})

	t.Run("passes WriteHeader through to underlying response writer", func(t *testing.T) {
		mock := &mockRW{Buffer: new(bytes.Buffer)}
		w := middleware.NewResponseWriter(mock)

		w.WriteHeader(501)
		testutils.AssertEqual(t, 501, mock.StatusCode)
	})

	t.Run("returns underlying writer state", func(t *testing.T) {
		mock := &mockRW{Err: errors.New("io issue"), Hdr: http.Header{"X-Example": []string{"true"}}}
		w := middleware.NewResponseWriter(mock)

		_, err := w.Write([]byte(`!!!`))
		testutils.AssertNotNil(t, err)

		testutils.AssertEqual(t, w.Header().Get("X-Example"), "true")
	})
}

func TestResponseWriter_StatusCode(t *testing.T) {
	t.Run("empty by default", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockRW{Buffer: new(bytes.Buffer)})
		testutils.AssertEqual(t, 0, w.StatusCode())
	})

	t.Run("written too when writing header", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockRW{Buffer: new(bytes.Buffer)})
		w.WriteHeader(404)
		testutils.AssertEqual(t, 404, w.StatusCode())
	})

	t.Run("written too when writing", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockRW{Buffer: new(bytes.Buffer)})
		w.Write([]byte(`{"example":true}`))
		testutils.AssertEqual(t, 200, w.StatusCode())
	})
}

func TestResponseWriter_Status(t *testing.T) {
	t.Run("empty by default", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockRW{Buffer: new(bytes.Buffer)})
		testutils.AssertEqual(t, "", w.Status())
	})

	t.Run("written too when writing header", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockRW{Buffer: new(bytes.Buffer)})
		w.WriteHeader(404)
		testutils.AssertEqual(t, "404 Not Found", w.Status())
	})

	t.Run("written too when writing", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockRW{Buffer: new(bytes.Buffer)})
		w.Write([]byte(`{"example":true}`))
		testutils.AssertEqual(t, "200 OK", w.Status())
	})
}

func TestResponseWriter_BodySize(t *testing.T) {
	t.Run("empty by default", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockRW{Buffer: new(bytes.Buffer)})
		testutils.AssertEqual(t, 0, w.BodySize())
	})

	t.Run("written too when writing", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockRW{Buffer: new(bytes.Buffer)})
		w.Write([]byte(`{"example":`))
		w.Write([]byte(`true}`))
		testutils.AssertEqual(t, 16, w.BodySize())
	})
}

func TestResponseWriter_Flush(t *testing.T) {
	t.Run("is a no-op when underlying response writer does not implement", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockRW{Buffer: new(bytes.Buffer)})
		flusher, _ := w.(http.Flusher)
		assertNotPanics(t, func() { flusher.Flush() })
	})

	t.Run("when implemented, passes through to underlying response writer", func(t *testing.T) {
		mock := &mockHijackerFlusher{}
		w := middleware.NewResponseWriter(mock)
		flusher, _ := w.(http.Flusher)
		flusher.Flush()
		testutils.AssertTrue(t, mock.CalledFlush)
	})
}

func TestResponseWriter_Hijack(t *testing.T) {
	t.Run("returns an error when underlying response writer does not implement", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockRW{Buffer: new(bytes.Buffer)})
		hijacker, _ := w.(http.Hijacker)
		assertNotPanics(t, func() {
			conn, rw, err := hijacker.Hijack()
			testutils.AssertNil(t, conn)
			testutils.AssertNil(t, rw)
			testutils.AssertEqual(t, err.Error(), "the underlying ResponseWriter does not implement http.Hijacker")
		})
	})

	t.Run("when implemented, passes through to underlying response writer", func(t *testing.T) {
		mock := &mockHijackerFlusher{}
		w := middleware.NewResponseWriter(mock)
		hijacker, _ := w.(http.Hijacker)
		_, _, _ = hijacker.Hijack() // TODO(PH): IDK, maybe should check these?
		testutils.AssertTrue(t, mock.CalledHijack)
	})
}

func assertNotPanics(t *testing.T, fn func()) {
	t.Helper()

	didPanic := true
	defer func() {
		if didPanic {
			t.Errorf("did panic")
		}
	}()

	fn()
	didPanic = false
}
