package log_test

import (
	"context"
	"testing"

	"github.com/secureworks/logger/log"
	_ "github.com/secureworks/logger/testlogger"
	"github.com/stretchr/testify/require"
)

func TestLog_ContextUtilities(t *testing.T) {
	t.Run("Logger", func(t *testing.T) {
		ctx := context.Background()
		logger, _ := log.Open("test", nil)

		ctx = log.CtxWithLogger(ctx, logger)

		require.Equal(t, logger, log.LoggerFromCtx(ctx))
	})

	t.Run("Entry", func(t *testing.T) {
		ctx := context.Background()
		logger, _ := log.Open("test", nil)
		entry := logger.Entry(log.INFO).Async()

		ctx = log.CtxWithEntry(ctx, entry)

		require.Equal(t, entry, log.EntryFromCtx(ctx))
	})
}
