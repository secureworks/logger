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

	"github.com/secureworks/logger/log"
	"github.com/secureworks/logger/log/internal/common"
)

type noopClose struct{}

func (noopClose) Close() error { return nil }

// NewHTTPServer returns an http.Server with its BaseContext set to
// logger as its value. If srvLvl is valid then the http.Server's
// ErrorLog field will also be set, in which case the returned io.Closer
// should be closed when finished.
func NewHTTPServer(logger log.Logger, srvLvl log.Level) (*http.Server, io.Closer) {
	ctx := log.CtxWithLogger(context.Background(), logger)

	srv := &http.Server{
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	var ioc io.Closer = noopClose{}
	if srvLvl.IsValid() {
		wc := logger.WriteCloser(srvLvl)
		ioc = wc

		srv.ErrorLog = stdlog.New(wc, "[HTTP SERVER] ", stdlog.LstdFlags)
	}

	return srv, ioc
}

// NewHTTPRequestMiddleware returns net/http compatible middleware for
// logging requests that pass through it at the provided level. It will
// also insert an Async log.Entry into the request context such that
// downstream handlers can use it. It will call entry.Send when done,
// and capture panics. If lvl is invalid, the default level will be
// used.
func NewHTTPRequestMiddleware(logger log.Logger, lvl log.Level) func(http.Handler) http.Handler {
	if !lvl.IsValid() {
		lvl = log.Level(0)
	}

	addIfPresent := func(k string, r *http.Request, e log.Entry) {
		if v := r.Header.Get(k); v != "" {
			e.WithStr(strings.ToLower(k), v)
		}

		// Mechanism for checking context.
	}

	// Retrieves the app name on from the x-request-id header making the
	// following assumptions:
	//   - The format for this header's value is: {pod-or-app-name}:{uuid}
	//   - If pod name is present it is in the following form
	//     {appName}-{hash}-{hash}
	//   - If empty is passed, "" is returned, if the value doesn't
	//     contain ":" "" is returned
	getSourceApp := func(requestID string) string {
		reqParts := strings.Split(requestID, ":")
		if len(reqParts) == 1 {
			// Not colon delimited, don't want it.
			return ""
		}
		podNameDelim := "-"
		if !strings.Contains(reqParts[0], podNameDelim) {
			// Short app name, still want it.
			return reqParts[0]
		}
		parts := strings.Split(reqParts[0], podNameDelim)
		// Dash delimited app name (might be a pod name, want it).
		return strings.Join(parts[0:len(parts)-2], podNameDelim)
	}

	logEntry := func(w http.ResponseWriter, r *http.Request, entry log.Entry, start time.Time) {
		entry.WithStr(log.ReqDuration, time.Since(start).String())
		path := r.RequestURI
		if path == "" {
			path = r.URL.Path
		}
		entry.WithStr("http_path", path)
		entry.WithStr("http_method", r.Method)
		entry.WithStr("http_remote_addr", r.RemoteAddr)
		addIfPresent(log.XRequestID, r, entry)
		addIfPresent(log.XTraceID, r, entry)
		addIfPresent(log.XSpanID, r, entry)
		addIfPresent(log.XTenantCtx, r, entry)
		addIfPresent(log.XEnvironment, r, entry)

		if requestID := r.Header.Get(log.XRequestID); requestID != "" {
			if srcApp := getSourceApp(requestID); srcApp != "" {
				entry.WithStr("src.app", srcApp)
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
