package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/VerticalOps/fakesentry"
	"github.com/getsentry/sentry-go"
	"github.com/secureworks/logger/log"
	"github.com/secureworks/logger/middleware"
	"github.com/secureworks/logger/testlogger"
)

const SentryDSN = `http://thisis:myfakeauth@localhost/1`

// NewConfigWithBuffer generates a default testing config (log.Level is
// set to log.INFO and EnableErrStack is true) and a linked output
// buffer that the logger writes too.
func NewConfigWithBuffer(t *testing.T, logLevel log.Level) (*log.Config, *bytes.Buffer) {
	t.Helper()

	buf := make([]byte, 0, 100)
	out := bytes.NewBuffer(buf)
	config := log.DefaultConfig(func(envvar string) string {
		if envvar == log.SentryDSN.String() {
			return SentryDSN // Ensure a standard Sentry DSN.
		}
		return os.Getenv(envvar)
	})
	config.Level = logLevel
	config.Output = out
	config.EnableErrStack = true

	return config, out
}

// BindSentryClient attaches a Sentry server transport (from fake
// Sentry) to the Sentry SDK's CurrentHub .It assumes that a logger has
// been instantiated, which initializes the Sentry SDK.
// Note this is data-race free but not race-condition free on the Sentry Hub.
// Use with caution from multiple goroutines.
func BindSentryClient(t *testing.T, tcp *http.Transport) {
	t.Helper()

	clientOpts := sentry.CurrentHub().Client().Options()
	clientOpts.HTTPTransport = tcp
	client, err := sentry.NewClient(clientOpts)
	if err != nil {
		t.Fatalf("failed to init sentry client: %+v", err)
	}

	sentry.CurrentHub().BindClient(client)
}

// SentryServer generates a new fake Sentry server to field requests,
// and binds the "CurrentHub" client to it. The server instance is
// returned as well as a function that returns received messages (bytes)
// or times out and fails the test.
func SentryServer(t *testing.T, logMessages bool) (fakesentry.Server, func(t *testing.T) []byte) {
	t.Helper()

	messageCh := make(chan []byte, 1)

	sentrySrv := fakesentry.NewUnstartedServer()
	sentrySrv.Server = &http.Server{Handler: fakesentry.NewHandler(
		fakesentry.AsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Helper()

			jsonBytes, ok := fakesentry.FromRequest(r)
			if !ok {
				t.Fatalf("could not extract Sentry JSON message from request")
			}

			if testing.Verbose() && logMessages {
				buf := new(bytes.Buffer)
				json.Indent(buf, jsonBytes, "", "  ")
				fmt.Fprintf(os.Stderr, `
========================================================================
Sentry server message received:
%s
========================================================================
`, buf.Bytes())
			}

			select {
			case messageCh <- jsonBytes:
			default:
			}
		})),
	)}

	// Start the server and bind a client to the CurrentHub.
	go func() {
		_ = sentrySrv.Serve(sentrySrv.Listener())
	}()

	// Receives a message from the channel and returns it or times out.
	nextMessage := func(t *testing.T) []byte {
		var byt []byte

		timer := time.NewTimer(time.Millisecond * 500)
		select {
		case <-timer.C:
			t.Fatal("Failed to get logger data")
		case byt = <-messageCh:
			timer.Stop()
		}
		return byt
	}

	return sentrySrv, nextMessage
}

// RunMiddlewareAround wraps a default logging middleware setup around
// the given handler, executes the given request against it and returns
// the ResponseRecorder and the test logger involved.
func RunMiddlewareAround(
	t *testing.T,
	req *http.Request,
	entries *middleware.HTTPRequestLogAttributes,
	handler http.Handler,
) (*httptest.ResponseRecorder, *testlogger.Logger) {
	t.Helper()

	logger, _ := testlogger.New(log.DefaultConfig(nil))
	resp := httptest.NewRecorder()
	h := middleware.NewHTTPRequestMiddleware(
		logger,
		log.INFO,
		entries,
	)(handler)
	h.ServeHTTP(resp, req)

	return resp, logger.(*testlogger.Logger)
}
