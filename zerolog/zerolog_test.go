package zerolog_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/secureworks/logger/log"
	"github.com/stretchr/testify/require"

	"github.com/secureworks/logger/internal/testutils"
)

const (
	testMessage    = "test message contents"
	testFieldValue = "test-field-value"
	testErrorValue = "new error message"
)

func TestZerolog_New(t *testing.T) {
	t.Run("log level too low does not log", func(t *testing.T) {
		require := require.New(t)
		config, out := testutils.NewConfigWithBuffer(t, log.INFO)

		logger, err := log.Open("zerolog", config)
		require.Nil(err)

		logger.Debug().Msg(testMessage)

		data, err := ioutil.ReadAll(out)
		require.Nil(err)
		require.Equal(len(data), 0) // Nothing is logged for debug when at INFO.
	})

	t.Run("log level matches does log", func(t *testing.T) {
		require := require.New(t)
		config, out := testutils.NewConfigWithBuffer(t, log.DEBUG)

		logger, err := log.Open("zerolog", config)
		require.Nil(err)

		logger.Debug().Msg(testMessage)

		data, err := ioutil.ReadAll(out)
		require.Nil(err)
		require.Contains(string(data), testMessage)
	})

	t.Run("configuration with nil output", func(t *testing.T) {
		require := require.New(t)
		config := log.DefaultConfig(nil)
		config.Output = nil

		logger, err := log.Open("zerolog", config)
		require.Nil(err)
		require.NotNil(logger)
	})
}

func TestZerolog_Logging(t *testing.T) {
	require := require.New(t)

	config, out := testutils.NewConfigWithBuffer(t, log.INFO)
	logger, err := log.Open("zerolog", config)
	require.NoError(err)

	logger.Info().WithStr("meta", testFieldValue).Msg(testMessage)

	var fields struct {
		Level   string `json:"level"`
		Meta    string `json:"meta"`
		Message string `json:"message"`
	}
	err = json.Unmarshal(out.Bytes(), &fields)
	require.NoError(err)

	require.Equal("info", fields.Level)
	require.Equal(testFieldValue, fields.Meta)
	require.Equal(testMessage, fields.Message)
}

func TestZerolog_Errors(t *testing.T) {
	require := require.New(t)

	srv, sentryMsg := testutils.SentryServer(t, false)
	defer srv.Close()

	config, _ := testutils.NewConfigWithBuffer(t, log.INFO)
	logger, err := log.Open("zerolog", config)
	require.NoError(err)

	testutils.BindSentryClient(t, srv.Transport()) // After logger instantiated.

	logger.WithError(errors.New(testErrorValue)).WithStr("meta", testFieldValue).Msg(testMessage)

	var event *sentry.Event
	err = json.Unmarshal(sentryMsg(t), &event)
	require.NoError(err)

	// Error value.
	require.NotNil(event)
	require.Len(event.Exception, 1)
	require.Equal(testErrorValue, event.Exception[0].Value)

	// Stack trace.
	require.NotNil(event.Exception[0].Stacktrace)
	require.Greater(len(event.Exception[0].Stacktrace.Frames), 0)

	// Metadata fields.
	extra, ok := event.Extra["meta"]
	require.True(ok)
	require.Equal(testFieldValue, extra)
}
