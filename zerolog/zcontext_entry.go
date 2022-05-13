package zerolog

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog"
	"github.com/secureworks/logger/log"
)

type zcontext struct {
	zctx   zerolog.Context
	caller []string
	msg    string
	async  bool
	loglvl zerolog.Level
	lvl    zerolog.Level
}

var _ log.Entry = (*zcontext)(nil)

func (e *zcontext) Async() log.Entry {
	if e.notValid() {
		return e
	}
	e.async = !e.async
	return e
}

func (e *zcontext) Caller(skip ...int) log.Entry {
	if e.notValid() {
		return e
	}

	sk := 1
	if len(skip) > 0 {
		sk += skip[0]
	}

	// Originally it was planned to use zerologs Caller implementation
	// but the interface was changed during the design phase to allow the
	// Caller to be called multiple times which zerolog won't do without
	// adding dup fields.
	_, file, line, ok := runtime.Caller(sk)
	if !ok {
		return e
	}

	// TODO(IB): Use zerolog.CallerMarshalFunc?
	e.caller = append(e.caller, fmt.Sprintf("%s:%d", file, line))
	return e
}

func (e *zcontext) WithError(errs ...error) log.Entry {
	le := len(errs)
	if e.notValid() || le == 0 {
		return e
	}

	if le == 1 {
		e.zctx = e.zctx.Err(errs[0])
	} else {
		e.zctx = e.zctx.Errs(zerolog.ErrorFieldName, errs)
	}
	return e
}

func (e *zcontext) WithField(key string, val interface{}) log.Entry {
	if e.notValid() {
		return e
	}
	e.zctx = e.zctx.Interface(key, val)
	return e
}

func (e *zcontext) WithFields(fields map[string]interface{}) log.Entry {
	if e.notValid() || len(fields) == 0 {
		return e
	}
	e.zctx = e.zctx.Fields(fields)
	return e
}

func (e *zcontext) WithBool(key string, bls ...bool) log.Entry {
	lb := len(bls)
	if e.notValid() || lb == 0 {
		return e
	}

	if lb == 1 {
		e.zctx = e.zctx.Bool(key, bls[0])
	} else {
		e.zctx = e.zctx.Bools(key, bls)
	}
	return e
}

func (e *zcontext) WithDur(key string, durs ...time.Duration) log.Entry {
	ld := len(durs)
	if e.notValid() || ld == 0 {
		return e
	}

	if ld == 1 {
		e.zctx = e.zctx.Dur(key, durs[0])
	} else {
		e.zctx = e.zctx.Durs(key, durs)
	}
	return e
}

func (e *zcontext) WithInt(key string, is ...int) log.Entry {
	li := len(is)
	if e.notValid() || li == 0 {
		return e
	}

	if li == 1 {
		e.zctx = e.zctx.Int(key, is[0])
	} else {
		e.zctx = e.zctx.Ints(key, is)
	}
	return e
}

func (e *zcontext) WithUint(key string, us ...uint) log.Entry {
	lu := len(us)
	if e.notValid() || lu == 0 {
		return e
	}

	if lu == 1 {
		e.zctx = e.zctx.Uint(key, us[0])
	} else {
		e.zctx = e.zctx.Uints(key, us)
	}
	return e
}

func (e *zcontext) WithStr(key string, strs ...string) log.Entry {
	ls := len(strs)
	if e.notValid() || ls == 0 {
		return e
	}

	if ls == 1 {
		e.zctx = e.zctx.Str(key, strs[0])
	} else {
		e.zctx = e.zctx.Strs(key, strs)
	}
	return e
}

func (e *zcontext) WithTime(key string, ts ...time.Time) log.Entry {
	lt := len(ts)
	if e.notValid() || lt == 0 {
		return e
	}

	if lt == 1 {
		e.zctx = e.zctx.Time(key, ts[0])
	} else {
		e.zctx = e.zctx.Times(key, ts)
	}
	return e
}

func (e *zcontext) Trace() log.Entry { return e.setLevel(zerolog.TraceLevel) }
func (e *zcontext) Debug() log.Entry { return e.setLevel(zerolog.DebugLevel) }
func (e *zcontext) Info() log.Entry  { return e.setLevel(zerolog.InfoLevel) }
func (e *zcontext) Warn() log.Entry  { return e.setLevel(zerolog.WarnLevel) }
func (e *zcontext) Error() log.Entry { return e.setLevel(zerolog.ErrorLevel) }
func (e *zcontext) Panic() log.Entry { return e.setLevel(zerolog.PanicLevel) }
func (e *zcontext) Fatal() log.Entry { return e.setLevel(zerolog.FatalLevel) }

func (e *zcontext) Msgf(format string, vals ...interface{}) {
	e.Msg(fmt.Sprintf(format, vals...))
}

func (e *zcontext) Msg(msg string) {
	if e.notValid() {
		return
	}

	e.msg = msg
	if !e.async {
		e.Send()
	}
}

func (e *zcontext) Send() {
	if !e.enabled() {
		return
	}

	zl := e.zctx.Logger()
	ent := zl.WithLevel(e.lvl)

	if len(e.caller) > 0 {
		ent = ent.Strs(log.CallerField, e.caller)
	}

	ent.Msg(e.msg)

	// These are called by e.done here:
	//   - https://github.com/rs/zerolog/blob/791ca15d999a97768ffd3b040116f9f5a772661a/event.go
	//
	// They are disabled however by our use of 'WithLevel'/'NoLevel', so we retain the
	// functions here.
	//
	switch e.lvl {
	case zerolog.PanicLevel:
		panic(e.msg)
	case zerolog.FatalLevel:
		os.Exit(1)
	}
	e.msg = ""
}

// UnderlyingLogger implementation.

func (e *zcontext) GetLogger() interface{} {
	if e.notValid() {
		return nil
	}

	return e.zctx
}

func (e *zcontext) SetLogger(l interface{}) {
	if zctx, ok := l.(zerolog.Context); ok && !e.notValid() {
		e.zctx = zctx
	}
}

// Zerolog-specific methods.

// Entry utility functions.

func (e *zcontext) notValid() bool {
	return e == nil
}

func (e *zcontext) enabled() bool {
	return !e.notValid() && e.lvl >= e.loglvl
}

func (e *zcontext) setLevel(lvl zerolog.Level) log.Entry {
	if e.notValid() {
		return e
	}
	e.lvl = lvl
	return e
}
