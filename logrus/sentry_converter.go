package logrus

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/makasim/sentryhook"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/secureworks/logger/log"
	"github.com/secureworks/logger/log/internal/common"
)

func sentryConverter(entry *logrus.Entry, event *sentry.Event, hub *sentry.Hub) {
	// Default takes care of most of this for us.
	sentryhook.DefaultConverter(entry, event, hub)

	// Add our own additions.
	if iface, ok := entry.Data[log.PanicValue]; ok {
		pv, ok := iface.(string)
		if !ok {
			pv = fmt.Sprintf("%v", pv)
		}

		event.Exception = append(event.Exception, sentry.Exception{
			Value: pv,
		})
	}

	if st, ok := entry.Data[log.PanicStack].(errors.StackTrace); ok && len(st) > 0 {
		frames := make([]sentry.Frame, 0, len(st))

		for _, f := range st {
			// All of the methods we need are private, just go through
			// MarshalText which cannot fail.
			dat, _ := f.MarshalText()

			frames = append(frames, common.ParseFrame(string(dat)))
		}

		trace := &sentry.Stacktrace{
			Frames: frames,
		}

		if len(event.Exception) > 0 {
			// If we had a panic value above append to its Exception.
			event.Exception[len(event.Exception)-1].Stacktrace = trace
		} else {
			// Otherwise make a new one.
			event.Exception = append(event.Exception, sentry.Exception{
				Stacktrace: trace,
			})
		}
	}
}
