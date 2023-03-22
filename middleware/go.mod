module github.com/secureworks/logger/middleware

go 1.19

require (
	github.com/pkg/errors v0.9.1
	github.com/secureworks/errors v0.1.2
	github.com/secureworks/logger/drivers/testlogger v1.1.0
	github.com/secureworks/logger/log v1.1.0
)

replace (
	github.com/secureworks/logger/drivers/testlogger => ../drivers/testlogger
	github.com/secureworks/logger/log => ../log
)
