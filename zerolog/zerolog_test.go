package zerolog_test

import (
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/secureworks/errors"
	"github.com/secureworks/logger/internal/testutils"
	"github.com/secureworks/logger/log"
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

		data, err := io.ReadAll(out)
		testutils.AssertNil(t, err)
		testutils.AssertEqual(t, len(data), 0) // Nothing is logged for debug when at INFO.
	})

	t.Run("log level matches does log", func(t *testing.T) {
		config, out := testutils.NewConfigWithBuffer(t, log.DEBUG)

		logger, err := log.Open("zerolog", config)
		testutils.AssertNil(t, err)

		logger.Debug().Msg(testMessage)

		data, err := io.ReadAll(out)
		testutils.AssertNil(t, err)
		testutils.AssertStringContains(t, testMessage, string(data))
	})

	t.Run("configuration with nil output", func(t *testing.T) {
		config := log.DefaultConfig(nil)
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

func TestZerolog_Errors(t *testing.T) {
	config, out := testutils.NewConfigWithBuffer(t, log.INFO)
	logger, err := log.Open("zerolog", config)
	testutils.AssertNil(t, err)

	logger.WithError(errors.New(testErrorValue)).WithStr("meta", testFieldValue).Msg(testMessage)

	var fields struct {
		Error   string `json:"error"`
		Level   string `json:"level"`
		Meta    string `json:"meta"`
		Message string `json:"msg"`
		Stack   []struct {
			File string `json:"file"`
			Line int    `json:"line"`
			Func string `json:"function"`
		} `json:"stack"`
		Time time.Time `json:"time"`
	}
	err = json.Unmarshal(out.Bytes(), &fields)
	testutils.AssertNil(t, err)

	// Error value.
	testutils.AssertNotNil(t, fields)
	testutils.AssertEqual(t, testErrorValue, fields.Error)

	// Stack trace.
	var files, funcs []string
	for _, f := range fields.Stack {
		files = append(files, f.File)
		funcs = append(funcs, f.Func)
	}
	testutils.AssertNotNil(t, fields.Stack)
	testutils.AssertTrue(t, len(fields.Stack) > 0)
	testutils.AssertAnyStringContains(t, "zerolog_test.go", files)
	testutils.AssertAnyStringContains(t, "zerolog_test.TestZerolog_Errors", funcs)

	// Metadata fields.
	testutils.AssertEqual(t, testFieldValue, fields.Meta)

	// Nil error stack trace.
	testutils.AssertNotPanics(t, func() { logger.WithError(nil).Msg("done") })
}
