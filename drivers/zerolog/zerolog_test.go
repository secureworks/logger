package zerolog_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/secureworks/logger/log"
	"github.com/secureworks/logger/log/testutils"
)

const (
	testMessage    = "test message contents"
	testFieldValue = "test-field-value"
	testErrorValue = "new error message"
)

func TestZerolog_New(t *testing.T) {
	t.Run("log level too low does not log", func(t *testing.T) {
		config, out := testutils.NewConfigWithBuffer(t, log.INFO)

		logger, err := log.Open("zerolog", config)
		testutils.AssertNil(t, err)

		logger.Debug().Msg(testMessage)

		data, err := ioutil.ReadAll(out)
		testutils.AssertNil(t, err)
		testutils.AssertEqual(t, len(data), 0) // Nothing is logged for debug when at INFO.
	})

	t.Run("log level matches does log", func(t *testing.T) {
		config, out := testutils.NewConfigWithBuffer(t, log.DEBUG)

		logger, err := log.Open("zerolog", config)
		testutils.AssertNil(t, err)

		logger.Debug().Msg(testMessage)

		data, err := ioutil.ReadAll(out)
		testutils.AssertNil(t, err)
		testutils.AssertStringContains(t, testMessage, string(data))
	})

	t.Run("configuration with nil output", func(t *testing.T) {
		config := log.DefaultConfig()
		config.Output = nil

		logger, err := log.Open("zerolog", config)
		testutils.AssertNil(t, err)
		testutils.AssertNotNil(t, logger)
	})
}

func TestZerolog_Logging(t *testing.T) {
	config, out := testutils.NewConfigWithBuffer(t, log.INFO)
	logger, err := log.Open("zerolog", config)
	testutils.AssertNil(t, err)

	logger.Info().WithStr("meta", testFieldValue).Msg(testMessage)

	var fields struct {
		Level   string `json:"level"`
		Meta    string `json:"meta"`
		Message string `json:"message"`
	}
	err = json.Unmarshal(out.Bytes(), &fields)
	testutils.AssertNil(t, err)

	testutils.AssertEqual(t, "info", fields.Level)
	testutils.AssertEqual(t, testFieldValue, fields.Meta)
	testutils.AssertEqual(t, testMessage, fields.Message)
}

// func TestZerolog_Errors(t *testing.T) {
// 	srv, sentryMsg := testutils.SentryServer(t, false)
// 	defer srv.Close()
//
// 	config, _ := testutils.NewConfigWithBuffer(t, log.INFO)
// 	logger, err := log.Open("zerolog", config)
// 	testutils.AssertNil(t, err)
//
// 	testutils.BindSentryClient(t, srv.Transport()) // After logger instantiated.
//
// 	logger.WithError(errors.New(testErrorValue)).WithStr("meta", testFieldValue).Msg(testMessage)
//
// 	var event *sentry.Event
// 	err = json.Unmarshal(sentryMsg(t), &event)
// 	testutils.AssertNil(t, err)
//
// 	// Error value.
// 	testutils.AssertNotNil(t, event)
// 	testutils.AssertEqual(t, 1, len(event.Exception))
// 	testutils.AssertEqual(t, testErrorValue, event.Exception[0].Value)
//
// 	// Stack trace.
// 	testutils.AssertNotNil(t, event.Exception[0].Stacktrace)
// 	testutils.AssertTrue(t, len(event.Exception[0].Stacktrace.Frames) > 0)
//
// 	// Metadata fields.
// 	extra, ok := event.Extra["meta"]
// 	testutils.AssertTrue(t, ok)
// 	testutils.AssertEqual(t, testFieldValue, extra)
// }
