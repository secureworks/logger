module github.com/secureworks/logger

go 1.16

replace (
	github.com/secureworks/logger/internal => ./internal
	github.com/secureworks/logger/log => ./log
	github.com/secureworks/logger/logrus => ./logrus
	github.com/secureworks/logger/middleware => ./middleware
	github.com/secureworks/logger/testlogger => ./testlogger
	github.com/secureworks/logger/zerolog => ./zerolog
)

require (
	github.com/rs/zerolog v1.26.1
	github.com/secureworks/logger/log v0.0.0-00010101000000-000000000000
	github.com/secureworks/logger/logrus v0.0.0-00010101000000-000000000000
	github.com/secureworks/logger/middleware v0.0.0-00010101000000-000000000000
	github.com/secureworks/logger/testlogger v0.0.0-00010101000000-000000000000
	github.com/secureworks/logger/zerolog v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
)
