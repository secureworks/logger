package logger_test

import (
	"context"
	"testing"

	"github.com/secureworks/logger/internal/testutils"
	"github.com/secureworks/logger/log"
	_ "github.com/secureworks/logger/testlogger"
)

func TestLog_ContextUtilities(t *testing.T) {
	t.Run("Logger", func(t *testing.T) {
		ctx := context.Background()
		logger, _ := log.Open("test", nil)

		ctx = log.CtxWithLogger(ctx, logger)

		testutils.AssertEqual(t, logger, log.LoggerFromCtx(ctx))
	})

	t.Run("Entry", func(t *testing.T) {
		ctx := context.Background()
		logger, _ := log.Open("test", nil)
		entry := logger.Entry(log.INFO).Async()

		ctx = log.CtxWithEntry(ctx, entry)

		testutils.AssertEqual(t, entry, log.EntryFromCtx(ctx))
	})
}
