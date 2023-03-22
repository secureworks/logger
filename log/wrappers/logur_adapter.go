package wrappers

import (
	"context"
	"errors"

	"github.com/secureworks/logger/log"
)

// LogurAdapter is a Logur adapter for Logger.
type LogurAdapter struct {
	logger log.Logger
}

// NewLogurAdapter returns a new Logur logger. The given Logger may not
// be nil.
func NewLogurAdapter(logger log.Logger) (*LogurAdapter, error) {
	if logger == nil {
		return nil, errors.New("TODO(PH)")
	}
	return &LogurAdapter{logger: logger}, nil
}

func (l *LogurAdapter) Trace(msg string, fields ...map[string]any) {
	if l.logger.LogLevel() < log.TRACE {
		return
	}
	entry := l.logger.Entry(log.TRACE)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Msg(msg)
}

func (l *LogurAdapter) Debug(msg string, fields ...map[string]any) {
	if l.logger.LogLevel() < log.DEBUG {
		return
	}
	entry := l.logger.Entry(log.DEBUG)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Msg(msg)
}

func (l *LogurAdapter) Info(msg string, fields ...map[string]any) {
	if l.logger.LogLevel() < log.INFO {
		return
	}
	entry := l.logger.Entry(log.INFO)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Msg(msg)
}

func (l *LogurAdapter) Warn(msg string, fields ...map[string]any) {
	if l.logger.LogLevel() < log.WARN {
		return
	}
	entry := l.logger.Entry(log.WARN)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Msg(msg)
}

func (l *LogurAdapter) Error(msg string, fields ...map[string]any) {
	if l.logger.LogLevel() < log.ERROR {
		return
	}
	entry := l.logger.Entry(log.ERROR)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Msg(msg)
}

func (l *LogurAdapter) TraceContext(_ context.Context, msg string, fields ...map[string]any) {
	l.Trace(msg, fields...)
}

func (l *LogurAdapter) DebugContext(_ context.Context, msg string, fields ...map[string]any) {
	l.Debug(msg, fields...)
}

func (l *LogurAdapter) InfoContext(_ context.Context, msg string, fields ...map[string]any) {
	l.Info(msg, fields...)
}

func (l *LogurAdapter) WarnContext(_ context.Context, msg string, fields ...map[string]any) {
	l.Warn(msg, fields...)
}

func (l *LogurAdapter) ErrorContext(_ context.Context, msg string, fields ...map[string]any) {
	l.Error(msg, fields...)
}
