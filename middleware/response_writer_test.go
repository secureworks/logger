package middleware_test

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/http"
	"testing"

	"github.com/secureworks/logger/log/testutils"
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

type mockHijackerOnly struct{}

func (w *mockHijackerOnly) Write(_ []byte) (int, error) { return 0, nil }

func (w *mockHijackerOnly) WriteHeader(_ int) {}

func (w *mockHijackerOnly) Header() http.Header { return http.Header{} }

func (w *mockHijackerOnly) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}

type mockHTTP11RW struct {
	CalledFlush  bool
	CalledHijack bool
}

func (w *mockHTTP11RW) Write(_ []byte) (int, error) { return 0, nil }

func (w *mockHTTP11RW) WriteHeader(_ int) {}

func (w *mockHTTP11RW) Header() http.Header { return http.Header{} }

func (w *mockHTTP11RW) Flush() { w.CalledFlush = true }

func (w *mockHTTP11RW) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.CalledHijack = true
	return nil, nil, nil
}

type mockHTTP2RW struct {
	CalledFlush bool
	CalledPush  bool
}

func (w *mockHTTP2RW) Write(_ []byte) (int, error) { return 0, nil }

func (w *mockHTTP2RW) WriteHeader(_ int) {}

func (w *mockHTTP2RW) Header() http.Header { return http.Header{} }

func (w *mockHTTP2RW) Flush() { w.CalledFlush = true }

func (w *mockHTTP2RW) Push(_ string, _ *http.PushOptions) error {
	w.CalledPush = true
	return nil
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
		_, _ = w.Write([]byte(`{"example":true}`))
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
		_, _ = w.Write([]byte(`{"example":true}`))
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
		_, _ = w.Write([]byte(`{"example":`))
		_, _ = w.Write([]byte(`true}`))
		testutils.AssertEqual(t, 16, w.BodySize())
	})
}

func TestResponseWriter_Flush(t *testing.T) {
	t.Run("if underlying response writer does not implement alt interface; does not implement", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockRW{Buffer: new(bytes.Buffer)})
		_, ok := w.(http.Flusher)
		testutils.AssertFalse(t, ok)
	})

	t.Run("is a no-op when underlying alt response writer does not implement", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockHijackerOnly{})
		flusher, ok := w.(http.Flusher)
		testutils.AssertTrue(t, ok)
		assertNotPanics(t, func() { flusher.Flush() })
	})

	t.Run("when implemented, passes through to underlying response writer", func(t *testing.T) {
		mock1 := &mockHTTP11RW{}
		w := middleware.NewResponseWriter(mock1)
		flusher, ok := w.(http.Flusher)
		testutils.AssertTrue(t, ok)
		flusher.Flush()
		testutils.AssertTrue(t, mock1.CalledFlush)

		mock2 := &mockHTTP2RW{}
		w = middleware.NewResponseWriter(mock2)
		flusher, ok = w.(http.Flusher)
		testutils.AssertTrue(t, ok)
		flusher.Flush()
		testutils.AssertTrue(t, mock2.CalledFlush)
	})
}

func TestResponseWriter_Hijack(t *testing.T) {
	t.Run("if underlying response writer does not implement hijacker; does not implement", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockRW{Buffer: new(bytes.Buffer)})
		_, ok := w.(http.Hijacker)
		testutils.AssertFalse(t, ok)
	})

	t.Run("when implemented, passes through to underlying response writer", func(t *testing.T) {
		mock := &mockHTTP11RW{}
		w := middleware.NewResponseWriter(mock)
		hijacker, _ := w.(http.Hijacker)
		_, _, _ = hijacker.Hijack() // TODO(PH): IDK, maybe should check these?
		testutils.AssertTrue(t, mock.CalledHijack)
	})
}

func TestResponseWriter_Pusher(t *testing.T) {
	t.Run("if underlying response writer does not implement pusher; does not implement", func(t *testing.T) {
		w := middleware.NewResponseWriter(&mockRW{Buffer: new(bytes.Buffer)})
		_, ok := w.(http.Pusher)
		testutils.AssertFalse(t, ok)
	})

	t.Run("when implemented, passes through to underlying response writer", func(t *testing.T) {
		mock := &mockHTTP2RW{}
		w := middleware.NewResponseWriter(mock)
		pusher, _ := w.(http.Pusher)
		_ = pusher.Push("", nil) // TODO(PH): IDK, maybe should check these?
		testutils.AssertTrue(t, mock.CalledPush)
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
