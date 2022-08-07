module github.com/secureworks/logger/middleware

go 1.16

require (
	github.com/pkg/errors v0.9.1
	github.com/secureworks/logger/internal v1.0.0
	github.com/secureworks/logger/log v1.0.0
	github.com/secureworks/logger/testlogger v1.0.0
)

replace (
	github.com/secureworks/logger/log => ../log
	github.com/secureworks/logger/internal => ../internal
)
