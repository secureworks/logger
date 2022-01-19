// Package testlogger implements a unified logger without a driver. See
// the documentation associated with the Logger and Entry interfaces for
// their respective methods.
//
// Test loggers can be used for simple testing and mocking in an obvious
// way.
//
package testlogger

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/secureworks/logger/log"
)

// Register logger.
func init() {
	log.Register("test", func(config *log.Config, opts ...log.Option) (log.Logger, error) {
		return New(config, opts...)
	})
}

// New instantiates a new log.Logger that does nothing when its methods
// are called. This can also be retrieved using "test" with log.Open.
func New(config *log.Config, opts ...log.Option) (log.Logger, error) {
	if config == nil {
		// Env no-op ensures we don't have settings based on env.
		config = log.DefaultConfig(func(string) string { return "" })
	}
	logger := &Logger{
		Config:  config,
		Buffer:  &bytes.Buffer{},
		Entries: []*Entry{},
	}
	for _, opt := range opts {
		if err := opt(logger); err != nil {
			return nil, err
		}
	}
	return logger, nil
}

// Logger implementation.

type Logger struct {
	Config  *log.Config
	Buffer  *bytes.Buffer
	Entries []*Entry
}

var _ log.Logger = (*Logger)(nil)

func (l *Logger) IsLevelEnabled(lvl log.Level) bool {
	return lvl >= l.Config.Level
}

func (l *Logger) Entry(lvl log.Level) log.Entry {
	entry := &Entry{
		Logger: l,
		Level:  lvl,
		Fields: make(map[string]interface{}),
	}
	l.Entries = append(l.Entries, entry)
	return entry
}

func (l *Logger) Trace() log.Entry { return l.Entry(log.TRACE) }
func (l *Logger) Debug() log.Entry { return l.Entry(log.DEBUG) }
func (l *Logger) Info() log.Entry  { return l.Entry(log.INFO) }
func (l *Logger) Warn() log.Entry  { return l.Entry(log.WARN) }
func (l *Logger) Error() log.Entry { return l.Entry(log.ERROR) }
func (l *Logger) Panic() log.Entry { return l.Entry(log.PANIC) }
func (l *Logger) Fatal() log.Entry { return l.Entry(log.FATAL) }

func (l *Logger) WithError(err error) log.Entry {
	return l.Entry(l.Config.Level).WithError(err)
}

func (l *Logger) WithField(k string, val interface{}) log.Entry {
	e, _ := l.Entry(l.Config.Level).(*Entry)
	e.Fields[k] = val
	return e
}

func (l *Logger) WithFields(fields map[string]interface{}) log.Entry {
	e := l.Entry(l.Config.Level)
	for k, val := range fields {
		e.WithField(k, val)
	}
	return e
}

func (l *Logger) WriteCloser(_ log.Level) io.WriteCloser { return l }

// WriteCloser implementation.

func (l *Logger) Write(p []byte) (int, error) { return l.Buffer.Write(p) }
func (Logger) Close() error                   { return nil }

// Entry implementation.

type Entry struct {
	Logger  log.Logger
	Level   log.Level
	Fields  map[string]interface{}
	IsAsync bool
	Sent    bool
	Message string
}

var _ log.Entry = (*Entry)(nil)

func (e *Entry) Async() log.Entry { e.IsAsync = !e.IsAsync; return e }

func (e *Entry) WithField(k string, val interface{}) log.Entry {
	e.Fields[k] = val
	return e
}

func (e *Entry) WithFields(fields map[string]interface{}) log.Entry {
	for k, val := range fields {
		e.WithField(k, val)
	}
	return e
}

func (e *Entry) Caller(vals ...int) log.Entry                      { return e.WithField(log.CallerField, vals) }
func (e *Entry) WithError(errs ...error) log.Entry                 { return e.WithField(log.StackField, errs) }
func (e *Entry) WithBool(k string, vals ...bool) log.Entry         { return e.WithField(k, vals) }
func (e *Entry) WithDur(k string, vals ...time.Duration) log.Entry { return e.WithField(k, vals) }
func (e *Entry) WithInt(k string, vals ...int) log.Entry           { return e.WithField(k, vals) }
func (e *Entry) WithUint(k string, vals ...uint) log.Entry         { return e.WithField(k, vals) }
func (e *Entry) WithStr(k string, vals ...string) log.Entry        { return e.WithField(k, vals) }
func (e *Entry) WithTime(k string, vals ...time.Time) log.Entry    { return e.WithField(k, vals) }

func (e *Entry) Trace() log.Entry { e.Level = log.TRACE; return e }
func (e *Entry) Debug() log.Entry { e.Level = log.DEBUG; return e }
func (e *Entry) Info() log.Entry  { e.Level = log.INFO; return e }
func (e *Entry) Warn() log.Entry  { e.Level = log.WARN; return e }
func (e *Entry) Error() log.Entry { e.Level = log.ERROR; return e }
func (e *Entry) Panic() log.Entry { e.Level = log.PANIC; return e }
func (e *Entry) Fatal() log.Entry { e.Level = log.FATAL; return e }

func (e *Entry) Msg(msg string) {
	e.Message = msg
	if !e.IsAsync {
		e.Sent = true
	}
}

func (e *Entry) Msgf(format string, vals ...interface{}) { e.Msg(fmt.Sprintf(format, vals...)) }
func (e *Entry) Send()                                   { e.Sent = true }

// Test utilities.

// HasField returns if a given field has been set in the entry.
func (e *Entry) HasField(name string) bool {
	_, ok := e.Fields[name]
	return ok
}

// Field returns the value stored at a given field, or nil if none
// exists.
func (e *Entry) Field(name string) interface{} {
	return e.Fields[name]
}

// StringField returns a simple string for a given field; empty if no
// field exists, joining multiple values with ";".
func (e *Entry) StringField(name string) string {
	val, ok := e.Field(name).([]string)
	if !ok {
		return ""
	}
	return strings.Join(val, ";")
}

// RequestDuration returns the stored request duration field, if any
// exists. Any value less than zero indicates a missing or malformed
// field.
func (e *Entry) RequestDuration() time.Duration {
	val, ok := e.Field(log.ReqDuration).([]string)
	if !ok {
		return -3
	}
	if len(val) != 1 {
		return -2
	}
	dur, err := time.ParseDuration(val[0])
	if err != nil {
		return -1
	}
	return dur
}

// RequestMethod returns the stored request method field, if any exists.
func (e *Entry) RequestMethod() string {
	return e.StringField(log.ReqMethod)
}

// RequestPath returns the stored request path field, if any exists.
func (e *Entry) RequestPath() string {
	return e.StringField(log.ReqPath)
}

// RequestRemoteAddr returns the stored request remote address field, if
// any exists.
func (e *Entry) RequestRemoteAddr() string {
	return e.StringField(log.ReqRemoteAddr)
}