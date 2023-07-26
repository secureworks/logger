// Package log provides the unified interface for the Secureworks
// logger. This interface can use underlying logger implementations as
// drivers, including Logrus and Zerolog, along with support for
// reporting services including Sentry.
package log

import (
	"io"
	"strings"
	"time"
)

// Log levels for the unified interface. Underlying logger
// implementations must support these levels.
const (
	// TRACE Level.
	TRACE Level = iota + -2

	// DEBUG Level.
	DEBUG

	// INFO Level; this is the default (zero value).
	INFO

	// WARN Level.
	WARN

	// ERROR Level.
	ERROR

	// PANIC Level; note, depending on usage this will cause the logger to
	// panic.
	PANIC

	// FATAL Level; note, depending on usage this will cause the logger to
	// force a program exit.
	FATAL
)

// Level is the base type for logging levels supported by this package.
type Level int

// LevelFromString parses str and returns the closest level. If one
// isn't found the default level is returned.
func LevelFromString(str string) (lvl Level) {
	switch strings.ToUpper(str) {
	case "TRACE":
		lvl = TRACE
	case "DEBUG":
		lvl = DEBUG
	case "INFO":
		lvl = INFO
	case "WARN":
		lvl = WARN
	case "ERROR":
		lvl = ERROR
	case "PANIC":
		lvl = PANIC
	case "FATAL":
		lvl = FATAL
	}

	// Default case isn't needed, default is determined by enum zero
	// value.
	return
}

// IsValid checks if the current level is valid relative to known
// values.
func (l Level) IsValid() bool {
	return l >= TRACE && l <= FATAL
}

// IsEnabled checks if the level l is enabled relative to en.
func (l Level) IsEnabled(en Level) bool {
	return l.IsValid() && en.IsValid() && l >= en
}

// AllLevels is a convenience function returning all levels as a slice,
// ordered from lowest to highest precedence.
func AllLevels() []Level {
	return []Level{
		TRACE,
		DEBUG,
		INFO,
		WARN,
		ERROR,
		PANIC,
		FATAL,
	}
}

// Supported logging formats for the unified interface. To allow
// agnostic mixing of implmentations we default to JSON-formatting.
const (
	// JSONFormat is the default (zero value) format for loggers
	// registered with this package.
	JSONFormat LoggerFormat = 0

	// ImplementationDefaultFormat leaves the format up to the logger
	// implementation default.
	ImplementationDefaultFormat LoggerFormat = -1
)

// LoggerFormat is the base type for logging formats supported by this
// package.
type LoggerFormat int

// IsValid checks if a logger format is valid.
func (l LoggerFormat) IsValid() bool {
	switch l {
	case ImplementationDefaultFormat, JSONFormat:
		return true
	default:
		return false
	}
}

// Keys for standard logging fields. These keys can be used as map keys,
// JSON field names, or logger-implementation specific identifiers. By
// regularizing them we can make better assumptions about where to find
// and extract them.
const (
	// ReqDuration is a key for Logger data concerning HTTP request
	// logging.
	ReqDuration = "request_duration"

	// ReqPath is a key for Logger data concerning HTTP request logging.
	ReqPath = "http_path"

	// ReqMethod is a key for Logger data concerning HTTP request logging.
	ReqMethod = "http_method"

	// ReqRemoteAddr is a key for Logger data concerning HTTP request
	// logging.
	ReqRemoteAddr = "http_remote_addr"

	// PanicStack is a key for Logger data concerning errors and stack
	// traces.
	PanicStack = "panic_stack"

	// PanicValue is a key for Logger data concerning errors and stack
	// traces.
	PanicValue = "panic_value"

	// CallerField is a key for Logger data concerning errors and stack
	// traces.
	CallerField = "caller"

	// StackField is a key for Logger data concerning errors and stack
	// traces.
	StackField = "stack"
)

// Unified interface definitions.

// Logger is the minimum interface loggers should implement when used
// with CTPx packages.
type Logger interface {
	// WriteCloser returns an io.Writer that when written to writes logs
	// at the given level. It is the callers responsibility to call Close
	// when finished. This is particularly useful for redirecting the
	// output of other loggers or even Readers with the help of
	// io.TeeReader.
	WriteCloser(Level) io.WriteCloser

	// WithError attaches the given error into a new Entry and returns the
	// Entry.
	WithError(err error) Entry

	// WithField inserts the key and value into a new Entry (as tags or
	// metadata information) and returns the Entry.
	WithField(key string, value interface{}) Entry

	// WithFields inserts the given set of fields into a new Entry and
	// returns the Entry.
	WithFields(fields map[string]interface{}) Entry

	// Entry returns a new Entry at the provided log level.
	Entry(Level) Entry

	// Trace returns a new Entry at TRACE level.
	Trace() Entry

	// Debug returns a new Entry at DEBUG level.
	Debug() Entry

	// Info returns a new Entry at INFO level.
	Info() Entry

	// Warn returns a new Entry at WARN level.
	Warn() Entry

	// Error returns a new Entry at ERROR level.
	Error() Entry

	// Panic returns a new Entry at PANIC level. Implementations should
	// panic once the final message for the Entry is logged.
	Panic() Entry

	// Fatal returns a new Entry at FATAL level. Implementations should
	// exit non-zero once the final message for the Entry is logged.
	Fatal() Entry
}

// Entry is the primary interface by which individual log entries are
// made.
type Entry interface {
	// Async flips the current Entry to be asynchronous, or back if called
	// more than once. If set to asynchronous an Entry implementation
	// should not write to output until Send is called.
	Async() Entry

	// Send sends (writes) the current entry. This interface does not
	// define the behavior of calling this method more than once.
	Send()

	// Msgf formats and sets the final log message for this Entry. It will
	// also send the message if Async has not been set.
	Msgf(string, ...interface{})

	// Msg sets the final log message for this Entry. It will also send
	// the message if Async has not been set.
	Msg(msg string)

	// Caller embeds a caller value into the existing Entry. A caller
	// value is a filepath followed by line number. Skip determines the
	// number of additional stack frames to ascend when determining the
	// value. By default the caller of the method is the value used and
	// skip does not need to be supplied in that case. Caller may be
	// called multiple times on the Entry to build a stack or execution
	// trace.
	Caller(skip ...int) Entry

	// WithError attaches the given errors into a new Entry and returns
	// the Entry. Depending on the logger implementation, multiple errors
	// may be inserted as a slice of errors or a single multi-error.
	// Calling the method more than once will overwrite the attached
	// error(s) and not append them.
	WithError(errs ...error) Entry

	// WithField inserts the key and value into the Entry (as tags or
	// metadata information) and returns the Entry.
	WithField(key string, value interface{}) Entry

	// WithFields inserts the given set of fields into the Entry and
	// returns the Entry.
	WithFields(fields map[string]interface{}) Entry

	// WithStr is a type-safe convenience for injecting a string (or
	// strings, how they are stored is implmentation-specific) field.
	WithStr(key string, strs ...string) Entry

	// WithBool is a type-safe convenience for injecting a Boolean (or
	// Booleans, how they are stored is implmentation-specific) field.
	WithBool(key string, bls ...bool) Entry

	// WithDur is a type-safe convenience for injecting a time.Duration
	// (or time.Durations, how they are stored is implmentation-specific)
	// field.
	WithDur(key string, durs ...time.Duration) Entry

	// WithInt is a type-safe convenience for injecting an integer (or
	// integers, how they are stored is implmentation-specific) field.
	WithInt(key string, is ...int) Entry

	// WithUint is a type-safe convenience for injecting an unsigned
	// integer (or unsigned integers, how they are stored is
	// implmentation-specific) field.
	WithUint(key string, us ...uint) Entry

	// WithTime is a type-safe convenience for injecting a time.Time (or
	// time.Times, how they are stored is implmentation-specific) field.
	//
	// NOTE(IB): many loggers add a "time" key automatically and time
	// formatting may be dependant on configuration or logger choice.
	WithTime(key string, ts ...time.Time) Entry

	// Trace updates the Entry's level to TRACE.
	Trace() Entry

	// Debug updates the Entry's level to DEBUG.
	Debug() Entry

	// Info updates the Entry's level to INFO.
	Info() Entry

	// Warn updates the Entry's level to WARN.
	Warn() Entry

	// Error updates the Entry's level to ERROR.
	Error() Entry

	// Panic updates the Entry's level to PANIC. Implementations should
	// panic once the final message for the Entry is logged.
	Panic() Entry

	// Fatal updates the Entry's level to FATAL. Implementations should
	// exit non-zero once the final message for the Entry is logged.
	Fatal() Entry
}

// UnderlyingLogger is an escape hatch allowing Loggers registered with
// this package the option to return their underlying implementation, as
// well as reset it.
//
// NOTE(IB): this is currently required for CustomOptions to work.
type UnderlyingLogger interface {
	GetLogger() interface{}
	SetLogger(interface{})
}
