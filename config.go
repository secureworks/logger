package log

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

// EnvKey is a publicly documented string type for environment lookups
// performed for DefaultConfig. It is otherwise unspecial.
type EnvKey string

const (
	// LogLevel is the EnvKey used for looking up the log level.
	LogLevel EnvKey = "LOG_LEVEL"

	// ErrStack is the EnvKey used for looking up error/stack trace
	// logging; Value should be true || True || TRUE.
	ErrStack EnvKey = "ERROR_STACK"

	// SentryDSN is the EnvKey used for looking up the Sentry project DNS;
	// Empty value disables Sentry.
	SentryDSN EnvKey = "SENTRY_DSN"

	// SentryLevels is the EnvKey used for looking up which log levels
	// will be sent to Sentry; Values should be comma separated.
	SentryLevels EnvKey = "SENTRY_LEVELS"

	// Environment is the EnvKey used for looking up the current
	// deployment environment; Values commonly dev || prod.
	Environment EnvKey = "ENVIRONMENT"
)

func (ek EnvKey) String() string {
	return string(ek)
}

// Config defines common logger configuration options.
type Config struct {
	// Level is the level at which returned Logger's will be considered
	// enabled. For example, setting WARN then logging and sending a Debug
	// entry will cause the entry to not be logged.
	Level Level

	// LocalDevel, may be used by some Logger's for local debugging
	// changes.
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
	// TODO(IB): Configuration is still an issue, as discussed during
	// design phase. If we use consol + vault, most of our values should
	// become env vars. This is a work in progress.
	conf := new(Config)

	if env == nil {
		env = os.Getenv
	}

	if lvlStr := env(LogLevel.String()); lvlStr != "" {
		conf.Level = LevelFromString(lvlStr)
	}

	if errStackStr := env(ErrStack.String()); errStackStr != "" {
		conf.EnableErrStack = strings.ToUpper(errStackStr) == "TRUE"
	}

	sentryDSN := env(SentryDSN.String())

	if sentryDSN != "" {
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

		conf.Sentry.DSN = sentryDSN
		conf.Sentry.Levels = lvls
		// FIXME(PH): https://missione.atlassian.net/browse/CX-17598
		// conf.Sentry.Release = version.Get().Revision
		conf.Sentry.Server = host
		conf.Sentry.Env = env(Environment.String())
	}

	if _, err := url.Parse(sentryDSN); err != nil {
		conf.Sentry.DSN = ""
	}

	conf.Output = os.Stderr
	return conf
}

// NewLogger is a function type for Logger implemenations to register
// themselves.
type NewLogger func(*Config, ...Option) (Logger, error)

var (
	setup = make(map[string]NewLogger, 2)
)

// Register registers the provided NewLogger function under the given
// name for use with Open. Note, this method is not concurreny safe, nil
// NewLoggers or duplicate registration will cause a panic.
func Register(name string, nl NewLogger) {
	if _, ok := setup[name]; ok || nl == nil {
		panic(fmt.Errorf("log: %s already registered with logging package", name))
	}

	setup[name] = nl
}

// Open returns a new instance of the selected Logger with config and
// options.
func Open(name string, conf *Config, opts ...Option) (Logger, error) {
	nl, ok := setup[name]
	if !ok {
		return nil, fmt.Errorf("log: No logger by name (%s)", name)
	}

	if conf == nil {
		conf = DefaultConfig(nil)
	}

	return nl(conf, opts...)
}
