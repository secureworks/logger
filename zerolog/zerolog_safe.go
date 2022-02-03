//go:build safe
// +build safe

package zerolog

import (
	"github.com/rs/zerolog"
)

// Implements a noop for putEvent when not using "unsafe".
func putEvent(ent *zerolog.Event) {}

// Implements a noop for changeEventLevel when not using "unsafe".
func changeEventLevel(ent *zerolog.Event, lvl zerolog.Level) {}
