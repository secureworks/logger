package log_test

import (
	"errors"
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"

	"github.com/secureworks/logger/log"
	_ "github.com/secureworks/logger/logrus"
	_ "github.com/secureworks/logger/zerolog"
)

func ExampleOption() {
	optionFn := func(val interface{}) error {
		ul, _ := val.(log.UnderlyingLogger)
		logger, _ := ul.GetLogger().(*logrus.Logger)
		logger.Formatter = &logrus.JSONFormatter{
			PrettyPrint:      true,
			DisableTimestamp: true,
		}
		return nil
	}
	logger, _ := log.Open("logrus", &log.Config{Output: os.Stdout}, optionFn)

	logger.Info().WithStr("test").Msg("test message here")
	// Output: {
	//   "level": "info",
	//   "msg": "test message here"
	// }
}

type SingleHook struct{}

func (h SingleHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	e.Bool("test_hook", true)
}

func ExampleCustomOption_with_single_value() {
	logger, _ := log.Open(
		"zerolog",
		&log.Config{Output: os.Stdout},

		// Zerolog's Logger.Hook method is chainable (returns a new loger with
		// the given hook attached), so CustomOption will reset the underlying
		// logger to the result.
		//
		// See: https://pkg.go.dev/github.com/rs/zerolog#Logger.Hook
		//
		log.CustomOption("Hook", SingleHook{}),
	)

	logger.Info().Msg("test message here")
	// Output: {"level":"info","test_hook":true,"message":"test message here"}
}

func ExampleCustomOption_with_possible_error() {
	// If a CustomOption returns an error value that is not nil, that
	// error bubbles up through log.Open.
	loggerFailed, err := log.Open(
		"zerolog",
		&log.Config{Output: os.Stdout},
		log.CustomOption("Sample", func() (zerolog.Sampler, error) {
			return &zerolog.BasicSampler{N: 2}, errors.New("custom option failed")
		}),
	)

	if loggerFailed == nil {
		fmt.Printf("ERR: %q\n", err)
	}

	// If a CustomOption returns an error value that is nil, that value is
	// ignored and the named method receives the given value(s).
	loggerSuccess, _ := log.Open(
		"zerolog",
		&log.Config{Output: os.Stdout},
		log.CustomOption("Sample", func() (zerolog.Sampler, error) {
			return &zerolog.BasicSampler{N: 2}, nil
		}),
	)

	loggerSuccess.Info().Msg("test success message 1")
	loggerSuccess.Info().Msg("test success message 2")
	loggerSuccess.Info().Msg("test success message 3")
	loggerSuccess.Info().Msg("test success message 4")

	// Output: ERR: "custom option failed"
	// {"level":"info","message":"test success message 1"}
	// {"level":"info","message":"test success message 3"}
}
