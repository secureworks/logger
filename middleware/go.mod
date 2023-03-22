module github.com/secureworks/logger/middleware

go 1.19

require (
	github.com/pkg/errors v0.9.1
	github.com/secureworks/errors v0.1.2
	github.com/secureworks/logger/drivers/testlogger v1.1.0
	github.com/secureworks/logger/internal v1.1.0
	github.com/secureworks/logger/log v1.1.0
)

require (
	github.com/VerticalOps/fakesentry v0.0.0-20200925184942-401321fe17b3 // indirect
	github.com/getsentry/sentry-go v0.13.0 // indirect
	golang.org/x/sys v0.0.0-20220804214406-8e32c043e418 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)

replace (
	github.com/secureworks/logger/drivers/testlogger => ../drivers/testlogger
	github.com/secureworks/logger/log => ../log
)
