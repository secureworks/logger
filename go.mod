module github.com/secureworks/logger

go 1.16

require (
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.26.1
	github.com/secureworks/logger/internal v1.0.0
	github.com/secureworks/logger/log v1.0.0
	github.com/secureworks/logger/logrus v1.0.0
	github.com/secureworks/logger/middleware v1.0.0
	github.com/secureworks/logger/testlogger v1.0.0
	github.com/secureworks/logger/zerolog v1.0.0
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/sys v0.0.0-20220204135822-1c1b9b1eba6a // indirect
)

replace github.com/secureworks/logger/testlogger => ./testlogger

// FIXME
