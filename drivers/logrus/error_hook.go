package logrus

import (
	"github.com/sirupsen/logrus"

	"github.com/secureworks/logger/log"
)

// Implements a Logrus hook (https://github.com/sirupsen/logrus#hooks)
// to add a stack trace to the logging event.
//
// QUESTION(IB): Is this type necessary? There are tradeoffs doing it in
// the event versus a hook.
type errorHook struct{}

// Levels implements the Logrus Hook interface.
//
// Runs on all levels.
func (errorHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire ... TODO(PH)
//
// If the event does not have a stack trace field and an error that
// implements StackTracer, put the error's stack trace in the stack
// trace field.
func (errorHook) Fire(event *logrus.Entry) error {
	if _, ok := event.Data[log.StackField]; ok {
		return nil
	}
	st, ok := event.Data[logrus.ErrorKey].(interface {
		StackTrace() []uint
	})
	if !ok {
		return nil
	}

	event.Data[log.StackField] = st.StackTrace()
	return nil
}
