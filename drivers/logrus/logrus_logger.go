package logrus

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/secureworks/logger/log"
)

type logger struct {
	logger   *logrus.Logger
	errStack bool
}

var _ interface {
	log.Logger
	log.UnderlyingLogger
} = (*logger)(nil)

// newLogger instantiates a new log.Logger with a Logrus driver using
// the given configuration and Logrus options.
func newLogger(config *log.Config, opts ...log.Option) (log.Logger, error) {
	logrusLogger := logrus.New()

	if config.Output == nil {
		config.Output = os.Stderr
	}
	logrusLogger.SetOutput(config.Output)
	logrusLogger.SetLevel(levelToLogrusLevel(config.Level))
	logrusLogger.SetNoLock()

	if config.Format == log.JSONFormat {
		jsonF := &logrus.JSONFormatter{
			PrettyPrint: config.LocalDevel,
		}
		logrusLogger.SetFormatter(jsonF)
	}

	if config.EnableErrStack {
		logrusLogger.AddHook(errorHook{})
	}

	l := &logger{logger: logrusLogger, errStack: config.EnableErrStack}
	for _, opt := range opts {
		if err := opt(l); err != nil {
			return nil, err
		}
	}
	return l, nil
}

func (l *logger) IsLevelEnabled(level log.Level) bool {
	return l.logger.IsLevelEnabled(levelToLogrusLevel(level))
}

func (l *logger) LogLevel() log.Level {
	return logrusLevelToLevel(l.logger.Level)
}

func (l *logger) Print(v ...any) {
	l.Entry(l.LogLevel()).Msg(v)
}

func (l *logger) Printf(format string, v ...any) {
	l.Entry(l.LogLevel()).Msgf(format, v...)
}

func (l *logger) WithError(err error) log.Entry {
	entry := l.Error()
	return entry.WithError(err)
}

func (l *logger) WithField(key string, val any) log.Entry {
	entry := l.Entry(0)
	return entry.WithField(key, val)
}

func (l *logger) WithFields(fields map[string]any) log.Entry {
	entry := l.Entry(0)
	return entry.WithFields(fields)
}

func (l *logger) Entry(level log.Level) log.Entry {
	return l.newEntry(levelToLogrusLevel(level))
}

func (l *logger) Trace() log.Entry { return l.newEntry(logrus.TraceLevel) }
func (l *logger) Debug() log.Entry { return l.newEntry(logrus.DebugLevel) }
func (l *logger) Info() log.Entry  { return l.newEntry(logrus.InfoLevel) }
func (l *logger) Warn() log.Entry  { return l.newEntry(logrus.WarnLevel) }
func (l *logger) Error() log.Entry { return l.newEntry(logrus.ErrorLevel) }
func (l *logger) Panic() log.Entry { return l.newEntry(logrus.PanicLevel) }
func (l *logger) Fatal() log.Entry { return l.newEntry(logrus.FatalLevel) }

func (l *logger) WriteCloser(level log.Level) io.WriteCloser {
	return l.logger.WriterLevel(levelToLogrusLevel(level))
}

func (l *logger) GetLogger() any {
	return l.logger
}

func (l *logger) SetLogger(i any) {
	if lgLogger, ok := i.(*logrus.Logger); ok {
		l.logger = lgLogger
	}
	if lgLogger, ok := i.(logrus.Logger); ok {
		l.logger = &lgLogger
	}
}

func (l *logger) newEntry(level logrus.Level) *entry {
	return &entry{
		entry:       logrus.NewEntry(l.logger),
		errStack:    l.errStack,
		loggerLevel: l.logger.Level,
		level:       level,
	}
}
