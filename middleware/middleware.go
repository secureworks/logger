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

	var ioc io.Closer = noopCloser{}
	if srvLvl.IsValid() {
		wc := logger.WriteCloser(srvLvl)
		ioc = wc

		srv.ErrorLog = stdlog.New(wc, "[HTTP SERVER] ", stdlog.LstdFlags)
	}

	return srv, ioc
}

// HTTPRequestMiddlewareEntries allows us to inject two different ways
// to automatically log aspects of a request: headers and synthetics.
// Headers is a list of header names that will be set as fields in the
// log if they are present in the request; synthetics are fields
// generated from some combination or process applied to the request.
type HTTPRequestMiddlewareEntries struct {
	Headers    []string
	Synthetics map[string]func(*http.Request) string
}

// NewHTTPRequestMiddleware returns net/http compatible middleware for
// logging requests that pass through it at the provided level. It will
// also insert an Async log.Entry into the request context such that
// downstream handlers can use it. It will call entry.Send when done,
// and capture panics. If lvl is invalid, the default level will be
// used.
func NewHTTPRequestMiddleware(logger log.Logger, lvl log.Level, logEngtries *HTTPRequestMiddlewareEntries) func(http.Handler) http.Handler {
	if !lvl.IsValid() {
		lvl = log.Level(log.INFO)
	}

	logEntry := func(w http.ResponseWriter, r *http.Request, entry log.Entry, start time.Time) {
		entry.WithStr(log.ReqDuration, time.Since(start).String())
		path := r.RequestURI
		if path == "" {
			path = r.URL.Path
		}
		entry.WithStr(log.ReqMethod, r.Method)
		entry.WithStr(log.ReqPath, path)
		entry.WithStr(log.ReqRemoteAddr, r.RemoteAddr)
		if logEngtries != nil {
			for _, header := range logEngtries.Headers {
				addIfPresent(header, r, entry)
			}
			for header, valueFn := range logEngtries.Synthetics {
				addIfAvailable(header, valueFn(r), entry)
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

			defer logEntry(w, r, entry, time.Now())
			next.ServeHTTP(w, r)
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

type noopCloser struct{}

func (noopCloser) Close() error { return nil }
