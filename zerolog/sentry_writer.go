package zerolog

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"

	"github.com/secureworks/logger/internal/common"
	"github.com/secureworks/logger/log"
)

// An unfortunate but necessary type, as Zerolog's hook interface is
// not conducive to data extraction. Older Sentry writers used
// raven-go instead of sentry-go and zerolog authors themselves seem to
// prefer using io.Writer instead of the hook interface. See also:
//   - https://github.com/rs/zerolog/blob/72acd6cfe8bbbf5c52bfc805a3889c6941499c95/journald/journald.go#L37
//   - https://github.com/rs/zerolog/blob/72acd6cfe8bbbf5c52bfc805a3889c6941499c95/console.go#L86
//   - https://github.com/rs/zerolog/issues/93
//   - https://gist.github.com/asdine/f821abe6189a04250ae61b77a3048bd9
//
type sentryWriter struct {
	lvlField []byte
	hub      *sentry.Hub
	lvlSet   map[zerolog.Level]bool
}

// Create a new Sentry writer attached to CurrentHub for the given set
// of levels. This writer is a peer to Zerolog in the logger
// implementation, skipping the Zerolog hook system entirely.
func newSentryWriter(hub *sentry.Hub, lvls ...log.Level) *sentryWriter {
	if hub == nil {
		hub = sentry.CurrentHub()
	}
	lvlSet := make(map[zerolog.Level]bool, len(lvls))
	for _, lvl := range lvls {
		lvlSet[lvlToZerolog(lvl)] = true
	}

	return &sentryWriter{
		lvlField: []byte(fmt.Sprintf(`"%s"`, zerolog.LevelFieldName)),
		hub:      hub,
		lvlSet:   lvlSet,
	}
}

// Write parses the message sent to Zerolog, and if it warrants sending
// an error to Sentry then it extracts the necessary information. It is
// a full write-to-Sentry implementation.
func (sw *sentryWriter) Write(msg []byte) (n int, err error) {
	n = len(msg)

	// Get the log level, and if it meets the threshold for Sentry.
	zlvl, ok := sw.checkLevel(msg)
	if !ok {
		return
	}

	// Extract JSON log entry into a basic container.
	data := make(map[string]interface{})
	err = json.Unmarshal(msg, &data)
	if err != nil {
		return
	}
	delete(data, zerolog.LevelFieldName) // Remove the level field.
	if len(data) == 0 {
		return
	}

	event := sentry.NewEvent()
	event.Level = zerologLevelToSentry(zlvl)

	// Get the log message if available.
	if msg, ok := data[zerolog.MessageFieldName].(string); ok {
		event.Message = msg
		delete(data, zerolog.MessageFieldName)
	}

	// Get the log error and stack trace (if available) and push on to
	// Sentry event Exception field if found.
	exc := sentryExceptionFromFields(
		data, zerolog.ErrorFieldName, zerolog.ErrorStackFieldName)
	if exc != nil {
		event.Exception = append(event.Exception, *exc)
	}

	// Get the log panic and stack trace (if available). Push on to the
	// Sentry event Exception field if found.
	exc = sentryExceptionFromFields(
		data, log.PanicValue, log.PanicStack)
	if exc != nil {
		event.Exception = append(event.Exception, *exc)
	}

	// Additional values as "Extra" vs "Tags" vs "Breadcrumbs."
	event.Extra = data
	sw.hub.CaptureEvent(event)

	return
}

// checkLevel looks up the level of the message given in bytes, and
// returns if it warrants being sent to Sentry.
func (sw *sentryWriter) checkLevel(msg []byte) (zerolog.Level, bool) {
	if msg == nil || len(msg) < len(sw.lvlField) {
		return 0, false
	}

	// The level field will usually be in the latter part of the message.
	i := bytes.Index(msg[len(msg)/2:], sw.lvlField)
	if i == -1 {
		// Try the full slice, wasn't in the latter half (or was split
		// between halfs).
		i = bytes.Index(msg, sw.lvlField)
	} else {
		i += (len(msg) / 2) // Found in back: add the len we skipped.
	}
	if i == -1 {
		return 0, false // Level not found, do not send error to Sentry.
	}

	// Parse the level field.
	startingOffset := i + len(sw.lvlField) + 2 // Ie: `...level":`
	if startingOffset >= len(msg) {
		return 0, false
	}
	i = bytes.IndexByte(msg[startingOffset:], '"')
	if i == -1 {
		return 0, false
	}
	i += startingOffset
	lvl, err := zerolog.ParseLevel(string(msg[startingOffset:i]))
	return lvl, err == nil && sw.lvlSet[lvl]
}

// sentryExceptionFromFields attempts to structure a Sentry Exception
// from an error message and/or stack trace.
func sentryExceptionFromFields(data map[string]interface{}, msgf string, stf string) *sentry.Exception {
	var exc *sentry.Exception

	// Look up error message field.
	if iface, ok := data[msgf]; ok {
		pv, ok := iface.(string)
		if !ok {
			pv = fmt.Sprintf("%v", pv)
		}
		delete(data, msgf)
		exc = &sentry.Exception{Value: pv}
	}

	// Either add a stack trace or create an exception with a stack trace
	// if necessary.
	if iface, ok := data[stf].([]interface{}); ok {
		frames := common.ParseFrames(iface...)
		if len(frames) > 0 {
			delete(data, stf)
			if exc == nil {
				exc = new(sentry.Exception)
			}
			exc.Stacktrace = &sentry.Stacktrace{Frames: frames}
		}
	}

	return exc
}

// Map zerolog log levels to Sentry log levels.
func zerologLevelToSentry(lvl zerolog.Level) sentry.Level {
	switch lvl {
	case zerolog.TraceLevel, zerolog.DebugLevel:
		return sentry.LevelDebug
	case zerolog.InfoLevel:
		return sentry.LevelInfo
	case zerolog.WarnLevel:
		return sentry.LevelWarning
	case zerolog.ErrorLevel, zerolog.PanicLevel:
		return sentry.LevelError
	case zerolog.FatalLevel:
		return sentry.LevelFatal
	default:
		return sentry.LevelWarning
	}
}
