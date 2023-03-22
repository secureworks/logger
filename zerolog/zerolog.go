// Package zerolog implements a log.Logger with a Zerolog driver. See
// the documentation associated with the log.Logger, log.Entry and
// log.UnderlyingLogger interfaces for their respective methods.
package zerolog

import (
	"github.com/rs/zerolog"

	"github.com/secureworks/logger/internal/common"
	"github.com/secureworks/logger/log"
)

func init() {

	// These are package vars in Zerolog so putting them here is less race-y
	// than setting them in newLogger.
	zerolog.ErrorStackFieldName = log.StackField
	zerolog.ErrorStackMarshaler = func(err error) interface{} {
		st, _ := common.WithStackTrace(err)
		return st.StackTrace()
	}

	// Register logger.
	log.Register("zerolog", newLogger)
}

func levelToZerolog(level log.Level) zerolog.Level {
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
	case log.FATAL:
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}
