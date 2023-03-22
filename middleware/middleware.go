// Package middleware has HTTP server middleware that provides access to
// a shared logger and simplifies writing canonical log lines with a
// shared entry per request. It also automates basic request parameter
// logging.
//
// NewHTTPServer binds a logger to an HTTP server so that it will be
// present in every request context:
//
//	// Create a logger.
//	logger, err := log.Open("zerolog", nil)
//
//	// Generate a new HTTP server with the logger.
//	srv, logCloser = middleware.NewHTTPServer(logger, log.INFO)
//	defer logCloser.Close()
//
//	// Now every request will have access to the logger in the context.
//	srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    logger := log.LoggerFromContext(r.Context())
//	})
//	srv.ListenAndServe()
//
// NewHTTPRequestMiddleware injects an entry into the context of the the
// request that it writes after the succeeding handlers process it. This
// entry has certain aspects of the request logged by default, but can
// be configured to include more. Subsequent handlers can also extract
// the entry and update it to allow for canonical log line logging.
//
//	// Create a logger.
//	logger, err := log.Open("zerolog", nil)
//
//	// Define request attributes to log.
//	attrs := middleware.HTTPRequestLogAttributes{
//	    Headers: []string{"X-Trace-Id", "X-Request-Id"},
//	    Synthetics: map[string]func(*http.Request) string{
//	        "req.unmod-uri": func(r *http.Request) string { return r.RequestURI },
//	    },
//	}
//
//	// Generate the middleware and then wrap subsequent handlers.
//	middlewareFn := middleware.NewHTTPRequestMiddleware(logger, log.INFO, attrs)
//	handler := middlewareFn(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    w.WriteHeader(http.StatusOK)
//	}))
//	handler.ServeHTTP(resp, req)
package middleware

import (
	"context"
	"fmt"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/secureworks/logger/internal/common"
	"github.com/secureworks/logger/log"
)

// NewHTTPServer returns an http.Server with its BaseContext set to
// logger as its value. If srvLvl is valid then the http.Server's
// ErrorLog field will also be set, in which case the returned io.Closer
// should be closed when finished.
func NewHTTPServer(logger log.Logger, srvLvl log.Level) (*http.Server, io.Closer) {
	ctx := log.CtxWithLogger(context.Background(), logger)

	srv := &http.Server{
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	var cl io.Closer
	if srvLvl.IsValid() {
		wc := logger.WriteCloser(srvLvl)
		cl = io.Closer(wc)
		srv.ErrorLog = stdlog.New(wc, "[HTTP SERVER] ", stdlog.LstdFlags)
	}

	return srv, cl
}

// HTTPRequestLogAttributes allows us to inject two different ways
// to automatically log aspects of a request: headers and synthetics.
// Headers is a list of header names that will be set as fields in the
// log if they are present in the request; synthetics are fields
// generated from some combination or process applied to the request.
//
// If desired, the default attributes may also be skipped.
type HTTPRequestLogAttributes struct {
	Headers            []string
	Synthetics         map[string]func(*http.Request) string
	SyntheticsResponse map[string]func(ResponseWriter) string
	SkipDuration       bool
	SkipMethod         bool
	SkipPath           bool
	SkipRemoteAddr     bool
}

// NewHTTPRequestMiddleware returns net/http compatible middleware for
// logging requests that pass through it at the provided level. It will
// also insert an Async log.Entry into the request context such that
// downstream handlers can use it. It will call entry.Send when done,
// and capture panics. If lvl is invalid, the default level will be
// used.
func NewHTTPRequestMiddleware(logger log.Logger, lvl log.Level, attrs *HTTPRequestLogAttributes) func(http.Handler) http.Handler {
	if !lvl.IsValid() {
		lvl = log.INFO
	}

	logEntry := func(w ResponseWriter, r *http.Request, entry log.Entry, start time.Time) {
		if attrs == nil || attrs != nil && !attrs.SkipMethod {
			entry.WithStr(log.ReqMethod, r.Method)
		}
		if attrs == nil || attrs != nil && !attrs.SkipPath {
			path := r.RequestURI
			if path == "" {
				path = r.URL.Path
			}
			entry.WithStr(log.ReqPath, path)
		}
		if attrs == nil || attrs != nil && !attrs.SkipRemoteAddr {
			entry.WithStr(log.ReqRemoteAddr, r.RemoteAddr)
		}
		if attrs == nil || attrs != nil && !attrs.SkipDuration {
			entry.WithStr(log.ReqDuration, time.Since(start).String())
		}
		if attrs != nil {
			for _, header := range attrs.Headers {
				addIfPresent(header, r, entry)
			}
			for header, valueFn := range attrs.Synthetics {
				addIfAvailable(header, valueFn(r), entry)
			}
			for header, valueFn := range attrs.SyntheticsResponse {
				addIfAvailable(header, valueFn(w), entry)
			}
		}

		if pv := recover(); pv != nil {
			pve, ok := pv.(error)
			if !ok {
				pve = fmt.Errorf("%v", pv)
			}

			st, _ := common.WithStackTrace(pve)

			entry.Error().WithFields(map[string]interface{}{
				// Try to keep PanicValue field consistent as a string.
				log.PanicValue: fmt.Sprintf("%v", pv),
				log.PanicStack: st.StackTrace(),
			})

			w.WriteHeader(http.StatusInternalServerError)
		}

		entry.Send()
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			entry := logger.Entry(lvl).Async()

			ctx := log.CtxWithEntry(r.Context(), entry)
			r = r.WithContext(ctx)

			// Wrap the response writer with the logger version.
			w2 := NewResponseWriter(w)

			defer logEntry(w2, r, entry, time.Now())
			next.ServeHTTP(w2, r)
		})
	}
}

func addIfPresent(name string, r *http.Request, e log.Entry) {
	if value := r.Header.Get(name); value != "" {
		e.WithStr(strings.ToLower(name), value)
	}
}

func addIfAvailable(name string, value string, e log.Entry) {
	if name != "" && value != "" {
		e.WithStr(strings.ToLower(name), value)
	}
}
