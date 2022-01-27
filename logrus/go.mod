module github.com/secureworks/logger/logrus

go 1.16

replace (
	github.com/secureworks/logger/internal => ../internal
	github.com/secureworks/logger/log => ../log
	github.com/secureworks/logger/middleware => ../middleware
	github.com/secureworks/logger/testlogger => ../testlogger
)

require (
	github.com/getsentry/sentry-go v0.12.0
	github.com/makasim/sentryhook v0.4.0
	github.com/pkg/errors v0.9.1
	github.com/secureworks/logger/internal v0.0.0-00010101000000-000000000000
	github.com/secureworks/logger/log v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
)
