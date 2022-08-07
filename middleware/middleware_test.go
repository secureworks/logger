package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/secureworks/logger/internal/testutils"
	"github.com/secureworks/logger/log"
	"github.com/secureworks/logger/middleware"
	"github.com/secureworks/logger/testlogger"
	_ "github.com/secureworks/logger/testlogger"
)

func TestNewHTTPServer(t *testing.T) {
	var c io.Closer
	srv := httptest.NewUnstartedServer(nil)
	logger, _ := log.Open("test", nil)
	srv.Config, c = middleware.NewHTTPServer(logger, log.INFO)
	defer c.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.LoggerFromCtx(r.Context())
		testutils.AssertNotNil(t, logger)
	})
	srv.Start()

	resp, err := srv.Client().Get(srv.URL)
	testutils.AssertNil(t, err)
	defer resp.Body.Close()

	testutils.AssertEqual(t, http.StatusOK, resp.StatusCode)
}

func TestHTTPRequestMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	resp, logger := runMiddlewareAround(t, req, nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry := log.EntryFromCtx(r.Context())
		entry.WithStr("Meta", "data").Msg("message here")
		w.WriteHeader(http.StatusOK)
	}))
	testutils.AssertEqual(t, http.StatusOK, resp.Code)
	testutils.AssertEqual(t, 1, len(logger.Entries))
	entry := logger.Entries[0]

	testutils.AssertTrue(t, entry.IsAsync)
	testutils.AssertTrue(t, entry.Sent)
	testutils.AssertEqual(t, log.INFO, entry.Level)
	testutils.AssertEqual(t, "message here", entry.Message)
	testutils.AssertEqual(t, "data", entry.StringField("Meta"))

	testutils.AssertTrue(t, entry.RequestDuration() > time.Duration(0))
	testutils.AssertEqual(t, req.Method, entry.RequestMethod())
	testutils.AssertEqual(t, req.URL.Path, entry.RequestPath())
	testutils.AssertEqual(t, req.RemoteAddr, entry.RequestRemoteAddr())
}

func TestHTTPRequestLogAttributes(t *testing.T) {
	uID := "uuid-uuid-uuid-uuid"
	rID := "my-pod-name-1234567-aaaaa:" + uID
	tID := "trace_it"
	sID := "span_it"
	cID := "5000"
	env := "pilot"

	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	req.Header.Set("X-Request-Id", rID)
	req.Header.Set("X-Trace-Id", tID)
	req.Header.Set("X-Span-Id", sID)
	req.Header.Set("X-Tenant-Ctx", cID)
	req.Header.Set("X-Environment", env)

	resp, logger := runMiddlewareAround(t, req,
		func(r *http.Request, w middleware.ResponseWriter, e log.Entry, tm time.Time) {
			middleware.DefaultLogFunc(r, w, e, tm) // Default headers.
			middleware.AddHeader(e, "x-trace-id", r.Header)
			middleware.AddHeader(e, "x-span-id", r.Header)
			middleware.AddHeader(e, "x-tenant-ctx", r.Header)
			middleware.AddHeader(e, "x-environment", r.Header)
			middleware.AddHeader(e, "x-other", r.Header) // Should fail.

			if reqID := r.Header.Get("x-request-id"); strings.Contains(reqID, ":") {
				e.WithStr("req.uuid", strings.Split(reqID, ":")[1])
			} // Should not fail.
			if reqID := r.Header.Get("x-request-id"); strings.Contains(reqID, "|") {
				e.WithStr("req.other", strings.Split(reqID, "|")[1])
			} // Should fail.
		},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"status":"OK"}`))
		}),
	)
	testutils.AssertEqual(t, http.StatusOK, resp.Code)
	testutils.AssertEqual(t, 1, len(logger.Entries))
	entry := logger.Entries[0]

	testutils.AssertEqual(t, rID, entry.StringField("x-request-id"))
	testutils.AssertEqual(t, tID, entry.StringField("x-trace-id"))
	testutils.AssertEqual(t, sID, entry.StringField("x-span-id"))
	testutils.AssertEqual(t, cID, entry.StringField("x-tenant-ctx"))
	testutils.AssertEqual(t, env, entry.StringField("x-environment"))
	testutils.AssertEqual(t, uID, entry.StringField("req.uuid"))
	testutils.AssertEqual(t, "200", entry.StringField("http.status_code"))
	testutils.AssertEqual(t, "15B", entry.StringField("http.body_size"))

	testutils.AssertFalse(t, entry.HasField("x-other"))
	testutils.AssertFalse(t, entry.HasField("req.other"))
	testutils.AssertFalse(t, entry.HasField("res.other"))
}

func TestHTTPRequestMiddlewarePanic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	res, logger := runMiddlewareAround(t, req, nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("this is fine")
	}))
	testutils.AssertEqual(t, http.StatusInternalServerError, res.Code)
	testutils.AssertEqual(t, 1, len(logger.Entries))
	entry := logger.Entries[0]

	testutils.AssertTrue(t, entry.IsAsync)
	testutils.AssertTrue(t, entry.Sent)
	testutils.AssertEqual(t, log.ERROR, entry.Level)

	testutils.AssertTrue(t, entry.RequestDuration() > time.Duration(0))
	testutils.AssertEqual(t, req.Method, entry.RequestMethod())
	testutils.AssertEqual(t, req.URL.Path, entry.RequestPath())
	testutils.AssertEqual(t, req.RemoteAddr, entry.RequestRemoteAddr())

	pv, ok := entry.Fields[log.PanicValue].(string)
	testutils.AssertTrue(t, ok)
	testutils.AssertEqual(t, "this is fine", pv)

	st, ok := entry.Fields[log.PanicStack].(errors.StackTrace)
	testutils.AssertTrue(t, ok)
	testutils.AssertTrue(t, len(st) > 0)
}

// runMiddlewareAround wraps a default logging middleware setup around
// the given handler, executes the given request against it and returns
// the ResponseRecorder and the test logger involved.
func runMiddlewareAround(
	t *testing.T,
	req *http.Request,
	logFn middleware.LogFunc,
	handler http.Handler,
) (*httptest.ResponseRecorder, *testlogger.Logger) {
	t.Helper()

	logger, _ := testlogger.New(log.DefaultConfig(nil))
	resp := httptest.NewRecorder()
	h := middleware.NewHTTPRequestMiddleware(
		logger,
		log.INFO,
		logFn,
	)(handler)
	h.ServeHTTP(resp, req)

	return resp, logger.(*testlogger.Logger)
}
