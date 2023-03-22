// Package log provides the unified interface for the Secureworks
// logger. This interface can use underlying logger implementations as
// drivers, including Logrus and Zerolog, along with support for
// reporting services including Sentry.
package log

import (
	"io"
)

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

// UnderlyingLogger is an escape hatch allowing Loggers registered with
// this package the option to return their underlying implementation, as
// well as reset it.
//
// NOTE(IB): this is currently required for CustomOptions to work.
type UnderlyingLogger interface {
	GetLogger() interface{}
	SetLogger(interface{})
}
