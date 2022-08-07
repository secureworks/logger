// Package middleware has HTTP server middleware that provides access to
// a shared logger and simplifies writing canonical log lines with a
// shared entry per request. It also automates basic request parameter
// logging.
//
// NewHTTPServer binds a logger to an HTTP server so that it will be
// present in every request context:
//
//     // Create a logger.
//     logger, err := log.Open("zerolog", nil)
//
//     // Generate a new HTTP server with the logger.
//     srv, logCloser = middleware.NewHTTPServer(logger, log.INFO)
//     defer logCloser.Close()
//
//     // Now every request will have access to the logger in the context.
//     srv.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//         logger := log.LoggerFromCtx(r.Context())
//     })
//     srv.ListenAndServe()
//
// NewHTTPRequestMiddleware injects an entry into the context of the the
// request that it writes after the succeeding handlers process it. This
// entry has certain aspects of the request logged by default, but can
// be configured to include more. Subsequent handlers can also extract
// the entry and update it to allow for canonical log line logging.
//
//     // Create a logger.
//     logger, err := log.Open("zerolog", nil)
//
//     // Define request attributes to log.
//     attrs := middleware.HTTPRequestLogAttributes{
//         Headers: []string{"X-Trace-Id", "X-Request-Id"},
//         Synthetics: map[string]func(*http.Request) string{
//             "req.unmod-uri": func(r *http.Request) string { return r.RequestURI },
//         },
//     }
//
//     // Generate the middleware and then wrap subsequent handlers.
//     middlewareFn := middleware.NewHTTPRequestMiddleware(logger, log.INFO, attrs)
//     handler := middlewareFn(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//         w.WriteHeader(http.StatusOK)
//     }))
//     handler.ServeHTTP(resp, req)
//
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

// LogFunc ... FIXME
type LogFunc func(*http.Request, ResponseWriter, log.Entry, time.Time)

// DefaultLogFunc ... FIXME
func DefaultLogFunc(r *http.Request, w ResponseWriter, e log.Entry, start time.Time) {
	if value := r.Header.Get("x-request-id"); value != "" {
		e.WithStr(log.ReqID, value)
	}
	e.WithStr(log.ReqMethod, r.Method)
	path := r.RequestURI
	if path == "" {
		path = r.URL.Path
	}
	e.WithStr(log.ReqPath, path)
	e.WithInt(log.ResStatusCode, w.StatusCode())
	e.WithStr(log.ReqRemoteAddr, r.RemoteAddr)
	e.WithStr(log.ReqDuration, time.Since(start).String())
	e.WithInt("http.body_size", w.BodySize())
}

// NewHTTPRequestMiddleware returns net/http compatible middleware for
// logging requests that pass through it at the provided level. It will
// also insert an Async log.Entry into the request context such that
// downstream handlers can use it. It will call entry.Send when done,
// and capture panics. If lvl is invalid, the default level will be
// used.
func NewHTTPRequestMiddleware(logger log.Logger, lvl log.Level, logFn LogFunc) func(http.Handler) http.
	Handler {
	if !lvl.IsValid() {
		lvl = log.INFO
	}
	if logFn == nil {
		logFn = DefaultLogFunc
	}

	logEntry := func(w ResponseWriter, r *http.Request, entry log.Entry, start time.Time) {
		logFn(r, w, entry, start)
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

func AddHeader(e log.Entry, name string, h http.Header) {
	if value := h.Get(name); value != "" {
		e.WithStr(strings.ToLower(name), value)
	}
}
