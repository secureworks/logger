// Package logger is the home of the Secureworks logger. The Secureworks
// logger is a unified interface for various logger implementations to
// allow an organization with multiple different logging strategies or
// needs that are implemented in different libraries to share best
// practices, tools and utilities for logging. The unified interface
// also focuses on ease of use for type-safe logging and flexible
// approaches.
//
// The Secureworks logger is also integrated with reporting hooks,
// currently Sentry, so that users can focus on generating logs and get
// such error reporting for free.
//
// Finally, the Secureworks logger makes testing assertions around
// logging easy (using the "test" testlogger driver).
//
// To use the logger you need to import the driver package you want to
// use and then generate a logger with log.Open and whatever
// configuration you want. The simplest setup would look like:
//
//	package main
//
//	import (
//	    "github.com/secureworks/logger/log"
//	    _ "github.com/secureworks/logger/zerolog"
//	)
//
//	func main() {
//	    config := log.DefaultConfig()
//	    config.EnableErrStack = true
//
//	    logger, err := log.Open("zerolog", config)
//	    if err != nil {
//	        panic(err)
//	    }
//
//	    logger.Debug().Msg("logger instantiated")
//	}
//
// See the examples for common use cases.
package logger
