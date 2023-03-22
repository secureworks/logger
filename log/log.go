// Package log provides the unified interface for the Secureworks
// logger. This interface can use underlying logger implementations as
// drivers, including Logrus and Zerolog, along with support for
// reporting services including Sentry.
package log

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
