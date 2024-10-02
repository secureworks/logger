package logrus

import (
	"github.com/sirupsen/logrus"

	"github.com/secureworks/logger/internal/common"
	"github.com/secureworks/logger/log"
)

// errorHook implements a Logrus hook
// (https://github.com/sirupsen/logrus#hooks) to add a stack trace to
// the logging event.
//
// QUESTION(IB): Is this type necessary? There are tradeoffs doing it in
// the event versus a hook.
type errorHook struct{}

// Levels ensures the hook runs on all levels.
func (errorHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire ensures that if the event does not have a stack trace field and
// an error that implements StackTracer, put the error's stack trace in
// the stack trace field.
func (errorHook) Fire(entry *logrus.Entry) error {
	if _, ok := entry.Data[log.StackField]; ok {
		return nil
	}
	st, ok := entry.Data[logrus.ErrorKey].(common.StackTracer)
	if !ok {
		return nil
	}

	entry.Data[log.StackField] = st.StackTrace()
	return nil
}
