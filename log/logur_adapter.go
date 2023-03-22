package log

import (
	"context"
	"errors"
)

// LogurAdapter is a Logur adapter for Logger.
type LogurAdapter struct {
	logger Logger
}

// NewLogurAdapter returns a new Logur logger. The given Logger may not
// be nil.
func NewLogurAdapter(logger Logger) (*LogurAdapter, error) {
	if logger == nil {
		return nil, errors.New("TODO(PH)")
	}
	return &LogurAdapter{logger: logger}, nil
}

func (l *LogurAdapter) Trace(msg string, fields ...map[string]any) {
	if l.logger.LogLevel() < TRACE {
		return
	}
	entry := l.logger.Entry(TRACE)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Msg(msg)
}

func (l *LogurAdapter) Debug(msg string, fields ...map[string]any) {
	if l.logger.LogLevel() < DEBUG {
		return
	}
	entry := l.logger.Entry(DEBUG)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Msg(msg)
}

func (l *LogurAdapter) Info(msg string, fields ...map[string]any) {
	if l.logger.LogLevel() < INFO {
		return
	}
	entry := l.logger.Entry(INFO)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Msg(msg)
}

func (l *LogurAdapter) Warn(msg string, fields ...map[string]any) {
	if l.logger.LogLevel() < WARN {
		return
	}
	entry := l.logger.Entry(WARN)
	if len(fields) > 0 {
		entry = entry.WithFields(fields[0])
	}
	entry.Msg(msg)
}

func (l *LogurAdapter) Error(msg string, fields ...map[string]any) {
	if l.logger.LogLevel() < ERROR {
		return
	}
	entry := l.logger.Entry(ERROR)
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
