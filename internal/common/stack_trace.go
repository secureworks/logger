package common

import "github.com/secureworks/errors"

// StackTracer defines a common interface for extracting stack
// information. Sentry and several other packages use this interface.
type StackTracer interface {
	StackTrace() errors.Frames
}

type stackTracer struct {
	frames errors.Frames
}

func (st stackTracer) StackTrace() errors.Frames {
	return st.frames
}

// WithStackTrace ensures that an error has a stack trace, and pairs it
// with a StackTracer.
//
// It checks if the given error implements StackTracer and returns it if
// it does. Otherwise, it will wrap the error in an error type that
// implements StackTracer and returns both.
//
// If nil is passed then nil is returned.
func WithStackTrace(err error) (StackTracer, error) {
	if err == nil {
		return nil, nil
	}
	framer, ok := err.(interface{ Frames() errors.Frames })
	if ok {
		return stackTracer{framer.Frames()}, nil
	}
	err = errors.WithStackTrace(err)
	return stackTracer{errors.FramesFrom(err)}, err
}
