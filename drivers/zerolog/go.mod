module github.com/secureworks/logger/drivers/zerolog

go 1.19

require (
	github.com/rs/zerolog v1.27.0
	github.com/secureworks/errors v0.1.2
	github.com/secureworks/logger/log v1.1.0
)

require (
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	golang.org/x/sys v0.0.0-20220804214406-8e32c043e418 // indirect
)

replace github.com/secureworks/logger/log => ../../log
