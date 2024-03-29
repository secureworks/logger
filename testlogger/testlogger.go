// Package testlogger implements a logger with a testing driver. See the
// documentation associated with the Logger and Entry interfaces for
// their respective methods.
//
// Test loggers can be used for simple testing and mocking in an obvious
// way.
//
package testlogger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/secureworks/logger/log"
)

// Register logger.
func init() {
	log.Register("test", func(config *log.Config, opts ...log.Option) (log.Logger, error) {
		return New(config, opts...)
	})
}

// New instantiates a new *testlogger.Logger that can be used for
// testing. This can also be retrieved using "test" with log.Open and
// asserting the type, eg:
//
//     l, _ := log.Open("test", nil)
//     logger, _ := l.(*testlogger.Logger)
//
func New(config *log.Config, opts ...log.Option) (*Logger, error) {
	if config == nil {
		// Env no-op ensures we don't have settings based on env.
		config = log.DefaultConfig(func(string) string { return "" })
	}

	logger := &Logger{
		Config:            config,
		WriteCloserBuffer: &bytes.Buffer{}, // For WriteCloser.
		ExitFn:            os.Exit,
		entries:           []*Entry{},
		// multiple go routines can append to entries at the same time, so we will
		// use this mutux to lock any access made to the entries field
		entriesMutex: sync.Mutex{},
	}

	// Change default output, as long as os.Stdout (for examples) is not set.
	if logger.Config.Output != os.Stdout {
		logger.Config.Output = &bytes.Buffer{}
	}

	for _, opt := range opts {
		if err := opt(logger); err != nil {
			return nil, err
		}
	}
	return logger, nil
}

// MustNew instantiates a new *testlogger.Logger that can be used for
// testing. It will panic if there are any initialization errors.
func MustNew(config *log.Config, opts ...log.Option) *Logger {
	logger, err := New(config, opts...)
	if err != nil {
		panic(err)
	}
	return logger
}

// Logger implementation.

// Logger is public so that we can cast the log.Logger interface to it
// and access its fields.
type Logger struct {
	// Config provides access to the current log.Config. This is useful
	// for testing / ensuring configuration from the environment is being
	// set correctly. Also, Config.Output is by default a buffer that
	// can be asserted against.
	Config *log.Config

	// WriteCloserBuffer is the buffer that is wrapped by the WriteCloser
	// and can be asserted against in testing.
	WriteCloserBuffer *bytes.Buffer

	// ExitFn is called when sending a log.FATAL entry. It defaults to
	// os.Exit.
	ExitFn func(int)

	// entries holds a list of all the entries generated by the logger,
	// so that we can make assertions against them.
	entries []*Entry

	entriesMutex sync.Mutex

	underlyingLoggerValue interface{}
}

var _ log.Logger = (*Logger)(nil)

// GetEntries can be used to the logs that have been posted up to the start of program or since
// last call to GetEntries (which ever is most recent)
// to call this method, you will need to cast the logger to testlogger.Logger
func (l *Logger) GetEntries() []*Entry {
	l.entriesMutex.Lock()
	defer l.entriesMutex.Unlock()
	rtn := l.entries
	l.entries = []*Entry{}
	return rtn
}

func (l *Logger) IsLevelEnabled(lvl log.Level) bool {
	return lvl >= l.Config.Level
}

func (l *Logger) Entry(lvl log.Level) log.Entry {
	entry := &Entry{
		Logger: l,
		Level:  lvl,
		Fields: make(map[string]interface{}),
	}
	l.entriesMutex.Lock()
	defer l.entriesMutex.Unlock()
	l.entries = append(l.entries, entry)
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

func (l *Logger) Write(p []byte) (int, error) { return l.WriteCloserBuffer.Write(p) }
func (l *Logger) Close() error                { return nil }

// GetLogger will return the value set in SetLogger.
func (l *Logger) GetLogger() interface{} {
	return l.underlyingLoggerValue
}

// SetLogger will set the value that is retrieved in GetLogger.
func (l *Logger) SetLogger(v interface{}) {
	l.underlyingLoggerValue = v
}

// Entry implementation.

// Entry is public so that we can cast the log.Entry interface to it
// and access its fields and the test utility methods: HasField, Field,
// StringField, and the Request… helpers.
type Entry struct {
	// Logger is a reference to the testlogger.Logger that generated the
	// entry.
	Logger *Logger

	// Level is the current level of the entry.
	Level log.Level

	// Fields is a map of all the fields set using With… methods.
	Fields map[string]interface{}

	// IsAsync holds the the current "async" state of the entry.
	IsAsync bool

	// Sent is true if the message has been "sent," ie written to the
	// output buffer.
	Sent bool

	// Message stores the message field.
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

func (e *Entry) Caller(vals ...int) log.Entry {
	return e.WithField(log.CallerField, vals)
}

func (e *Entry) WithError(errs ...error) log.Entry {
	if len(errs) == 0 {
		return e
	}
	if len(errs) == 1 {
		return e.WithField("error", errs[0].Error())
	}
	return e.WithField("error", errs)
}

func (e *Entry) WithBool(k string, vals ...bool) log.Entry {
	if len(vals) == 0 {
		return e
	}
	if len(vals) == 1 {
		return e.WithField(k, vals[0])
	}
	return e.WithField(k, vals)
}

func (e *Entry) WithDur(k string, vals ...time.Duration) log.Entry {
	if len(vals) == 0 {
		return e
	}
	if len(vals) == 1 {
		return e.WithField(k, vals[0])
	}
	return e.WithField(k, vals)
}

func (e *Entry) WithInt(k string, vals ...int) log.Entry {
	if len(vals) == 0 {
		return e
	}
	if len(vals) == 1 {
		return e.WithField(k, vals[0])
	}
	return e.WithField(k, vals)
}

func (e *Entry) WithUint(k string, vals ...uint) log.Entry {
	if len(vals) == 0 {
		return e
	}
	if len(vals) == 1 {
		return e.WithField(k, vals[0])
	}
	return e.WithField(k, vals)
}

func (e *Entry) WithStr(k string, vals ...string) log.Entry {
	if len(vals) == 0 {
		return e
	}
	if len(vals) == 1 {
		return e.WithField(k, vals[0])
	}
	return e.WithField(k, vals)
}

func (e *Entry) WithTime(k string, vals ...time.Time) log.Entry {
	if len(vals) == 0 {
		return e
	}
	if len(vals) == 1 {
		return e.WithField(k, vals[0])
	}
	return e.WithField(k, vals)
}

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
		e.Send()
	}
}

func (e *Entry) Msgf(format string, vals ...interface{}) {
	e.Msg(fmt.Sprintf(format, vals...))
}

// Send writes a JSON version of the fields with any message and the
// level.
func (e *Entry) Send() {
	fields := e.Fields
	fields["level"] = StringFromLevel(e.Level)
	if e.Message != "" {
		fields["message"] = e.Message
	}

	byt, err := json.Marshal(fields)
	if err == nil {
		e.Logger.entriesMutex.Lock()
		_, _ = e.Logger.Config.Output.Write(byt)
		e.Logger.entriesMutex.Unlock()
		e.Sent = true
	}

	switch e.Level {
	case log.FATAL:
		e.Logger.ExitFn(1)
	case log.PANIC:
		panic(&testloggerError{Entry: e, msg: string(byt)})
	}
}

type testloggerError struct {
	*Entry
	msg string
}

func (e *testloggerError) Error() string {
	return e.msg
}

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
		if val, ok := e.Field(name).(string); ok {
			return val
		}
		return ""
	}
	return strings.Join(val, ";")
}

// RequestDuration returns the stored request duration field, if any
// exists. Any value less than zero indicates a missing or malformed
// field. Assumes that only one duration was set.
func (e *Entry) RequestDuration() time.Duration {
	val, ok := e.Field(log.ReqDuration).(string)
	if !ok {
		return -2
	}
	dur, err := time.ParseDuration(val)
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

// StringFromLevel is a convenience for printing out log.Levels.
func StringFromLevel(lvl log.Level) string {
	switch lvl {
	case log.TRACE:
		return "TRACE"
	case log.DEBUG:
		return "DEBUG"
	case log.INFO:
		return "INFO"
	case log.WARN:
		return "WARN"
	case log.ERROR:
		return "ERROR"
	case log.PANIC:
		return "PANIC"
	case log.FATAL:
		return "FATAL"
	}
	return "UNKN"
}
