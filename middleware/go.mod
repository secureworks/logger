module github.com/secureworks/logger/middleware

go 1.16

replace github.com/secureworks/logger/middleware v0.0.0-20220203224804-3685f341a5e1 => ./

require (
	github.com/pkg/errors v0.9.1
	github.com/secureworks/logger/internal v0.0.0-20220203224804-3685f341a5e1
	github.com/secureworks/logger/log v0.0.0-20220203224804-3685f341a5e1
	github.com/secureworks/logger/testlogger v0.0.0-20220203224804-3685f341a5e1
	github.com/stretchr/testify v1.7.0
)
