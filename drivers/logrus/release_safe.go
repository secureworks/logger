//go:build safe
// +build safe

package logrus

import "github.com/sirupsen/logrus"

// Implements a noop for releaseEntry when not using "unsafe".
func releaseEntry(log *logrus.Logger, ent *logrus.Entry) {}
