//go:build !safe
// +build !safe

package logrus

import (
	// NOTE(IB): linter wants me to add a comment so others can commit the
	// same sin. "unsafe" needs to be at least underscore imported to use
	// go:linkname.
	_ "unsafe"

	"github.com/sirupsen/logrus"
)

// For some reason Logrus doesn't expose the releaseEntry method, only
// the NewEntry one.
//
//go:linkname releaseEntry github.com/sirupsen/logrus.(*Logger).releaseEntry
func releaseEntry(log *logrus.Logger, ent *logrus.Entry)
