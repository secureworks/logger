package log

import "io"

// Logger is the minimum interface loggers should implement when used
// with CTPx packages.
type Logger interface {

	// TODO(PH)
	LogLevel() Level

	// WriteCloser returns an io.Writer that, when written to, writes logs
	// at the given level. It is the callers responsibility to call Close
	// when finished. This is particularly useful for redirecting the output
	// of other loggers or even Readers with the help of io.TeeReader.
	WriteCloser(Level) io.WriteCloser

	StandardLogger
	LevelLogger
}

type StandardLogger interface {

	// TODO(PH)
	Print(v ...any)

	// TODO(PH)
	Printf(format string, v ...any)
}

type LevelLogger interface {

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
}

// UnderlyingLogger is an escape hatch allowing Loggers registered with
// this package the option to return their underlying implementation, as
// well as reset it.
//
// This is currently required for CustomOptions to work.
type UnderlyingLogger interface {

	// TODO(PH)
	GetLogger() any

	// TODO(PH)
	SetLogger(any)
}
