package middleware

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
)

// ResponseWriter implements an interface that allows logging middleware
// access to standard information about the underlying response that is
// unattainable in the default http.ResponseWriter interface.
//
// In order not to break basic net/http interfaces that are commonly
// implemented (http.Flusher and http.Hijacker), we handle these as
// well, passing to the underlying response writer if it implements them
// in turn. This writer does not support http.Pusher.
//
// One issue is that our implementation fulfills both http.Flusher and
// http.Hijacker regardless of whether the underlying response writer
// does. Often, calling code will assert against an interface to check
// if an http.ResponseWriter may also implement these: this can lead to
// false positives when using this library.
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
	return &responseWriter{ResponseWriter: w}
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

// Implement http.Flusher and http.Hijacker.

func (w *responseWriter) Flush() {
	flusher, ok := w.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}
	if !w.HasBeenWrittenTo() {
		w.WriteHeader(http.StatusOK)
	}
	flusher.Flush()
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := w.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("the underlying ResponseWriter does not implement http.Hijacker")
	}
	return hijacker.Hijack()
}
