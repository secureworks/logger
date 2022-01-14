package logger_test

import (
	"os"

	"github.com/secureworks/logger/log"
	_ "github.com/secureworks/logger/zerolog"
)

// Options (specifically CustomOption) can be used to specificy behavior
// for the logger implementation beyond the values in the Config. For
// more examples see log.Option and log.CustomOption.
func Example_usingOptions() {
	config := log.DefaultConfig(nil)
	config.Output = os.Stdout

	// This CustomOption is attaching a hook using Zerolog's Logger.Hook method.
	// See: https://pkg.go.dev/github.com/rs/zerolog#Logger.Hook
	//
	logger, _ := log.Open("zerolog", config, log.CustomOption("Hook", SingleHook{}))

	logger.Info().Msg("test message here")
	// Output: {"level":"info","test_hook":true,"message":"test message here"}
}
