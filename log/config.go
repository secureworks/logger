package log

import (
	"fmt"
	"io"
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

	// EnableErrStack is the env var representing whether we shall enable
	// error stack gathering and logging. Relevant values include: "true",
	// "True", "TRUE".
	EnableErrStack EnvKey = "ERROR_STACK"

	// Environment is the env var representing the current deployment
	// environment. Values commonly used could be "dev", "prod", etc.
	//
	// TODO(PH): do we keep or remove this?
	Environment EnvKey = "ENVIRONMENT"
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
	return config
}

// NOTE(PH): increase as we add logger implementations.
var loggerFactories = make(map[string]newLoggerFn, 4)

// newLoggerFn is a function type for Logger implemenations to register
// themselves.
type newLoggerFn func(*Config, ...Option) (Logger, error)

// Open returns a new instance of the selected Logger with config and
// options.
func Open(name string, config *Config, opts ...Option) (Logger, error) {
	nl, ok := loggerFactories[name]
	if !ok {
		return nil, fmt.Errorf("log: No logger by name (%s)", name)
	}

	if config == nil {
		config = DefaultConfig(nil)
	}

	return nl(config, opts...)
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
