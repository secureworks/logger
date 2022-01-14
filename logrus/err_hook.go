package logrus

import (
	"github.com/sirupsen/logrus"

	"github.com/secureworks/logger/log"
	"github.com/secureworks/logger/log/internal/common"
)

// TODO(IB): Is this type necessary? There are tradeoffs doing it in the
// event versus a hook.
type errHook struct{}

func (errHook) Levels() []logrus.Level { return logrus.AllLevels }

func (errHook) Fire(event *logrus.Entry) (err error) {
	// QUESTION(IB): Already has a stack?
	if _, ok := event.Data[log.StackField]; ok {
		return
	}

	// QUESTION(IB): Doesn't have a StackTracer?
	st, ok := event.Data[logrus.ErrorKey].(common.StackTracer)
	if !ok {
		return
	}

	event.Data[log.StackField] = st.StackTrace()
	return
}
