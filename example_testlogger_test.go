package logger_test

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/secureworks/logger/log"
	"github.com/secureworks/logger/testlogger"
)

func Example_usingTestlogger() {
	var logger log.Logger
	logger, _ = testlogger.New(nil)

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
		entry := tl.GetEntries()[0]
		fmt.Println(entry.Message)
		fmt.Println(entry.Fields["error"])
		fmt.Println(entry.StringField("tfield"))
	}

	// Output: {"error":"error message","level":"INFO","message":"test message","tfield":"test-value"}
	// test message
	// error message
	// test-value
}
