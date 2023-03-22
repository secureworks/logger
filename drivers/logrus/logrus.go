// Package logrus implements a logger with a Logrus driver. See the
// documentation associated with the Logger, Entry and UnderlyingLogger
// interfaces for their respective methods.
package logrus

import (
	"github.com/sirupsen/logrus"

	"github.com/secureworks/logger/log"
)

func init() {
	log.Register("logrus", newLogger)
}

func levelToLogrusLevel(level log.Level) logrus.Level {
	switch level {
	case log.TRACE:
		return logrus.TraceLevel
	case log.DEBUG:
		return logrus.DebugLevel
	case log.INFO:
		return logrus.InfoLevel
	case log.WARN:
		return logrus.WarnLevel
	case log.ERROR:
		return logrus.ErrorLevel
	case log.PANIC:
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

func logrusLevelToLevel(level logrus.Level) log.Level {
	switch level {
	case logrus.TraceLevel:
		return log.TRACE
	case logrus.DebugLevel:
		return log.DEBUG
	case logrus.InfoLevel:
		return log.INFO
	case logrus.WarnLevel:
		return log.WARN
	case logrus.ErrorLevel:
		return log.ERROR
	case logrus.PanicLevel:
		return log.PANIC
	default:
		return log.INFO
	}
}
