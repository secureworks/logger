module github.com/secureworks/logger/internal

go 1.16

replace (
	github.com/secureworks/logger/internal => ./
	github.com/secureworks/logger/log => ../log
	github.com/secureworks/logger/middleware => ../middleware
	github.com/secureworks/logger/testlogger => ../testlogger
)

require (
	github.com/VerticalOps/fakesentry v0.0.0-20200925184942-401321fe17b3
	github.com/getsentry/sentry-go v0.12.0
	github.com/pkg/errors v0.9.1
	github.com/secureworks/logger/log v0.0.0-00010101000000-000000000000
	github.com/secureworks/logger/middleware v0.0.0-00010101000000-000000000000
	github.com/secureworks/logger/testlogger v0.0.0-00010101000000-000000000000
)
