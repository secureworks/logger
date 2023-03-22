package logger_test

import (
	"context"
	"testing"

	_ "github.com/secureworks/logger/drivers/testlogger"
	"github.com/secureworks/logger/log"
	"github.com/secureworks/logger/log/testutils"
)

func TestLog_ContextUtilities(t *testing.T) {
	t.Run("Logger", func(t *testing.T) {
		ctx := context.Background()
		logger, _ := log.Open("test", nil)

		ctx = log.ContextWithLogger(ctx, logger)

		testutils.AssertEqual(t, logger, log.LoggerFromContext(ctx))
	})

	t.Run("Entry", func(t *testing.T) {
		ctx := context.Background()
		logger, _ := log.Open("test", nil)
		entry := logger.Entry(log.INFO).Async()

		ctx = log.ContextWithEntry(ctx, entry)

		testutils.AssertEqual(t, entry, log.EntryFromContext(ctx))
	})
}
