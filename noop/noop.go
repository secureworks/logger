// Package noop implements a unified logger without a driver. See the
// documentation associated with the Logger and Entry interfaces for
// their respective methods.
//
// Noop loggers can be used for testing or to satisfy a logger
// expectation without actually writing any logs.
//
package noop

import (
	"io"
	"time"

	"github.com/secureworks/logger/log"
)

// Register logger.
func init() {
	log.Register("noop", func(_ *log.Config, _ ...log.Option) (log.Logger, error) {
		return New(), nil
	})
}

// New returns an implementation of Logger that does nothing when its
// methods are called. This can also be retrieved using "noop" with
// "Open", though in that case any configuration is ignored.
func New() log.Logger {
	return noopLogger{}
}

// Logger implementation.

type noopLogger struct{}

var _ log.Logger = (*noopLogger)(nil)

func (noopLogger) IsLevelEnabled(lvl log.Level) bool             { return false }
func (noopLogger) WithError(_ error) log.Entry                   { return noopEntry{} }
func (noopLogger) WithField(_ string, _ interface{}) log.Entry   { return noopEntry{} }
func (noopLogger) WithFields(_ map[string]interface{}) log.Entry { return noopEntry{} }

func (noopLogger) Entry(_ log.Level) log.Entry { return noopEntry{} }
func (noopLogger) Trace() log.Entry            { return noopEntry{} }
func (noopLogger) Debug() log.Entry            { return noopEntry{} }
func (noopLogger) Info() log.Entry             { return noopEntry{} }
func (noopLogger) Warn() log.Entry             { return noopEntry{} }
func (noopLogger) Error() log.Entry            { return noopEntry{} }
func (noopLogger) Panic() log.Entry            { return noopEntry{} }
func (noopLogger) Fatal() log.Entry            { return noopEntry{} }

func (n noopLogger) WriteCloser(_ log.Level) io.WriteCloser { return n }

// WriteCloser implementation.

func (noopLogger) Write(p []byte) (int, error) { return len(p), nil }
func (noopLogger) Close() error                { return nil }

// Entry implementation.

type noopEntry struct{}

var _ log.Entry = (*noopEntry)(nil)

func (n noopEntry) Async() log.Entry          { return n }
func (n noopEntry) Caller(_ ...int) log.Entry { return n }

func (n noopEntry) WithError(_ ...error) log.Entry                 { return n }
func (n noopEntry) WithField(_ string, _ interface{}) log.Entry    { return n }
func (n noopEntry) WithFields(_ map[string]interface{}) log.Entry  { return n }
func (n noopEntry) WithBool(_ string, _ ...bool) log.Entry         { return n }
func (n noopEntry) WithDur(_ string, _ ...time.Duration) log.Entry { return n }
func (n noopEntry) WithInt(_ string, _ ...int) log.Entry           { return n }
func (n noopEntry) WithUint(_ string, _ ...uint) log.Entry         { return n }
func (n noopEntry) WithStr(_ string, _ ...string) log.Entry        { return n }
func (n noopEntry) WithTime(_ string, _ ...time.Time) log.Entry    { return n }

func (n noopEntry) Trace() log.Entry { return n }
func (n noopEntry) Debug() log.Entry { return n }
func (n noopEntry) Info() log.Entry  { return n }
func (n noopEntry) Warn() log.Entry  { return n }
func (n noopEntry) Error() log.Entry { return n }
func (n noopEntry) Panic() log.Entry { return n }
func (n noopEntry) Fatal() log.Entry { return n }

func (noopEntry) Msgf(_ string, _ ...interface{}) {}
func (noopEntry) Msg(_ string)                    {}
func (noopEntry) Send()                           {}
