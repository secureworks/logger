package log

import "time"

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

	// Caller embeds a caller value into the existing Entry. A caller value
	// is a filepath followed by line number. Skip determines the number of
	// additional stack frames to ascend when determining the value. By
	// default, the caller of the method is the value used and skip does not
	// need to be supplied in that case. Caller may be called multiple times
	// on the Entry to build a stack or execution trace.
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
