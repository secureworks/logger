package zerolog

import (
	"io"
	"os"

	"github.com/rs/zerolog"

	"github.com/secureworks/logger/log"
)

type logger struct {
	logger          *zerolog.Logger
	level           zerolog.Level
	reusableEntries bool
	errStack        bool
}

var _ interface {
	log.Logger
	log.UnderlyingLogger
} = (*logger)(nil)

// newLogger instantiates a new log.Logger with a Zerolog driver using
// the given configuration and Zerolog options.
func newLogger(config *log.Config, opts ...log.Option) (log.Logger, error) {
	level := levelToZerologLevel(config.Level)
	l := &logger{
		errStack: config.EnableErrStack,
		level:    level,
	}

	output := config.Output
	if output == nil {
		output = os.Stderr
	}

	zlog := zerolog.New(output).Level(level)
	l.logger = &zlog

	// Apply options.
	for _, opt := range opts {
		if err := opt(l); err != nil {
			return nil, err
		}
	}
	return l, nil
}

func (l *logger) IsLevelEnabled(level log.Level) bool {
	return levelToZerologLevel(level) <= l.level
}

func (l *logger) LogLevel() log.Level {
	return zerologLevelToLevel(l.level)
}

func (l *logger) Print(v ...any) {
	l.Entry(l.LogLevel()).Msg(v)
}

func (l *logger) Printf(format string, v ...any) {
	l.Entry(l.LogLevel()).Msgf(format, v...)
}

func (l *logger) Entry(level log.Level) log.Entry {
	return l.newEntry(levelToZerologLevel(level))
}

func (l *logger) Trace() log.Entry { return l.newEntry(zerolog.TraceLevel) }
func (l *logger) Debug() log.Entry { return l.newEntry(zerolog.DebugLevel) }
func (l *logger) Info() log.Entry  { return l.newEntry(zerolog.InfoLevel) }
func (l *logger) Warn() log.Entry  { return l.newEntry(zerolog.WarnLevel) }
func (l *logger) Error() log.Entry { return l.newEntry(zerolog.ErrorLevel) }
func (l *logger) Panic() log.Entry { return l.newEntry(zerolog.PanicLevel) }
func (l *logger) Fatal() log.Entry { return l.newEntry(zerolog.FatalLevel) }

func (l *logger) WriteCloser(level log.Level) io.WriteCloser {
	return writeLevelCloser{log: l, level: level}
}

func (l *logger) GetLogger() any {
	if l.notValid() {
		return nil
	}
	return l.logger
}

func (l *logger) SetLogger(i any) {
	if zLogger, ok := i.(*zerolog.Logger); ok && !l.notValid() {
		l.logger = zLogger
	}
	if zLogger, ok := i.(zerolog.Logger); ok && !l.notValid() {
		l.logger = &zLogger
	}
}

// Zerolog-specific methods.

// DisabledEntry is an assertable method/interface if someone wants to
// disable zerolog events at runtime.
func (l *logger) DisabledEntry() log.Entry {
	return (*entry[*zerolog.Event])(nil)
}

func (l *logger) newEntry(level zerolog.Level) log.Entry {
	if l.notValid() {
		return l.DisabledEntry()
	}

	if l.reusableEntries {
		zContext := l.logger.With()
		if l.errStack {
			zContext = zContext.Stack()
		}
		return &entry[zerolog.Context]{
			entry:       any(zContext).(zerologEntry[zerolog.Context]),
			caller:      make([]string, 0, 1),
			loggerLevel: l.level,
			level:       level,
		}
	}

	// We have to use NoLevel, or we can't change them after the fact:
	//   - https://github.com/rs/zerolog/blob/7825d863376faee2723fc99c061c538bd80812c8/log.go#L419
	//   - https://github.com/rs/zerolog/pull/255
	//
	// Our own *entry type will write the level as needed.
	//
	// See: https://github.com/rs/zerolog/issues/408
	zEvent := l.logger.WithLevel(zerolog.NoLevel)
	if l.errStack {
		zEvent = zEvent.Stack()
	}
	return &entry[*zerolog.Event]{
		entry:       any(zEvent).(zerologEntry[*zerolog.Event]),
		caller:      make([]string, 0, 1),
		loggerLevel: l.level,
		level:       level,
	}
}

func (l *logger) notValid() bool {
	return l == nil || l.logger == nil
}

// writeLevelCloser is the WriteCloser hook implementation.
type writeLevelCloser struct {
	log   log.Logger
	level log.Level
}

// Write is an implementation of the zerolog.Logger.Write method but
// with level support.
func (wlc writeLevelCloser) Write(p []byte) (n int, err error) {
	n = len(p)
	if n > 0 && p[n-1] == '\n' {
		// Trim CR added by stdlog.
		p = p[0 : n-1]
	}
	wlc.log.Entry(wlc.level).Msg(string(p))
	return
}

func (wlc writeLevelCloser) Close() error {
	return nil
}
