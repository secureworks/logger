package logrus

import (
	"github.com/sirupsen/logrus"

	"github.com/secureworks/logger/internal/common"
	"github.com/secureworks/logger/log"
)

// Implements a Logrus hook (https://github.com/sirupsen/logrus#hooks)
// to add a stack trace to the logging event.
type errorHook struct{}

// Levels returns on all Logrus levels (the levels this hook is
// triggered on).
func (errorHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

// Fire runs the hook. If the event does not have a stack trace field
// and an error that implements StackTracer, put the error's stack trace
// in the stack trace field.
func (errorHook) Fire(event *logrus.Entry) error {
	if _, ok := event.Data[log.StackField]; ok {
		return nil
	}
	st, ok := event.Data[logrus.ErrorKey].(common.StackTracer)
	if !ok {
		return nil
	}

	event.Data[log.StackField] = st.StackTrace()
	return nil
}
