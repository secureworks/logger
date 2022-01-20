package logger_test

import (
	"errors"
	"fmt"
	"os"

	"github.com/secureworks/logger/log"
	_ "github.com/secureworks/logger/zerolog"
)

func Example() {
	// You can set configuration from the env ...
	os.Setenv(string(log.LogLevel), "DEBUG")
	os.Setenv(string(log.LocalDevel), "true")

	config := log.DefaultConfig(os.Getenv) // Same as passing nil here.

	// ... but setting the output must be done directly:
	config.Output = os.Stdout

	fmt.Println()

	// If the second (config) argument passed to log.Open is nil then
	// log.DefaultConfig(nil) is assumed.
	logger, _ := log.Open("zerolog", config)

	// Entry and the associated methods (Error, Info, etc) create log
	// "entries:" specific log lines that will be written to the log. The
	// entries are, by default, logged as JSON with a series of fields.
	entry := logger.Entry(log.INFO)

	// With… methods add fields to an entry. They may include some
	// specific handling based on the type being written, if the logger
	// implementation demands it, but also are useful for ensuring how the
	// value is represented in JSON.
	entry.WithField("meta", "data")
	entry.WithBool("bool", false)

	// You may use Msg or Msgf to add a log message field and write.
	entry.Msg("standard message")

	// You may use WithError, WithField or WithFields directly on a Logger
	// to create an entry at the default log level and give it those
	// fields.
	errEntry := logger.WithError(errors.New("error message"))

	// You may write an entry without adding a "message" field by using
	// Send.
	errEntry.Send()

	// You can also use Async to make an entry not write when the message
	// is written. In this case the message can be overwritten and more
	// fields added. This is especially useful for canonical log lines.
	asyncEntry := logger.Info().Async()
	asyncEntry.Msg("async message")
	asyncEntry.WithStr("meta", "data")
	asyncEntry.Msg("async message: now with meta data")
	asyncEntry.Send()

	// Output:
	// {"meta":"data","bool":false,"level":"info","message":"standard message"}
	// {"error":"error message","level":"error"}
	// {"meta":"data","level":"info","message":"async message: now with meta data"}
}
