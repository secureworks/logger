package log

import "strings"

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
	}

	// Default case isn't needed, default is determined by enum zero
	// value.
	return
}

// IsValid checks if the current level is valid relative to known
// values.
func (l Level) IsValid() bool {
	return l >= TRACE && l <= PANIC
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
	}
}
