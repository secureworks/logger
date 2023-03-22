module github.com/secureworks/logger

go 1.19

require (
	github.com/rs/zerolog v1.27.0
	github.com/secureworks/logger/drivers/testlogger v1.1.0
	github.com/secureworks/logger/log v1.1.0
	github.com/secureworks/logger/logrus v1.1.0
	github.com/secureworks/logger/middleware v1.1.1
	github.com/secureworks/logger/zerolog v1.1.0
	github.com/sirupsen/logrus v1.9.0
)

require (
	github.com/getsentry/sentry-go v0.13.0 // indirect
	github.com/makasim/sentryhook v0.4.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/secureworks/errors v0.1.2 // indirect
	github.com/secureworks/logger/internal v1.1.0 // indirect
	golang.org/x/sys v0.0.0-20220804214406-8e32c043e418 // indirect
)

replace (
	github.com/secureworks/logger/drivers/testlogger => ./drivers/testlogger
	github.com/secureworks/logger/log => ./log
	github.com/secureworks/logger/middleware => ./middleware
)
