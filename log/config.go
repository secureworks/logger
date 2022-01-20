package log

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// These EnvKeys describe environmental variables used to set Config
// variables.
const (
	// LogLevel is the env var representing the log level. Values should
	// use our logger's representation: "TRACE", "DEBUG", etc.
	LogLevel EnvKey = "LOG_LEVEL"

	// LocalDevel is the env var representing the local debugging setting
	// for some logger implementations. Relevant values include: "true",
	// "True", "TRUE".
	LocalDevel EnvKey = "LOG_LOCAL_DEV"

	// Format is the env var representing the log format we want to use.
	// Relevant values include: "0" (JSONFormat) and "-1"
	// (ImplementationDefaultFormat).
	Format EnvKey = "LOG_FORMAT"

	// EnableErrStack is the env var representing whether we shall enables error
	// stack gathering and logging. Relevant values include: "true",
	// "True", "TRUE".
	EnableErrStack EnvKey = "ERROR_STACK"

	// SentryDSN is the env var representing the Sentry project DNS. An
	// empty value disables Sentry.
	SentryDSN EnvKey = "SENTRY_DSN"

	// SentryLevels is the env var representing which log levels will be
	// sent to Sentry. Values should use our logger's representation:
	// "TRACE", "DEBUG", etc., and be comma-separated.
	SentryLevels EnvKey = "SENTRY_LEVELS"

	// Environment is the env var representing the current deployment
	// environment. Values commonly used could be "dev", "prod", etc.
	Environment EnvKey = "ENVIRONMENT"

	// Release is the env var representing the current program release or
	// revision
	Release EnvKey = "SENTRY_RELEASE"

	// Server is the env var representing the current server or hostname.
	Server EnvKey = "SENTRY_SERVER"

	// SentryDebug is the env var representing the debug status for
	// Sentry. Relevant values include: "true", "True", "TRUE".
	SentryDebug EnvKey = "SENTRY_DEBUG"
)

// EnvKey is a publicly documented string type for environment lookups
// performed for DefaultConfig.
type EnvKey string

// String converts an EnvKey to a string.
func (ek EnvKey) String() string {
	return string(ek)
}

// Config defines common logger configuration options.
type Config struct {
	// Level is the level at which returned Logger's will be considered
	// enabled. For example, setting WARN then logging and sending a Debug
	// entry will cause the entry to not be logged.
	Level Level

	// LocalDevel, may be used by some logger implementations for local
	// debugging.
	LocalDevel bool

	// Format is the format the Logger should log in.
	Format LoggerFormat

	// EnableErrStack enables error stack gathering and logging.
	EnableErrStack bool

	// Output is the io.Writer the Logger will write messages to.
	Output io.Writer

	// Sentry is a sub-config type for configurating Sentry if desired. No
	// other portion of this struct is considered if DSN is not set and
	// valid.
	Sentry struct {
		// DSN is the Sentry DSN.
		DSN string

		// Release is the program release or revision.
		Release string

		// Env is the deployment environment; "prod", "dev", etc.
		Env string

		// Server is the server or hostname.
		Server string

		// Levels are the log levels that will trigger an event to be sent
		// to Sentry.
		Levels []Level

		// Debug is a passthrough for Sentry debugging.
		Debug bool
	}
}

// DefaultConfig returns a Config instance with sane defaults. env is a
// callback for looking up EnvKeys, it is set to os.Getenv if nil.
// Fields and values returned by this function can be altered.
func DefaultConfig(env func(string) string) *Config {
	config := new(Config)
	if env == nil {
		env = os.Getenv
	}

	// Level defaults to 0, ie INFO.
	if lvlStr := env(LogLevel.String()); lvlStr != "" {
		config.Level = LevelFromString(lvlStr)
	}
	if errStackStr := env(EnableErrStack.String()); errStackStr != "" {
		config.EnableErrStack = strings.ToUpper(errStackStr) == "TRUE"
	}
	if localDevel := env(LocalDevel.String()); localDevel != "" {
		config.LocalDevel = strings.ToUpper(localDevel) == "TRUE"
	}
	if format := env(Format.String()); format != "" {
		f, err := strconv.ParseInt(format, 10, 64)
		if err == nil { // FIXME(PH): swallows errors...
			config.Format = LoggerFormat(f)
		}
	}
	config.Output = os.Stderr // May not be set via environment.

	// SentryDSN must be set to use Sentry, so we only configure other
	// Sentry settings if it exists.
	sentryDSN := env(SentryDSN.String())
	if sentryDSN != "" {
		// Parse SentryLevels, fall back on FATAL, PANIC, ERROR.
		lvls := []Level{FATAL, PANIC, ERROR}
		split := strings.Split(env(SentryLevels.String()), ",")
		if len(split) > 0 && split[0] != "" {
			lvlSet := make(map[Level]bool, len(split))
			for _, lvl := range split {
				lvlSet[LevelFromString(lvl)] = true
			}

			lvls = make([]Level, 0, len(lvlSet))
			for lvl := range lvlSet {
				lvls = append(lvls, lvl)
			}
		}

		host, _ := os.Hostname()
		if server := env(Server.String()); server != "" {
			host = server
		}

		config.Sentry.DSN = sentryDSN
		config.Sentry.Levels = lvls
		config.Sentry.Release = env(Release.String())
		config.Sentry.Server = host
		config.Sentry.Env = env(Environment.String())
		if debug := env(SentryDebug.String()); debug != "" {
			config.Sentry.Debug = strings.ToUpper(debug) == "TRUE"
		}
	}
	if _, err := url.Parse(sentryDSN); err != nil {
		config.Sentry.DSN = ""
	}

	return config
}

// NOTE(PH): increase as we add logger implementations.
var loggerFactories = make(map[string]newLoggerFn, 4)

// newLoggerFn is a function type for Logger implemenations to register
// themselves.
type newLoggerFn func(*Config, ...Option) (Logger, error)

// Open returns a new instance of the selected Logger with config and
// options.
func Open(name string, conf *Config, opts ...Option) (Logger, error) {
	nl, ok := loggerFactories[name]
	if !ok {
		return nil, fmt.Errorf("log: No logger by name (%s)", name)
	}

	if conf == nil {
		conf = DefaultConfig(nil)
	}

	return nl(conf, opts...)
}

// Register registers the provided newLoggerFn function under the given
// name for use with Open. Note, this method is not concurreny safe, nil
// newLoggerFns or duplicate registration will cause a panic.
func Register(name string, nl func(*Config, ...Option) (Logger, error)) {
	if _, ok := loggerFactories[name]; ok || nl == nil {
		panic(fmt.Errorf("log: %s already registered with logging package", name))
	}
	loggerFactories[name] = nl
}

// Unregister removes any registered newLoggerFn function for the given
// name. Mostly useful for testing.
func Unregister(name string) bool {
	if _, ok := loggerFactories[name]; ok {
		delete(loggerFactories, name)
		return true
	}
	return false
}
