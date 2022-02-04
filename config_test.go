package logger_test

import (
	"os"
	"sort"
	"strconv"
	"testing"

	"github.com/secureworks/logger/internal/testutils"
	"github.com/secureworks/logger/log"
	"github.com/secureworks/logger/testlogger"
)

var defaultConfig = &log.Config{
	Level:          log.INFO,
	LocalDevel:     false,
	Format:         log.JSONFormat,
	EnableErrStack: false,
	Output:         os.Stderr,
	Sentry: struct {
		DSN     string
		Release string
		Env     string
		Server  string
		Levels  []log.Level
		Debug   bool
	}{
		DSN:     "",
		Release: "",
		Env:     "",
		Server:  "",
		Levels:  nil,
		Debug:   false,
	},
}

var loadedConfig = &log.Config{
	Level:          log.DEBUG,
	LocalDevel:     true,
	Format:         log.ImplementationDefaultFormat,
	EnableErrStack: true,
	Output:         os.Stderr,
	Sentry: struct {
		DSN     string
		Release string
		Env     string
		Server  string
		Levels  []log.Level
		Debug   bool
	}{
		DSN:     "https://example.com/test",
		Release: "app-1-a",
		Env:     "prod",
		Server:  "app-main",
		Levels:  []log.Level{log.FATAL, log.PANIC, log.ERROR, log.WARN},
		Debug:   true,
	},
}

func TestDefaultConfig(t *testing.T) {
	t.Run("with an empty environment", func(t *testing.T) {
		config := log.DefaultConfig(func(string) string { return "" }) // Ignore env.
		testutils.AssertEqual(t, defaultConfig, config)
	})

	t.Run("with environment variables", func(t *testing.T) {
		fakeenv := map[string]string{
			"LOG_LEVEL":      "DEBUG",
			"LOG_LOCAL_DEV":  "true",
			"LOG_FORMAT":     strconv.Itoa(int(log.ImplementationDefaultFormat)),
			"ERROR_STACK":    "true",
			"SENTRY_DSN":     loadedConfig.Sentry.DSN,
			"SENTRY_LEVELS":  "FATAL,PANIC,ERROR,WARN",
			"SENTRY_RELEASE": loadedConfig.Sentry.Release,
			"ENVIRONMENT":    loadedConfig.Sentry.Env,
			"SENTRY_SERVER":  loadedConfig.Sentry.Server,
			"SENTRY_DEBUG":   "true",
		}

		config := log.DefaultConfig(func(varname string) string { return fakeenv[varname] })
		// Simplest way to ensure we don't get false negatives.
		sort.Slice(
			config.Sentry.Levels, func(i, j int) bool {
				return config.Sentry.Levels[i] > config.Sentry.Levels[j]
			},
		)
		testutils.AssertEqual(t, loadedConfig, config)
	})

	t.Run("with Sentry config but missing Sentry DSN", func(t *testing.T) {
		fakeenv := map[string]string{
			"SENTRY_LEVELS":  "FATAL,PANIC,ERROR,WARN",
			"SENTRY_RELEASE": "app-1-a",
			"ENVIRONMENT":    "prod",
			"SENTRY_SERVER":  "app-main",
			"SENTRY_DEBUG":   "true",
		}

		config := log.DefaultConfig(func(varname string) string {
			return fakeenv[varname]
		})
		testutils.AssertEqual(t, defaultConfig, config)
	})
}

func TestOpenRegister(t *testing.T) {
	t.Run("Open before Register fails", func(t *testing.T) {
		logger, err := log.Open("newlogger", nil)
		testutils.AssertNil(t, logger)
		testutils.AssertEqual(t, "log: No logger by name (newlogger)", err.Error())
	})

	t.Run("Open after Register succeeds", func(t *testing.T) {
		defer log.Unregister("newlogger")

		log.Register("newlogger", func(c *log.Config, opts ...log.Option) (log.Logger, error) {
			return testlogger.New(c, opts...)
		})

		logger, err := log.Open("newlogger", nil)
		testutils.AssertNil(t, err)
		testutils.AssertNotNil(t, logger)
	})

	t.Run("Open with config sets config", func(t *testing.T) {
		logger, err := log.Open("test", nil)
		testutils.AssertNil(t, err)

		// Test logger uses a bytes.Buffer for Output by default instead of
		// os.Stderr, so let's just reset that.
		config := logger.(*testlogger.Logger).Config
		config.Output = os.Stderr
		testutils.AssertEqual(t, defaultConfig, config)

		logger, err = log.Open("test", loadedConfig)
		testutils.AssertNil(t, err)

		config = logger.(*testlogger.Logger).Config
		config.Output = os.Stderr
		testutils.AssertEqual(t, loadedConfig, config)
	})
}
