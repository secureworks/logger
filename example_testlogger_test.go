package logger_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/secureworks/logger/log"
	"github.com/secureworks/logger/testlogger"
)

func Example_usingTestlogger() {
	os.Setenv(string(log.Environment), "test")

	var logger log.Logger
	if strings.ToLower(os.Getenv(string(log.Environment))) == "test" {
		logger, _ = log.Open("test", nil)
	} else {
		logger, _ = log.Open("zerolog", nil)
	}

	// Alternatively you can use testlogger.New. But the above allows us
	// to dynamically set the logger based on the environment.

	entry := logger.Info()
	entry.WithStr("tfield", "test-value")
	entry.WithError(errors.New("error message"))
	entry.Msg("test message")

	if tl, ok := logger.(*testlogger.Logger); ok {
		// You can access the entire output as a bytes.Buffer.
		fmt.Println(tl.Config.Output.(*bytes.Buffer).String())

		// You can also access specific entries and use utility helpers.
		fmt.Println(tl.Entries()[0].Message)
		fmt.Println(tl.Entries()[0].Fields["error"])
		fmt.Println(tl.Entries()[0].StringField("tfield"))
	}

	// Output: {"error":"error message","level":"INFO","message":"test message","tfield":"test-value"}
	// test message
	// error message
	// test-value
}
