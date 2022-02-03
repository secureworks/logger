//go:build !safe
// +build !safe

package zerolog

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/secureworks/logger/log"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestZerolog_UnsafeEventLevelChange(t *testing.T) {
	require := require.New(t)

	// 'blank'/new event defaults to Debug in zerolog
	// there is no way to create a new Event with something else
	// outside of chaining one from a logger with a different level
	event := zerolog.Dict()

	// Since there is also no method/field for getting the level itself
	// we will use reflect. You can use reflect to (indirectly) read the value of a
	// private field, but you cannot use reflect to set it.
	// While some operations below are unsafe in their own right
	// the point of this test is to alert us when something in zerolog changes
	// so any panic/failure is fine.
	val := reflect.ValueOf(event).Elem().FieldByName("level")
	require.True(val.IsValid(), "cannot find 'level' struct field in zerolog.Event")

	// Make sure we have expected default value, and fields line up
	// Cannot use String method, but we can use Int
	require.Equal(int64(zerolog.DebugLevel), val.Int())

	// Change field
	changeEventLevel(event, zerolog.WarnLevel)

	// Make sure change is correct
	require.Equal(int64(zerolog.WarnLevel), val.Int())
}

// checkLevel implements the zerolog.Hook interface to make sure
// the expected level is passed when Run.
type checkLevel struct {
	req     *require.Assertions
	toCheck zerolog.Level
}

func (cl *checkLevel) Run(e *zerolog.Event, lvl zerolog.Level, msg string) {
	cl.req.Equal(cl.toCheck, lvl, "incorrect log level")
}

func TestZerolog_ExpectedHookBehavior(t *testing.T) {
	require := require.New(t)

	zlog, err := log.Open("zerolog", nil, log.CustomOption(
		"Hook",
		&checkLevel{req: require, toCheck: zerolog.ErrorLevel},
	), log.CustomOption("Output", ioutil.Discard))
	require.Nil(err)

	// Hook will run and check values
	zlog.Error().Msg("foobar")
}
