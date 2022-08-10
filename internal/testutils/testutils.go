package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/VerticalOps/fakesentry"
	"github.com/getsentry/sentry-go"

	"github.com/secureworks/logger/log"
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
				_ = json.Indent(buf, jsonBytes, "", "  ")
				_, _ = fmt.Fprintf(os.Stderr, `
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
		t.Helper()

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

// AssertTrue is a semantic test assertion for object truthiness.
func AssertTrue(t *testing.T, object bool) {
	t.Helper()
	if !object {
		t.Errorf("is not true")
	}
}

// AssertFalse is a semantic test assertion for object truthiness.
func AssertFalse(t *testing.T, object bool) {
	t.Helper()
	if object {
		t.Errorf("is not false")
	}
}

// AssertEqual is a semantic test assertion for object equality.
func AssertEqual(t *testing.T, expected interface{}, actual interface{}) {
	t.Helper()
	assertEquality(t, expected, actual, true)
}

// AssertNotEqual is a semantic test assertion for object equality.
func AssertNotEqual(t *testing.T, expected interface{}, actual interface{}) {
	t.Helper()
	assertEquality(t, expected, actual, false)
}

// AssertNearEqual is a semantic test assertion for numeric accuracy.
func AssertNearEqual(t *testing.T, expected int64, actual int64, delta int64) {
	t.Helper()

	diff := actual - expected
	if diff < -delta || diff > delta {
		t.Errorf(
			"is not within delta (%d):\nexpected: %d\nactual: %d\n",
			delta,
			expected,
			actual,
		)
	}
}

// AssertSame is a semantic test assertion for referential equality.
func AssertSame(t *testing.T, expected interface{}, actual interface{}) {
	t.Helper()
	assertSameness(t, expected, actual, true)
}

// AssertNotSame is a semantic test assertion for referential equality.
func AssertNotSame(t *testing.T, expected interface{}, actual interface{}) {
	t.Helper()
	assertSameness(t, expected, actual, false)
}

// AssertNil is a semantic test assertion for nility.
func AssertNil(t *testing.T, object interface{}) {
	t.Helper()
	assertNility(t, object, true)
}

// AssertNotNil is a semantic test assertion for nility.
func AssertNotNil(t *testing.T, object interface{}) {
	t.Helper()
	assertNility(t, object, false)
}

// AssertStringContains is a semantic test assertion for partial string matching.
func AssertStringContains(t *testing.T, expectedContained string, actualContaining string) {
	t.Helper()
	if !strings.Contains(actualContaining, expectedContained) {
		t.Errorf(
			"does not contain:\nexpected to contain: %s\nactual: %s\n",
			strings.Trim(expectedContained, "\n"),
			strings.Trim(actualContaining, "\n"),
		)
	}
}

// NOTE(PH): does not handle bytes well, update if we need to check
// them.
func assertEquality(t *testing.T, expected interface{}, actual interface{}, wantEqual bool) {
	t.Helper()

	if expected == nil && actual == nil && !wantEqual {
		t.Errorf("is equal:\nexpected: %s\nactual: %s\n", expected, actual)
	}

	isEqual := reflect.DeepEqual(expected, actual)
	if wantEqual && !isEqual {
		t.Errorf("not equal:\nexpected: %s\nactual: %s\n", expected, actual)
	}
	if !wantEqual && isEqual {
		t.Errorf("is equal:\nexpected: %s\nactual: %s\n", expected, actual)
	}
}

func assertSameness(t *testing.T, expected interface{}, actual interface{}, wantSame bool) {
	t.Helper()

	isSame := false
	if expected == actual {
		isSame = true
		expectedPtr := reflect.ValueOf(expected)
		actualPtr := reflect.ValueOf(actual)
		if expectedPtr.Kind() != reflect.Ptr || actualPtr.Kind() != reflect.Ptr {
			isSame = false
		}

		expectedType := reflect.TypeOf(expected)
		actualType := reflect.TypeOf(actual)
		if isSame && (expectedType != actualType) {
			isSame = false
		}
	}
	if wantSame && !isSame {
		t.Errorf("not same:\nexpected: %s\nactual: %s\n", expected, actual)
	}
	if !wantSame && isSame {
		t.Errorf("is same:\nexpected: %s\nactual: %s\n", expected, actual)
	}
}

func assertNility(t *testing.T, object interface{}, wantNil bool) {
	t.Helper()

	isNil := object == nil
	if !isNil {
		value := reflect.ValueOf(object)
		isNilable := false
		switch value.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
			isNilable = true
		default:
		}
		if isNilable && value.IsNil() {
			isNil = true
		}
	}
	if wantNil && !isNil {
		t.Errorf("not nil: %s\n", object)
	}
	if !wantNil && isNil {
		t.Errorf("is nil: %s\n", object)
	}
}
