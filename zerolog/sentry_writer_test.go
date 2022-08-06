package zerolog

import (
	"encoding/json"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"

	"github.com/secureworks/logger/internal/common"
	"github.com/secureworks/logger/internal/testutils"
	"github.com/secureworks/logger/log"
)

func TestZerolog_SentryWriter_CheckLevel(t *testing.T) {
	t.Run("fails to find a level", func(t *testing.T) {
		writer := newSentryWriter(log.ERROR, log.PANIC, log.FATAL)

		zlvl, ok := writer.checkLevel(
			[]byte(`{"message":"test message here","error":"error message"}`))
		testutils.AssertFalse(t, ok)
		testutils.AssertEqual(t, zerolog.Level(0), zlvl)

		zlvl, ok = writer.checkLevel(
			[]byte(`"message":"test message here ...`))
		testutils.AssertFalse(t, ok)
		testutils.AssertEqual(t, zerolog.Level(0), zlvl)
	})

	t.Run("level found but is too low", func(t *testing.T) {
		writer := newSentryWriter(log.ERROR, log.PANIC, log.FATAL)

		zlvl, ok := writer.checkLevel(
			[]byte(`{"message":"test message here","error":"error message","level":"warn"}`))
		testutils.AssertFalse(t, ok)
		testutils.AssertEqual(t, zerolog.WarnLevel, zlvl)

		zlvl, ok = writer.checkLevel(
			[]byte(`"level":"warn","message":"test message here ...`))
		testutils.AssertFalse(t, ok)
		testutils.AssertEqual(t, zerolog.WarnLevel, zlvl)
	})

	t.Run("level is found and meets threshold", func(t *testing.T) {
		writer := newSentryWriter(log.ERROR, log.PANIC, log.FATAL)

		zlvl, ok := writer.checkLevel(
			[]byte(`{"message":"test message here","error":"error message","level":"error"}`))
		testutils.AssertTrue(t, ok)
		testutils.AssertEqual(t, zerolog.ErrorLevel, zlvl)

		zlvl, ok = writer.checkLevel(
			[]byte(`"level":"error","message":"test message here ...`))
		testutils.AssertTrue(t, ok)
		testutils.AssertEqual(t, zerolog.ErrorLevel, zlvl)
	})
}

func TestZerolog_SentryWriter_Write(t *testing.T) {
	srv, sentryMsg := testutils.SentryServer(t, false)
	defer srv.Close()

	_ = common.InitSentry(sentry.ClientOptions{
		Dsn:           testutils.SentryDSN,
		HTTPTransport: srv.Transport(),
	})

	writer := newSentryWriter(log.ERROR, log.PANIC, log.FATAL)
	n, err := writer.Write(
		[]byte(`{"message":"test message here","error":"error message","level":"error","panic_value":"panic message"}`))
	testutils.AssertTrue(t, n > 0)
	testutils.AssertNil(t, err)

	var event *sentry.Event
	err = json.Unmarshal(sentryMsg(t), &event)
	testutils.AssertNil(t, err)

	testutils.AssertNotNil(t, event)
	testutils.AssertEqual(t, 2, len(event.Exception))
	testutils.AssertEqual(t, "error message", event.Exception[0].Value)
	testutils.AssertEqual(t, "panic message", event.Exception[1].Value)
}
