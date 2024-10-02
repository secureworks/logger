package common

import "github.com/secureworks/errors"

// StackTracer defines a common interface for extracting stack
// information. Sentry and several other packages use this interface.
type StackTracer interface {
	StackTrace() errors.Frames
}

type stackTracer struct {
	err    error
	frames errors.Frames
}

func (st stackTracer) Error() string {
	return st.err.Error()
}

func (st stackTracer) Unwrap() error {
	return st.err
}

func (st stackTracer) Frames() errors.Frames {
	return st.frames
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
func WithStackTrace(err error, skipFrames int) (StackTracer, error) {
	if err == nil {
		return nil, nil
	}

	var frames errors.Frames
	if framer, ok := err.(interface{ Frames() errors.Frames }); ok {
		frames = framer.Frames()
	} else {
		frames = errors.CallStackAt(skipFrames)
	}
	st := stackTracer{err: err, frames: frames}
	return st, st
}
