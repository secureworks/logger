package logrus

import (
	"fmt"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/secureworks/errors"
	"github.com/secureworks/logger/log"
)

type entry struct {
	entry       *logrus.Entry
	errStack    bool
	message     string
	async       bool
	loggerLevel logrus.Level
	level       logrus.Level
}

var _ interface {
	log.Entry
	log.UnderlyingLogger
} = (*entry)(nil)

func (e *entry) Async() log.Entry {
	e.async = !e.async
	return e
}

func (e *entry) IsLevelEnabled(level log.Level) bool {
	return levelToLogrusLevel(level) <= e.loggerLevel
}

func (e *entry) Caller(skip ...int) log.Entry {
	if e == nil {
		return e
	}

	sk := 1
	if len(skip) > 0 {
		sk += skip[0]
	}

	_, file, line, ok := runtime.Caller(sk)
	if !ok {
		return e
	}

	// Not normal Logrus: append to existing field; nil won't panic.
	cls, _ := e.entry.Data[log.CallerField].([]string)
	cls = append(cls, fmt.Sprintf("%s:%d", file, line))
	e.entry.Data[log.CallerField] = cls

	return e
}

func (e *entry) WithError(errs ...error) log.Entry {
	if len(errs) == 0 || e == nil {
		return e
	}

	err := errors.NewMultiError(errs...).ErrorOrNil()
	if e.errStack && len(errors.FramesFrom(err)) == 0 {
		err = errors.WithFrames(err, errors.CallStackAt(1))
	}
	return e.WithField(logrus.ErrorKey, err)
}

func (e *entry) WithField(key string, val any) log.Entry {
	// The deferred functions args are eval'd when defer is called not
	// when the deferred function is run.
	defer releaseEntry(e.entry.Logger, e.entry)
	e.entry = e.entry.WithField(key, val)
	return e
}

func (e *entry) WithFields(fields map[string]any) log.Entry {
	defer releaseEntry(e.entry.Logger, e.entry)
	e.entry = e.entry.WithFields(fields)
	return e
}

func (e *entry) WithBool(key string, bls ...bool) log.Entry {
	if e == nil || len(bls) == 0 {
		return e
	}

	var i any = bls[0]
	if len(bls) > 1 {
		i = bls
	}
	return e.WithField(key, i)
}

func (e *entry) WithDur(key string, durs ...time.Duration) log.Entry {
	if e == nil || len(durs) == 0 {
		return e
	}

	var i any = durs[0]
	if len(durs) > 1 {
		i = durs
	}
	return e.WithField(key, i)
}

func (e *entry) WithInt(key string, is ...int) log.Entry {
	if e == nil || len(is) == 0 {
		return e
	}

	var i any = is[0]
	if len(is) > 1 {
		i = is
	}
	return e.WithField(key, i)
}

func (e *entry) WithUint(key string, us ...uint) log.Entry {
	if e == nil || len(us) == 0 {
		return e
	}

	var i any = us[0]
	if len(us) > 1 {
		i = us
	}
	return e.WithField(key, i)
}

func (e *entry) WithStr(key string, strs ...string) log.Entry {
	if e == nil || len(strs) == 0 {
		return e
	}

	// String allocates when placed into empty interface 🙁.
	var i any = strs[0]
	if len(strs) > 1 {
		i = strs
	}
	return e.WithField(key, i)
}

func (e *entry) WithTime(key string, ts ...time.Time) log.Entry {
	if e == nil || len(ts) == 0 {
		return e
	}

	var i any = ts[0]
	if len(ts) > 1 {
		i = ts
	}

	// Avoid using WithTime here from Logrus as we don't want to
	// unnecessarily override time value.
	return e.WithField(key, i)
}

func (e *entry) Trace() log.Entry { e.level = logrus.TraceLevel; return e }
func (e *entry) Debug() log.Entry { e.level = logrus.DebugLevel; return e }
func (e *entry) Info() log.Entry  { e.level = logrus.InfoLevel; return e }
func (e *entry) Warn() log.Entry  { e.level = logrus.WarnLevel; return e }
func (e *entry) Error() log.Entry { e.level = logrus.ErrorLevel; return e }
func (e *entry) Panic() log.Entry { e.level = logrus.PanicLevel; return e }
func (e *entry) Fatal() log.Entry { e.level = logrus.FatalLevel; return e }

func (e *entry) Msgf(format string, v ...any) {
	e.Msg(fmt.Sprintf(format, v...))
}

func (e *entry) Msg(v any) {
	e.message = fmt.Sprintf("%v", v)
	if !e.async {
		e.Send()
	}
}

func (e *entry) Send() {
	if e == nil || e.entry == nil {
		return
	}

	defer releaseEntry(e.entry.Logger, e.entry)

	switch e.level {
	case logrus.PanicLevel:
		e.entry.Panic(e.message)
	case logrus.FatalLevel:
		e.entry.Fatal(e.message)
	default:
		e.entry.Log(e.level, e.message)
	}

	e.entry = nil
}

// UnderlyingLogger implementation.

func (e *entry) GetLogger() any {
	return e.entry
}

func (e *entry) SetLogger(l any) {
	if ent, ok := l.(*logrus.Entry); ok {
		e.entry = ent
	}
}
