package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

// ResponseWriter implements an interface that allows logging middleware
// access to standard information about the underlying response that is
// unattainable in the default http.ResponseWriter interface.
//
// In order not to break basic net/http interfaces that are commonly
// implemented (http.Flusher, http.Hijacker http.Pusher), we handle
// these as well, passing to the underlying response writer if it
// implements them in turn.
//
// One issue is that the available implementations will always fulfill
// http.Flusher, even when only http.Hijacker or http.Pusher is
// implemented by the underlying response writer. While this mirrors the
// most common implementations (the standard HTTP/1.1 and HTTP/2
// response writers, eg), it may lead to false positive http.Flusher
// type assertions. Since an unimplemented call to Flush is a no-op,
// this can be regarded as a minor issue.
//
// Does not hold a separate response body buffer. Log in the application
// for this sort of data.
type ResponseWriter interface {
	http.ResponseWriter

	// StatusCode returns the status code of the response. If not written
	// yet, this returns 0.
	StatusCode() int

	// Status returns the HTTP/1.1-standard name for the status code. If
	// not written yet, this returns "".
	Status() string

	// BodySize returns the size of the response body.
	BodySize() int
}

// NewResponseWriter returns a logging-specific ResponseWriter that
// wraps an http.ResponseWriter, for use in logging middleware.
func NewResponseWriter(w http.ResponseWriter) ResponseWriter {
	var rw ResponseWriter = &responseWriter{ResponseWriter: w}

	f, isFlusher := w.(http.Flusher)

	if h, ok := w.(http.Hijacker); ok {
		hrw := &hijackerResponseWriter{ResponseWriter: rw, Hijacker: h}
		if isFlusher {
			hrw.Flusher = f
		}
		return hrw
	}
	if p, ok := w.(http.Pusher); ok {
		prw := &pusherResponseWriter{ResponseWriter: rw, Pusher: p}
		if isFlusher {
			prw.Flusher = f
		}
		return prw
	}
	if isFlusher {
		return &flusherResponseWriter{ResponseWriter: rw, Flusher: f}
	}
	return rw
}

// responseWriter implements the ResponseWriter interface.
type responseWriter struct {
	http.ResponseWriter

	statusCode int
	status     string
	bodySize   int
}

var _ http.ResponseWriter = (*responseWriter)(nil)

// Getters.

func (w *responseWriter) StatusCode() int {
	return w.statusCode
}

func (w *responseWriter) Status() string {
	return w.status
}

func (w *responseWriter) BodySize() int {
	return w.bodySize
}

// HasBeenWrittenTo returns whether the ResponseWriter has been written to.
func (w *responseWriter) HasBeenWrittenTo() bool {
	return w.statusCode != 0
}

// Implement (wrap) http.ResponseWriter.

func (w *responseWriter) WriteHeader(code int) {
	if w.HasBeenWrittenTo() {
		return
	}
	w.statusCode = code
	w.status = fmt.Sprintf("%d %s", code, http.StatusText(code))
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.HasBeenWrittenTo() {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.bodySize += n
	return n, err
}

// Implement alternate interfaces.

type flusherResponseWriter struct {
	ResponseWriter
	http.Flusher
}

func (w *flusherResponseWriter) Flush() {
	if wc, ok := w.ResponseWriter.(interface{ HasBeenWrittenTo() bool }); ok && !wc.HasBeenWrittenTo() {
		w.WriteHeader(http.StatusOK)
	}
	w.Flusher.Flush()
}

type hijackerResponseWriter struct {
	ResponseWriter
	http.Hijacker
	http.Flusher
}

func (w *hijackerResponseWriter) Flush() {
	if w.Flusher == nil {
		return
	}
	if wc, ok := w.ResponseWriter.(interface{ HasBeenWrittenTo() bool }); ok && !wc.HasBeenWrittenTo() {
		w.WriteHeader(http.StatusOK)
	}
	w.Flusher.Flush()
}

func (w *hijackerResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.Hijacker.Hijack()
}

type pusherResponseWriter struct {
	ResponseWriter
	http.Pusher
	http.Flusher
}

func (w *pusherResponseWriter) Flush() {
	if w.Flusher == nil {
		return
	}
	if wc, ok := w.ResponseWriter.(interface{ HasBeenWrittenTo() bool }); ok && !wc.HasBeenWrittenTo() {
		w.WriteHeader(http.StatusOK)
	}
	w.Flusher.Flush()
}

func (w *pusherResponseWriter) Push(target string, opts *http.PushOptions) error {
	return w.Pusher.Push(target, opts)
}
