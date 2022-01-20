package zerolog

import (
	"encoding/json"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
	"github.com/secureworks/logger/internal/common"
	"github.com/secureworks/logger/internal/testutils"
	"github.com/secureworks/logger/log"
	"github.com/stretchr/testify/require"
)

func TestZerolog_SentryWriter_CheckLevel(t *testing.T) {
	t.Run("fails to find a level", func(t *testing.T) {
		writer := newSentryWriter(log.ERROR, log.PANIC, log.FATAL)

		zlvl, ok := writer.checkLevel(
			[]byte(`{"message":"test message here","error":"error message"}`))
		require.False(t, ok)
		require.Equal(t, zerolog.Level(0), zlvl)

		zlvl, ok = writer.checkLevel(
			[]byte(`"message":"test message here ...`))
		require.False(t, ok)
		require.Equal(t, zerolog.Level(0), zlvl)
	})

	t.Run("level found but is too low", func(t *testing.T) {
		writer := newSentryWriter(log.ERROR, log.PANIC, log.FATAL)

		zlvl, ok := writer.checkLevel(
			[]byte(`{"message":"test message here","error":"error message","level":"warn"}`))
		require.False(t, ok)
		require.Equal(t, zerolog.WarnLevel, zlvl)

		zlvl, ok = writer.checkLevel(
			[]byte(`"level":"warn","message":"test message here ...`))
		require.False(t, ok)
		require.Equal(t, zerolog.WarnLevel, zlvl)
	})

	t.Run("level is found and meets threshold", func(t *testing.T) {
		writer := newSentryWriter(log.ERROR, log.PANIC, log.FATAL)

		zlvl, ok := writer.checkLevel(
			[]byte(`{"message":"test message here","error":"error message","level":"error"}`))
		require.True(t, ok)
		require.Equal(t, zerolog.ErrorLevel, zlvl)

		zlvl, ok = writer.checkLevel(
			[]byte(`"level":"error","message":"test message here ...`))
		require.True(t, ok)
		require.Equal(t, zerolog.ErrorLevel, zlvl)
	})
}

func TestZerolog_SentryWriter_Write(t *testing.T) {
	require := require.New(t)

	srv, sentryMsg := testutils.SentryServer(t, false)
	defer srv.Close()

	common.InitSentry(sentry.ClientOptions{
		Dsn:           testutils.SentryDSN,
		HTTPTransport: srv.Transport(),
	})

	writer := newSentryWriter(log.ERROR, log.PANIC, log.FATAL)
	n, err := writer.Write(
		[]byte(`{"message":"test message here","error":"error message","level":"error","panic_value":"panic message"}`))
	require.Greater(n, 0)
	require.NoError(err)

	var event *sentry.Event
	err = json.Unmarshal(sentryMsg(t), &event)
	require.NoError(err)

	require.NotNil(event)
	require.Len(event.Exception, 2)
	require.Equal("error message", event.Exception[0].Value)
	require.Equal("panic message", event.Exception[1].Value)
}
