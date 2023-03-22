package log

import "context"

// Avoid collisions by using package-scoped types for context keys.
type ctxKey int

// Context keys for storing and retrieving loggers and entries in
// contexts.
const (
	// LoggerKey is the key value to use with context.Context for Logger
	// put and retrieval.
	LoggerKey ctxKey = iota + 1

	// EntryKey is the key value to use with context.Context for Logger
	// put and retrieval.
	EntryKey
)

// ContextWithLogger returns a context with Logger l as its value.
func ContextWithLogger(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, l)
}

// LoggerFromContext returns the Logger in ctx, or nil if none exists.
func LoggerFromContext(ctx context.Context) Logger {
	l, _ := ctx.Value(LoggerKey).(Logger)
	return l
}

// ContextWithEntry returns a context with Entry e as its value.
func ContextWithEntry(ctx context.Context, e Entry) context.Context {
	return context.WithValue(ctx, EntryKey, e)
}

// EntryFromContext returns the Entry in ctx, or nil if none exists.
func EntryFromContext(ctx context.Context) Entry {
	e, _ := ctx.Value(EntryKey).(Entry)
	return e
}
