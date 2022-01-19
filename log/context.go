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

// CtxWithLogger returns a context with Logger l as its value.
func CtxWithLogger(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, l)
}

// LoggerFromCtx returns the Logger in ctx, or nil if none exists.
func LoggerFromCtx(ctx context.Context) Logger {
	l, _ := ctx.Value(LoggerKey).(Logger)
	return l
}

// CtxWithEntry returns a context with Entry e as its value.
func CtxWithEntry(ctx context.Context, e Entry) context.Context {
	return context.WithValue(ctx, EntryKey, e)
}

// EntryFromCtx returns the Entry in ctx, or nil if none exists.
func EntryFromCtx(ctx context.Context) Entry {
	e, _ := ctx.Value(EntryKey).(Entry)
	return e
}
