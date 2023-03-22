// Package zerolog implements a log.Logger with a Zerolog driver. See
// the documentation associated with the log.Logger, log.Entry and
// log.UnderlyingLogger interfaces for their respective methods.
package zerolog

import (
	"time"

	"github.com/rs/zerolog"

	"github.com/secureworks/errors"
	"github.com/secureworks/logger/log"
)

func init() {

	// These are package vars in Zerolog so putting them here is less race-y
	// than setting them in newLogger.
	zerolog.ErrorStackFieldName = log.StackField
	zerolog.ErrorStackMarshaler = func(err error) any {
		if ff := errors.FramesFrom(err); len(ff) > 0 {
			return ff
		}
		return errors.CallStackAt(2)
	}

	// Register logger.
	log.Register("zerolog", newLogger)
}

func levelToZerologLevel(level log.Level) zerolog.Level {
	switch level {
	case log.TRACE:
		return zerolog.TraceLevel
	case log.DEBUG:
		return zerolog.DebugLevel
	case log.INFO:
		return zerolog.InfoLevel
	case log.WARN:
		return zerolog.WarnLevel
	case log.ERROR:
		return zerolog.ErrorLevel
	case log.PANIC:
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

func zerologLevelToLevel(level zerolog.Level) log.Level {
	switch level {
	case zerolog.TraceLevel:
		return log.TRACE
	case zerolog.DebugLevel:
		return log.DEBUG
	case zerolog.InfoLevel:
		return log.INFO
	case zerolog.WarnLevel:
		return log.WARN
	case zerolog.ErrorLevel:
		return log.ERROR
	case zerolog.PanicLevel:
		return log.PANIC
	default:
		return log.INFO
	}
}

type zerologEntryable interface {
	*zerolog.Event | zerolog.Context
}

type zerologEntry[T zerologEntryable] interface {
	Err(error) zerologEntry[T]
	Errs(string, []error) zerologEntry[T]
	Msg(string)
	Interface(string, any) zerologEntry[T]
	Fields(map[string]any) zerologEntry[T]
	Str(string, string) zerologEntry[T]
	Strs(string, []string) zerologEntry[T]
	Bool(string, bool) zerologEntry[T]
	Bools(string, []bool) zerologEntry[T]
	Time(string, time.Time) zerologEntry[T]
	Times(string, []time.Time) zerologEntry[T]
	Dur(string, time.Duration) zerologEntry[T]
	Durs(string, []time.Duration) zerologEntry[T]
	Int(string, int) zerologEntry[T]
	Ints(string, []int) zerologEntry[T]
	Uint(string, uint) zerologEntry[T]
	Uints(string, []uint) zerologEntry[T]
}
