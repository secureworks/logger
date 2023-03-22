//go:build !safe
// +build !safe

package zerolog

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/rs/zerolog"

	"github.com/secureworks/logger/log"
	"github.com/secureworks/logger/log/testutils"
)

func TestZerolog_UnsafeEventLevelChange(t *testing.T) {
	// A "blank," new Zerolog Event defaults to Debug. There is no way to
	// create a new Event with something else other than chaining one from a
	// logger with a different level.
	event := zerolog.Dict()

	// Since there is also no method / field for getting the level itself we
	// will use reflect. You can use reflect to (indirectly) read the value
	// of a private field, but you cannot use reflect to set it. While some
	// operations below are unsafe in their own right the point of this test
	// is to alert us when something in zerolog changes so any panic/failure
	// is fine.
	val := reflect.ValueOf(event).Elem().FieldByName("level")
	testutils.AssertTrue(t, val.IsValid())

	// Make sure we have the expected default value, and that fields line
	// up. Cannot use String method, but we can use Int.
	testutils.AssertEqual(t, int64(zerolog.DebugLevel), val.Int())

	// Change field and ensure correct.
	changeEventLevel(event, zerolog.WarnLevel)
	testutils.AssertEqual(t, int64(zerolog.WarnLevel), val.Int())
}

func TestZerolog_ExpectedHookBehavior(t *testing.T) {
	hook := &checkLevelHook{}
	logger, err := log.Open(
		"zerolog",
		nil,
		log.CustomOption("Hook", hook),
		log.CustomOption("Output", ioutil.Discard),
	)
	testutils.AssertNil(t, err)

	// Hook will run and store the error value passed.
	logger.Error().Msg("foobar")
	testutils.AssertEqual(t, zerolog.ErrorLevel, hook.EventLevel)
}

// checkLevelHook implements the zerolog.Hook interface to make sure
// the expected level is passed when Run.
type checkLevelHook struct {
	EventLevel zerolog.Level
}

func (cl *checkLevelHook) Run(_ *zerolog.Event, lvl zerolog.Level, _ string) {
	cl.EventLevel = lvl
}
