package logrus

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/makasim/sentryhook"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/secureworks/logger/internal/common"
	"github.com/secureworks/logger/log"
)

// The sentryhook library provides a basic Logrus hook for sending
// error logs to Sentry. Here we provide the hook with our specific
// log-entry-to-Sentry conversion.
func sentryConverter(entry *logrus.Entry, event *sentry.Event, hub *sentry.Hub) {
	// Handles structuring the contents of the logrus.ErrorKey event data
	// field on to event.Exception.
	sentryhook.DefaultConverter(entry, event, hub)

	// Add panic errors in the log.PanicValue event data field.
	if iface, ok := entry.Data[log.PanicValue]; ok {
		pv, ok := iface.(string)
		if !ok {
			pv = fmt.Sprintf("%v", pv)
		}
		event.Exception = append(event.Exception, sentry.Exception{Value: pv})
	}

	// If we have a panic value then either append its stack trace to the
	// Exception or make a new Exception with its stack trace.
	if st, ok := entry.Data[log.PanicStack].(errors.StackTrace); ok && len(st) > 0 {
		frames := make([]sentry.Frame, 0, len(st))
		for i, f := range st {
			dat, _ := f.MarshalText()
			frames[i] = common.ParseFrame(string(dat))
		}

		trace := &sentry.Stacktrace{Frames: frames}
		if len(event.Exception) > 0 {
			event.Exception[len(event.Exception)-1].Stacktrace = trace
		} else {
			event.Exception = []sentry.Exception{{Stacktrace: trace}}
		}
	}
}
