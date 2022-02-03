//go:build !safe
// +build !safe

package zerolog

import (
	// NOTE(IB): linter wants me to add a comment so others can commit the
	// same sin. unsafe needs to be at least underscore imported to use
	// go:linkname.
	_ "unsafe"

	"github.com/rs/zerolog"
)

// Since we want the zerolog logger implementation to support both
// changing levels or "Disabled", this is needed in order to retain
// zerologs performance for high event counts: if we don't write it
// won't be reused. See also:
//   - https://github.com/rs/zerolog/pull/255
//   - https://github.com/rs/zerolog/blob/7825d863376faee2723fc99c061c538bd80812c8/event.go#L79
//
//go:linkname putEvent github.com/rs/zerolog.putEvent
func putEvent(ent *zerolog.Event)
