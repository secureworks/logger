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
	"github.com/stretchr/testify/require"
)

func TestNewHTTPServer(t *testing.T) {
	require := require.New(t)

	var c io.Closer
	srv := httptest.NewUnstartedServer(nil)
	logger, _ := testlogger.New(nil)
	srv.Config, c = middleware.NewHTTPServer(logger, log.INFO)
	defer c.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry := log.EntryFromCtx(r.Context())
		require.NotNil(entry)
	})
	srv.Start()

	resp, err := srv.Client().Get(srv.URL)
	require.NoError(err)
	defer resp.Body.Close()

	require.Equal(http.StatusOK, resp.StatusCode)
}

func TestHTTPRequestMiddleware(t *testing.T) {
	require := require.New(t)

	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	resp, logger := testutils.RunMiddlewareAround(t, req, nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry := log.EntryFromCtx(r.Context())
		entry.WithStr("Meta", "data").Msg("message here")
		w.WriteHeader(http.StatusOK)
	}))
	require.Equal(http.StatusOK, resp.Code)
	require.Len(logger.Entries, 1)
	entry := logger.Entries[0]

	require.True(entry.IsAsync)
	require.True(entry.Sent)
	require.Equal(log.INFO, entry.Level)
	require.Equal("message here", entry.Message)
	require.Equal([]string{"data"}, entry.Field("Meta"))

	require.Greater(entry.RequestDuration(), time.Duration(0))
	require.Equal(req.Method, entry.RequestMethod())
	require.Equal(req.URL.Path, entry.RequestPath())
	require.Equal(req.RemoteAddr, entry.RequestRemoteAddr())
}

func TestHTTPRequestMiddlewareEntries(t *testing.T) {
	uID := "uuid-uuid-uuid-uuid"
	rID := "my-pod-name-1234567-aaaaa:" + uID
	tID := "trace_it"
	sID := "span_it"
	cID := "5000"
	env := "pilot"

	require := require.New(t)

	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	req.Header.Set("X-Request-Id", rID)
	req.Header.Set("X-Trace-Id", tID)
	req.Header.Set("X-Span-Id", sID)
	req.Header.Set("X-Tenant-Ctx", cID)
	req.Header.Set("X-Environment", env)

	resp, logger := testutils.RunMiddlewareAround(t, req,
		&middleware.HTTPRequestMiddlewareEntries{
			Headers: []string{
				"X-Request-Id",
				"X-Trace-Id",
				"X-Span-Id",
				"X-Tenant-Ctx",
				"X-Environment",
				"X-Other", // Should fail.
			},
			Synthetics: map[string]func(*http.Request) string{
				"req.uuid": func(r *http.Request) (val string) {
					if reqID := r.Header.Get("X-Request-Id"); strings.Contains(reqID, ":") {
						val = strings.Split(reqID, ":")[1]
					}
					return
				},
				// Should fail.
				"req.other": func(r *http.Request) (val string) {
					if reqID := r.Header.Get("X-Request-Id"); strings.Contains(reqID, "|") {
						val = strings.Split(reqID, "|")[1]
					}
					return
				},
			},
		},
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	require.Equal(http.StatusOK, resp.Code)
	require.Len(logger.Entries, 1)
	entry := logger.Entries[0]

	require.Equal(rID, entry.StringField("x-request-id"))
	require.Equal(tID, entry.StringField("x-trace-id"))
	require.Equal(sID, entry.StringField("x-span-id"))
	require.Equal(cID, entry.StringField("x-tenant-ctx"))
	require.Equal(env, entry.StringField("x-environment"))
	require.Equal(uID, entry.StringField("req.uuid"))

	require.False(entry.HasField("x-other"))
	require.False(entry.HasField("req.other"))
}

func TestHTTPRequestMiddlewarePanic(t *testing.T) {
	require := require.New(t)

	req := httptest.NewRequest(http.MethodGet, "/test/path", nil)
	res, logger := testutils.RunMiddlewareAround(t, req, nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("this is fine")
	}))
	require.Equal(http.StatusInternalServerError, res.Code)
	require.Len(logger.Entries, 1)
	entry := logger.Entries[0]

	require.True(entry.IsAsync)
	require.True(entry.Sent)
	require.Equal(log.ERROR, entry.Level)

	require.Greater(entry.RequestDuration(), time.Duration(0))
	require.Equal(req.Method, entry.RequestMethod())
	require.Equal(req.URL.Path, entry.RequestPath())
	require.Equal(req.RemoteAddr, entry.RequestRemoteAddr())

	pv, ok := entry.Fields[log.PanicValue].(string)
	require.True(ok)
	require.Equal("this is fine", pv)

	st, ok := entry.Fields[log.PanicStack].(errors.StackTrace)
	require.True(ok)
	require.Greater(len(st), 0)
}
