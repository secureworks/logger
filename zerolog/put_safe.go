//go:build !unsafe
// +build !unsafe

package zerolog

import (
	"github.com/rs/zerolog"
)

// Implements a noop for putEvent when not using "unsafe".
func putEvent(ent *zerolog.Event) {}
