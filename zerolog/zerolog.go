// Package zerolog implements a logger with a Zerolog driver. See the
// documentation associated with the Logger, Entry and UnderlyingLogger
// interfaces for their respective methods.
package zerolog

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/rs/zerolog"

	"github.com/secureworks/logger/internal/common"
	"github.com/secureworks/logger/log"
)

var skipLogFrameCount = zerolog.CallerSkipFrameCount + 1

//go:noinline
func errorStackMarshaler(err error) interface{} {
	if err == nil {
		return nil
	}
	st, _ := common.WithStackTrace(err, skipLogFrameCount)
	return st.StackTrace()
}

// Register logger.
func init() {
	// These are package vars in Zerolog so putting them here is less race-y
	// than setting them in newLogger.
	zerolog.ErrorStackFieldName = log.StackField
	zerolog.ErrorStackMarshaler = errorStackMarshaler

	log.Register("zerolog", newLogger)
}

// newLogger instantiates a new log.Logger with a Zerolog driver using
// the given configuration and Zerolog options.
func newLogger(config *log.Config, opts ...log.Option) (log.Logger, error) {
	zlvl := lvlToZerolog(config.Level)
	logger := &logger{
		errStack: config.EnableErrStack,
		lvl:      zlvl,
	}

	output := config.Output
	if output == nil {
		output = os.Stderr
	}

	zlog := zerolog.New(output).Level(zlvl)
	logger.lg = &zlog

	// Apply options.
	for _, opt := range opts {
		if err := opt(logger); err != nil {
			return nil, err
		}
	}
	return logger, nil
}

// Logger implementation.

type logger struct {
	lg       *zerolog.Logger
	lvl      zerolog.Level
	errStack bool
}

var _ log.Logger = (*logger)(nil)
var _ log.UnderlyingLogger = (*logger)(nil)

func (l *logger) IsLevelEnabled(lvl log.Level) bool {
	return lvlToZerolog(lvl) <= l.lvl
}

func (l *logger) WithError(err error) log.Entry {
	return l.Error().WithError(err)
}

func (l *logger) WithField(key string, val interface{}) log.Entry {
	return l.Entry(0).WithField(key, val)
}

func (l *logger) WithFields(fields map[string]interface{}) log.Entry {
	return l.Entry(0).WithFields(fields)
}

func (l *logger) Entry(lvl log.Level) log.Entry {
	return l.newEntry(lvlToZerolog(lvl))
}

func (l *logger) Trace() log.Entry { return l.newEntry(zerolog.TraceLevel) }
func (l *logger) Debug() log.Entry { return l.newEntry(zerolog.DebugLevel) }
func (l *logger) Info() log.Entry  { return l.newEntry(zerolog.InfoLevel) }
func (l *logger) Warn() log.Entry  { return l.newEntry(zerolog.WarnLevel) }
func (l *logger) Error() log.Entry { return l.newEntry(zerolog.ErrorLevel) }
func (l *logger) Panic() log.Entry { return l.newEntry(zerolog.PanicLevel) }
func (l *logger) Fatal() log.Entry { return l.newEntry(zerolog.FatalLevel) }

func (l *logger) WriteCloser(lvl log.Level) io.WriteCloser {
	return writeLevelCloser{log: l, lvl: lvl}
}

// UnderlyingLogger implementation.

func (l *logger) GetLogger() interface{} {
	if l.notValid() {
		return nil
	}
	return l.lg
}

func (l *logger) SetLogger(iface interface{}) {
	if lg, ok := iface.(*zerolog.Logger); ok && !l.notValid() {
		l.lg = lg
	}
	if lg, ok := iface.(zerolog.Logger); ok && !l.notValid() {
		l.lg = &lg
	}
}

// Zerolog-specific methods.

// DisabledEntry is an assertable method/interface if someone wants to
// disable zerolog events at runtime.
func (l *logger) DisabledEntry() log.Entry {
	return (*entry)(nil)
}

// Logger utility functions.

// Creates a new entry at the given level.
func (l *logger) newEntry(lvl zerolog.Level) log.Entry {
	if l.notValid() {
		return l.DisabledEntry()
	}

	// We have to use NoLevel or we can't change them after the fact:
	//   - https://github.com/rs/zerolog/blob/7825d863376faee2723fc99c061c538bd80812c8/log.go#L419
	//   - https://github.com/rs/zerolog/pull/255
	//
	// Our own *entry type will write the level as needed.
	//
	// See: https://github.com/rs/zerolog/issues/408
	ent := l.lg.WithLevel(zerolog.NoLevel)
	if l.errStack {
		ent = ent.Stack()
	}

	return &entry{
		ent:    ent,
		caller: make([]string, 0, 1),
		loglvl: l.lvl,
		lvl:    lvl,
	}
}

func (l *logger) notValid() bool {
	return l == nil || l.lg == nil
}

// Map log.Level to internal Zerolog log levels.
func lvlToZerolog(lvl log.Level) zerolog.Level {
	switch lvl {
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

// WriteCloser hook implementation.

type writeLevelCloser struct {
	log log.Logger
	lvl log.Level
}

// An implementation of zerolog.Logger.Write method but with level
// support.
func (wlc writeLevelCloser) Write(p []byte) (n int, err error) {
	n = len(p)
	if n > 0 && p[n-1] == '\n' {
		// Trim CR added by stdlog.
		p = p[0 : n-1]
	}
	wlc.log.Entry(wlc.lvl).Msg(string(p))
	return
}

func (wlc writeLevelCloser) Close() error {
	return nil
}

// Entry implementation.

type entry struct {
	ent    *zerolog.Event
	caller []string
	msg    string
	async  bool
	loglvl zerolog.Level
	lvl    zerolog.Level
}

var _ log.Entry = (*entry)(nil)
var _ log.UnderlyingLogger = (*entry)(nil)

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

func (e *entry) WithError(errs ...error) log.Entry {
	le := len(errs)
	if e.notValid() || le == 0 {
		return e
	}

	if le == 1 {
		e.ent = e.ent.Err(errs[0])
	} else {
		e.ent = e.ent.Errs(zerolog.ErrorFieldName, errs)
	}
	return e
}

func (e *entry) WithField(key string, val interface{}) log.Entry {
	if e.notValid() {
		return e
	}
	e.ent = e.ent.Interface(key, val)
	return e
}

func (e *entry) WithFields(fields map[string]interface{}) log.Entry {
	if e.notValid() || len(fields) == 0 {
		return e
	}
	e.ent = e.ent.Fields(fields)
	return e
}

func (e *entry) WithBool(key string, bls ...bool) log.Entry {
	lb := len(bls)
	if e.notValid() || lb == 0 {
		return e
	}

	if lb == 1 {
		e.ent = e.ent.Bool(key, bls[0])
	} else {
		e.ent = e.ent.Bools(key, bls)
	}
	return e
}

func (e *entry) WithDur(key string, durs ...time.Duration) log.Entry {
	ld := len(durs)
	if e.notValid() || ld == 0 {
		return e
	}

	if ld == 1 {
		e.ent = e.ent.Dur(key, durs[0])
	} else {
		e.ent = e.ent.Durs(key, durs)
	}
	return e
}

func (e *entry) WithInt(key string, is ...int) log.Entry {
	li := len(is)
	if e.notValid() || li == 0 {
		return e
	}

	if li == 1 {
		e.ent = e.ent.Int(key, is[0])
	} else {
		e.ent = e.ent.Ints(key, is)
	}
	return e
}

func (e *entry) WithUint(key string, us ...uint) log.Entry {
	lu := len(us)
	if e.notValid() || lu == 0 {
		return e
	}

	if lu == 1 {
		e.ent = e.ent.Uint(key, us[0])
	} else {
		e.ent = e.ent.Uints(key, us)
	}
	return e
}

func (e *entry) WithStr(key string, strs ...string) log.Entry {
	ls := len(strs)
	if e.notValid() || ls == 0 {
		return e
	}

	if ls == 1 {
		e.ent = e.ent.Str(key, strs[0])
	} else {
		e.ent = e.ent.Strs(key, strs)
	}
	return e
}

func (e *entry) WithTime(key string, ts ...time.Time) log.Entry {
	lt := len(ts)
	if e.notValid() || lt == 0 {
		return e
	}

	if lt == 1 {
		e.ent = e.ent.Time(key, ts[0])
	} else {
		e.ent = e.ent.Times(key, ts)
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

func (e *entry) Msg(msg string) {
	if e.notValid() {
		return
	}

	e.msg = msg
	if !e.async {
		e.Send()
	}
}

func (e *entry) Send() {
	if !e.enabled() {
		// If we cut out early && the entry is valid, recycle it.
		if !e.notValid() {
			putEvent(e.ent)
			e.ent = nil
		}

		return
	}

	// Nil out zerolog.Entry as we're done with it. Mostly helps gc and
	// disables future method calls on this type.
	defer func() { e.ent = nil }()

	if len(e.caller) > 0 {
		e.ent = e.ent.Strs(log.CallerField, e.caller)
	}
	e.ent = e.ent.Str(zerolog.LevelFieldName, zerolog.LevelFieldMarshalFunc(e.lvl))

	changeEventLevel(e.ent, e.lvl) // Change the level if we can, before calling Msg.
	e.ent.Msg(e.msg)               // Recycles the zerolog.Entry for us (do not call putEvent again).

	// These are called by e.done here:
	//   - https://github.com/rs/zerolog/blob/791ca15d999a97768ffd3b040116f9f5a772661a/event.go
	//
	// They are disabled however by our use of 'NoLevel', so we retain the
	// functions here.
	//
	switch e.lvl {
	case zerolog.PanicLevel:
		panic(e.msg)
	case zerolog.FatalLevel:
		os.Exit(1)
	}
}

// UnderlyingLogger implementation.

func (e *entry) GetLogger() interface{} {
	if e.notValid() {
		return nil
	}

	return e.ent
}

func (e *entry) SetLogger(l interface{}) {
	if ent, ok := l.(*zerolog.Event); ok && !e.notValid() {
		e.ent = ent
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
	if e.ent != nil {
		putEvent(e.ent)
		e.ent = nil
	}

	return e
}

// Entry utility functions.

func (e *entry) notValid() bool {
	return e == nil || e.ent == nil
}

func (e *entry) enabled() bool {
	return !e.notValid() && e.lvl >= e.loglvl
}

func (e *entry) setLevel(lvl zerolog.Level) log.Entry {
	if e.notValid() {
		return e
	}
	e.lvl = lvl
	return e
}
