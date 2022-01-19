package common

import (
	"fmt"
	"path/filepath"
	"sync/atomic"

	"github.com/getsentry/sentry-go"
)

var nonReentrant int32

// InitSentry is a shared function for initializing the global "hub"
// object that manages scopes and clients for Sentry.
//
// Unfortunately while Sentry supports several hubs, the default and
// most commonly used one is "CurrentHub", a package level variable that
// cannot be set up concurrently (it is data race free but not race
// condition free). Ideally each logger instance would have their own
// hub, but users will likely expect this to set up Sentry "in total".
// For now just do what we can so that multiple logger instances use the
// same thing if desired.
//
// QUESTION(PH): why not use sync.Once?
func InitSentry(opts sentry.ClientOptions) error {
	var err error
	if !atomic.CompareAndSwapInt32(&nonReentrant, 0, 1) {
		return nil
	}
	defer func() {
		if err != nil {
			atomic.StoreInt32(&nonReentrant, 0) // Allow another call if this fails.
		}
	}()

	// TODO(IB): handle fakesentry? Allows for interception / debugging of
	// data sent to sentry.
	//
	//     if opts.Debug && setupFakeSentry != nil { ... }
	//
	err = sentry.Init(opts)
	return err
}

// ParseFrame parses a single sentry.Frame from a string produced by a
// StackTracer.
func ParseFrame(str string) sentry.Frame {
	fnName, file, lineNo := parseFrameStr(str)

	return sentry.Frame{
		Function: fnName,
		Filename: filepath.Base(file), // This will become "." if file is empty.
		AbsPath:  file,
		Lineno:   lineNo,
	}
}

// ParseFrames returns a slice of sentry.Frames for string values
// produced by a StackTracer. It accepts interfaces as it is meant to be
// used with JSON marshaling; otherwise call ParseFrame directly.
func ParseFrames(vals ...interface{}) []sentry.Frame {
	frames := make([]sentry.Frame, 0, len(vals))
	for _, v := range vals {
		s, ok := v.(string)
		if !ok {
			break
		}
		frames = append(frames, ParseFrame(s))
	}
	return frames
}

// NOTE(IB): if this fails because stack was "unknown" then just fnName
// becomes unknown.
//
// Unwraps the frame description as structured in
// errors.Frame.MarshalText:
//   https://github.com/pkg/errors/blob/master/stack.go
func parseFrameStr(frame string) (fnName, file string, lineNo int) {
	//nolint:errcheck
	fmt.Sscanf(frame, "%s %s:%d", &fnName, &file, &lineNo)
	return
}
