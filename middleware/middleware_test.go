package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/secureworks/logger/log"
)

func TestHTTPBaseContext(t *testing.T) {
	srv := httptest.NewUnstartedServer(nil)
	noop := log.Noop()

	var c io.Closer
	srv.Config, c = NewHTTPServer(noop, 0)
	defer c.Close()

	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if log.LoggerFromCtx(r.Context()) == nil {
			t.Fatal("Nil logger in request scoped context")
		}
	})
	srv.Start()

	resp, err := srv.Client().Get(srv.URL)
	if err != nil {
		t.Fatalf("Failed to make http request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Non-200 status from http.Server: %d", resp.StatusCode)
	}
}

type testRequest struct {
	meta         string
	data         string
	msg          string
	method       string
	path         string
	code         int
	xRequestID   string
	xTraceID     string
	xSpanID      string
	xTenantCtx   string
	xEnvironment string
}

func TestHTTPRequestMiddleware(t *testing.T) {
	method := http.MethodGet
	path := "/foobar"

	tr := &testRequest{
		meta:   "meta",
		data:   "data",
		msg:    "hello world",
		method: method,
		path:   path,
		code:   http.StatusCreated,
	}

	req := httptest.NewRequest(method, path, nil)
	expectedEntity := &map[string]interface{}{
		tr.meta:            tr.data,
		"http_method":      tr.method,
		"http_path":        tr.path,
		"http_remote_addr": req.RemoteAddr,
	}
	positiveTest(req, t, tr, *expectedEntity)
}

func TestHTTPRequestMiddlewareDetails(t *testing.T) {
	method := http.MethodGet
	path := "/foobar"
	requestID := "my-pod-name-1234567-aaaaa:uuid-uuid-uuid"
	traceID := "trace_it"
	spanID := "span_it"
	tenantID := "5000"
	env := "pilot"

	tr := &testRequest{
		meta:         "meta",
		data:         "data",
		msg:          "hello world",
		method:       method,
		path:         path,
		code:         http.StatusCreated,
		xRequestID:   requestID,
		xTraceID:     traceID,
		xSpanID:      spanID,
		xTenantCtx:   tenantID,
		xEnvironment: env,
	}

	req := httptest.NewRequest(method, path, nil)
	req.RequestURI = ""
	req.Header.Set(log.XRequestID, requestID)
	req.Header.Set(log.XTraceID, traceID)
	req.Header.Set(log.XSpanID, spanID)
	req.Header.Set(log.XTenantCtx, tenantID)
	req.Header.Set(log.XEnvironment, env)

	expectedEntity := &map[string]interface{}{
		tr.meta:            tr.data,
		"http_method":      tr.method,
		"http_path":        tr.path,
		"http_remote_addr": req.RemoteAddr,
		"x-request-id":     requestID,
		"x-trace-id":       traceID,
		"x-span-id":        spanID,
		"x-tenant-context": tenantID,
		"x-environment":    env,
		"src.app":          "my-pod-name",
	}
	positiveTest(req, t, tr, *expectedEntity)
}

func positiveTest(req *http.Request, t *testing.T, testRequest *testRequest, expectedEntity map[string]interface{}) {
	var entry log.Entry // NOTE(IB): hacky...

	ml := mLog{log.Noop()}
	mid := NewHTTPRequestMiddleware(ml, 0)

	handler := mid(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry = log.EntryFromCtx(r.Context())
		entry.WithStr(testRequest.meta, testRequest.data).Msg(testRequest.msg)
		w.WriteHeader(testRequest.code)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != testRequest.code {
		t.Fatalf("Wrong code from http handler: %d", testRequest.code)
	}

	ent := entry.(*mEntry)
	if !ent.async || !ent.sent || ent.lvl != log.Level(0) || ent.msg != testRequest.msg {
		t.Fatal("Entry fields incorrect")
	}

	if _, ok := ent.vals[log.ReqDuration]; !ok {
		t.Fatal("Request duration does not exist in log entry")
	}

	delete(ent.vals, log.ReqDuration)
	deepEqual := reflect.DeepEqual(ent.vals, expectedEntity)

	if !deepEqual {
		t.Fatalf("Unequal values in log entry: %v", ent.vals)
	}
}

func TestHTTPRequestMiddlewarePanic(t *testing.T) {
	ml := mLog{log.Noop()}
	mid := NewHTTPRequestMiddleware(ml, 0)

	pv := "this is fine"
	var entry log.Entry
	handler := mid(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entry = log.EntryFromCtx(r.Context())
		panic(pv)
	}))

	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	ent := entry.(*mEntry)

	if v, ok := ent.vals[log.PanicValue].(string); !ok || v != pv {
		t.Fatalf("log.PanicValue not what was expected: %v", v)
	}

	if st, ok := ent.vals[log.PanicStack].(errors.StackTrace); !ok || len(st) == 0 {
		t.Fatalf("log.PanicStack not type that was expected: %T", ent.vals[log.PanicStack])
	}

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("Unexpected status code from panic/recover: %d", rec.Code)
	}
}

// Some mock types with certain methods shadowed.

type mLog struct {
	log.Logger
}

func (l mLog) Entry(lvl log.Level) log.Entry {
	return &mEntry{
		Entry: l.Logger.Entry(lvl),
		lvl:   lvl,
		vals:  make(map[string]interface{}),
	}
}

type mEntry struct {
	log.Entry
	async bool
	sent  bool
	lvl   log.Level
	msg   string
	vals  map[string]interface{}
}

func (m *mEntry) WithStr(key string, strs ...string) log.Entry {
	m.vals[key] = strings.Join(strs, "")
	return m
}

func (m *mEntry) WithFields(fields map[string]interface{}) log.Entry {
	for k, v := range fields {
		m.vals[k] = v
	}

	return m
}

func (m *mEntry) Error() log.Entry {
	return m
}

func (m *mEntry) Msg(msg string) {
	m.msg = msg
}

func (m *mEntry) Async() log.Entry {
	m.async = true
	return m
}

func (m *mEntry) Send() {
	m.sent = true
}
