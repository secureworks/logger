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

func (np noopLogger) WriteCloser(_ log.Level) io.WriteCloser { return np }
func (noopLogger) Write(p []byte) (int, error)               { return len(p), nil }
func (noopLogger) Close() error                              { return nil }

// Entry implementation.

type noopEntry struct{}

var _ log.Entry = (*noopEntry)(nil)

func (np noopEntry) Async() log.Entry          { return np }
func (np noopEntry) Caller(_ ...int) log.Entry { return np }

func (np noopEntry) WithError(_ ...error) log.Entry                 { return np }
func (np noopEntry) WithField(_ string, _ interface{}) log.Entry    { return np }
func (np noopEntry) WithFields(_ map[string]interface{}) log.Entry  { return np }
func (np noopEntry) WithBool(_ string, _ ...bool) log.Entry         { return np }
func (np noopEntry) WithDur(_ string, _ ...time.Duration) log.Entry { return np }
func (np noopEntry) WithInt(_ string, _ ...int) log.Entry           { return np }
func (np noopEntry) WithUint(_ string, _ ...uint) log.Entry         { return np }
func (np noopEntry) WithStr(_ string, _ ...string) log.Entry        { return np }
func (np noopEntry) WithTime(_ string, _ ...time.Time) log.Entry    { return np }

func (np noopEntry) Trace() log.Entry { return np }
func (np noopEntry) Debug() log.Entry { return np }
func (np noopEntry) Info() log.Entry  { return np }
func (np noopEntry) Warn() log.Entry  { return np }
func (np noopEntry) Error() log.Entry { return np }
func (np noopEntry) Panic() log.Entry { return np }
func (np noopEntry) Fatal() log.Entry { return np }

func (noopEntry) Msgf(_ string, _ ...interface{}) {}
func (noopEntry) Msg(_ string)                    {}
func (noopEntry) Send()                           {}
