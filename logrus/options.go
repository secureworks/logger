package logrus

import (
	"fmt"

	"github.com/secureworks/logger/log"
)

// ReusableEntries returns a log.Option to be used with the logrus driver.
// It instructs this driver to change the behavior of Send on the Entry interface
// such that it can be reusable (called multiple times). Read the log package
// interface documentation for more.
func ReusableEntries() log.Option {
	return func(i interface{}) error {
		l, ok := i.(*logger)
		if !ok {
			return fmt.Errorf("log/logrus: Unexpected type passed to log option: %T", i)
		}

		l.reusableEntries = true
		return nil
	}
}
