package zerolog

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog"

	"github.com/secureworks/logger/log"
)

type entry struct {
	entry       *zerolog.Event
	caller      []string
	message     string
	async       bool
	loggerLevel zerolog.Level
	level       zerolog.Level
}

var _ interface {
	log.Entry
	log.UnderlyingLogger
} = (*entry)(nil)

func (e *entry) Async() log.Entry {
	if e.notValid() {
		return e
	}
	e.async = !e.async
	return e
}

func (e *entry) Caller(skip ...int) log.Entry {
	if e.notValid() {
		return e
	}

	sk := 1
	if len(skip) > 0 {
		sk += skip[0]
	}

	// Originally it was planned to use zerolog's Caller implementation but
	// the interface was changed during the design phase to allow the Caller
	// to be called multiple times which zerolog won't do without adding dup
	// fields.
	_, file, line, ok := runtime.Caller(sk)
	if !ok {
		return e
	}

	// TODO(IB): Use zerolog.CallerMarshalFunc?
	e.caller = append(e.caller, fmt.Sprintf("%s:%d", file, line))
	return e
}

func (e *entry) WithError(errs ...error) log.Entry {
	le := len(errs)
	if e.notValid() || le == 0 {
		return e
	}

	if le == 1 {
		e.entry = e.entry.Err(errs[0])
	} else {
		e.entry = e.entry.Errs(zerolog.ErrorFieldName, errs)
	}
	return e
}

func (e *entry) WithField(key string, val interface{}) log.Entry {
	if e.notValid() {
		return e
	}
	e.entry = e.entry.Interface(key, val)
	return e
}

func (e *entry) WithFields(fields map[string]interface{}) log.Entry {
	if e.notValid() || len(fields) == 0 {
		return e
	}
	e.entry = e.entry.Fields(fields)
	return e
}

func (e *entry) WithBool(key string, bls ...bool) log.Entry {
	lb := len(bls)
	if e.notValid() || lb == 0 {
		return e
	}

	if lb == 1 {
		e.entry = e.entry.Bool(key, bls[0])
	} else {
		e.entry = e.entry.Bools(key, bls)
	}
	return e
}

func (e *entry) WithDur(key string, durs ...time.Duration) log.Entry {
	ld := len(durs)
	if e.notValid() || ld == 0 {
		return e
	}

	if ld == 1 {
		e.entry = e.entry.Dur(key, durs[0])
	} else {
		e.entry = e.entry.Durs(key, durs)
	}
	return e
}

func (e *entry) WithInt(key string, is ...int) log.Entry {
	li := len(is)
	if e.notValid() || li == 0 {
		return e
	}

	if li == 1 {
		e.entry = e.entry.Int(key, is[0])
	} else {
		e.entry = e.entry.Ints(key, is)
	}
	return e
}

func (e *entry) WithUint(key string, us ...uint) log.Entry {
	lu := len(us)
	if e.notValid() || lu == 0 {
		return e
	}

	if lu == 1 {
		e.entry = e.entry.Uint(key, us[0])
	} else {
		e.entry = e.entry.Uints(key, us)
	}
	return e
}

func (e *entry) WithStr(key string, strs ...string) log.Entry {
	ls := len(strs)
	if e.notValid() || ls == 0 {
		return e
	}

	if ls == 1 {
		e.entry = e.entry.Str(key, strs[0])
	} else {
		e.entry = e.entry.Strs(key, strs)
	}
	return e
}

func (e *entry) WithTime(key string, ts ...time.Time) log.Entry {
	lt := len(ts)
	if e.notValid() || lt == 0 {
		return e
	}

	if lt == 1 {
		e.entry = e.entry.Time(key, ts[0])
	} else {
		e.entry = e.entry.Times(key, ts)
	}
	return e
}

func (e *entry) Trace() log.Entry { return e.setLevel(zerolog.TraceLevel) }
func (e *entry) Debug() log.Entry { return e.setLevel(zerolog.DebugLevel) }
func (e *entry) Info() log.Entry  { return e.setLevel(zerolog.InfoLevel) }
func (e *entry) Warn() log.Entry  { return e.setLevel(zerolog.WarnLevel) }
func (e *entry) Error() log.Entry { return e.setLevel(zerolog.ErrorLevel) }
func (e *entry) Panic() log.Entry { return e.setLevel(zerolog.PanicLevel) }
func (e *entry) Fatal() log.Entry { return e.setLevel(zerolog.FatalLevel) }

func (e *entry) Msgf(format string, vals ...interface{}) {
	e.Msg(fmt.Sprintf(format, vals...))
}

func (e *entry) Msg(message string) {
	if e.notValid() {
		return
	}

	e.message = message
	if !e.async {
		e.Send()
	}
}

func (e *entry) Send() {
	if !e.enabled() {
		// If we cut out early && the entry is valid, recycle it.
		if !e.notValid() {
			putEvent(e.entry)
			e.entry = nil
		}

		return
	}

	// Nil out zerolog.Entry as we're done with it. Mostly helps gc and
	// disables future method calls on this type.
	defer func() { e.entry = nil }()

	if len(e.caller) > 0 {
		e.entry = e.entry.Strs(log.CallerField, e.caller)
	}
	e.entry = e.entry.Str(zerolog.LevelFieldName, zerolog.LevelFieldMarshalFunc(e.level))

	changeEventLevel(e.entry, e.level) // Change the level if we can, before calling Msg.
	e.entry.Msg(e.message)             // Recycles the zerolog.Entry for us (do not call putEvent again).

	// These are called by e.done here:
	//   - https://github.com/rs/zerolog/blob/791ca15d999a97768ffd3b040116f9f5a772661a/event.go
	//
	// They are disabled however by our use of 'NoLevel', so we retain the
	// functions here.
	//
	switch e.level {
	case zerolog.PanicLevel:
		panic(e.message)
	case zerolog.FatalLevel:
		os.Exit(1)
	}
}

func (e *entry) GetLogger() interface{} {
	if e.notValid() {
		return nil
	}

	return e.entry
}

func (e *entry) SetLogger(l interface{}) {
	if ent, ok := l.(*zerolog.Event); ok && !e.notValid() {
		e.entry = ent
	}
}

// Zerolog-specific methods.

// DisabledEntry is an assertable method/interface if someone wants to
// disable zerolog events at runtime.
func (e *entry) DisabledEntry() log.Entry {
	if e.notValid() {
		return e
	}

	// This will disable all other methods.
	if e.entry != nil {
		putEvent(e.entry)
		e.entry = nil
	}

	return e
}

// Entry utility functions.

func (e *entry) notValid() bool {
	return e == nil || e.entry == nil
}

func (e *entry) enabled() bool {
	return !e.notValid() && e.level >= e.loggerLevel
}

func (e *entry) setLevel(level zerolog.Level) log.Entry {
	if e.notValid() {
		return e
	}
	e.level = level
	return e
}
