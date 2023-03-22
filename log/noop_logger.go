package log

import (
	"io"
	"time"
)

// Register logger.
func init() {
	Register("noop", func(_ *Config, _ ...Option) (Logger, error) {
		return Noop(), nil
	})
}

// Noop instantiates a new Logger that does nothing when its methods
// are called. This can also be retrieved using "noop" with Open;
// keep in mind though that any configuration is ignored.
// Noop loggers can be used for testing or to satisfy a logger
// expectation without actually writing any logs.
func Noop() Logger {
	return noopLogger{}
}

// Logger implementation.

type noopLogger struct{}

var _ Logger = (*noopLogger)(nil)

func (noopLogger) LogLevel() Level { return INFO }

func (noopLogger) Print(_ ...any)                    {}
func (noopLogger) Printf(_ string, _ ...any)         {}
func (noopLogger) WithError(_ error) Entry           { return noopEntry{} }
func (noopLogger) WithField(_ string, _ any) Entry   { return noopEntry{} }
func (noopLogger) WithFields(_ map[string]any) Entry { return noopEntry{} }

func (noopLogger) Entry(_ Level) Entry { return noopEntry{} }
func (noopLogger) Trace() Entry        { return noopEntry{} }
func (noopLogger) Debug() Entry        { return noopEntry{} }
func (noopLogger) Info() Entry         { return noopEntry{} }
func (noopLogger) Warn() Entry         { return noopEntry{} }
func (noopLogger) Error() Entry        { return noopEntry{} }
func (noopLogger) Panic() Entry        { return noopEntry{} }
func (noopLogger) Fatal() Entry        { return noopEntry{} }

func (n noopLogger) WriteCloser(_ Level) io.WriteCloser { return n }

// WriteCloser implementation.

func (noopLogger) Write(p []byte) (int, error) { return len(p), nil }
func (noopLogger) Close() error                { return nil }

// Entry implementation.

type noopEntry struct{}

var _ Entry = (*noopEntry)(nil)

func (n noopEntry) Async() Entry                               { return n }
func (n noopEntry) IsLevelEnabled(level Level) bool            { return level >= INFO }
func (n noopEntry) Caller(_ ...int) Entry                      { return n }
func (n noopEntry) WithError(_ ...error) Entry                 { return n }
func (n noopEntry) WithField(_ string, _ any) Entry            { return n }
func (n noopEntry) WithFields(_ map[string]any) Entry          { return n }
func (n noopEntry) WithBool(_ string, _ ...bool) Entry         { return n }
func (n noopEntry) WithDur(_ string, _ ...time.Duration) Entry { return n }
func (n noopEntry) WithInt(_ string, _ ...int) Entry           { return n }
func (n noopEntry) WithUint(_ string, _ ...uint) Entry         { return n }
func (n noopEntry) WithStr(_ string, _ ...string) Entry        { return n }
func (n noopEntry) WithTime(_ string, _ ...time.Time) Entry    { return n }

func (n noopEntry) Trace() Entry { return n }
func (n noopEntry) Debug() Entry { return n }
func (n noopEntry) Info() Entry  { return n }
func (n noopEntry) Warn() Entry  { return n }
func (n noopEntry) Error() Entry { return n }
func (n noopEntry) Panic() Entry { return n }
func (n noopEntry) Fatal() Entry { return n }

func (noopEntry) Msgf(_ string, _ ...any) {}
func (noopEntry) Msg(_ any)               {}
func (noopEntry) Send()                   {}
