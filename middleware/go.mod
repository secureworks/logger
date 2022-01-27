module github.com/secureworks/logger/middleware

go 1.16

replace (
	github.com/secureworks/logger/internal => ../internal
	github.com/secureworks/logger/log => ../log
	github.com/secureworks/logger/middleware => ./
	github.com/secureworks/logger/testlogger => ../testlogger
)

require (
	github.com/pkg/errors v0.9.1
	github.com/secureworks/logger/internal v0.0.0-00010101000000-000000000000
	github.com/secureworks/logger/log v0.0.0-00010101000000-000000000000
	github.com/secureworks/logger/testlogger v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.7.0
)
