module github.com/secureworks/logger/middleware

go 1.16

replace github.com/secureworks/logger/middleware v0.0.0-20220120225355-031cd30cdd6a => ./

require (
	github.com/VerticalOps/fakesentry v0.0.0-20200925184942-401321fe17b3 // indirect
	github.com/getsentry/sentry-go v0.12.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/secureworks/logger/internal v0.0.0-20220127223102-22a7dfffc647
	github.com/secureworks/logger/log v0.0.0-20220127222910-3bc416d4c2aa
	github.com/secureworks/logger/testlogger v0.0.0-20220127223102-22a7dfffc647
	github.com/stretchr/testify v1.7.0
)
