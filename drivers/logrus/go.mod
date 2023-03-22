module github.com/secureworks/logger/drivers/logrus

go 1.19

require (
	github.com/secureworks/errors v0.1.2
	github.com/secureworks/logger/log v1.1.0
	github.com/sirupsen/logrus v1.9.0
)

require golang.org/x/sys v0.0.0-20220804214406-8e32c043e418 // indirect

replace github.com/secureworks/logger/log => ../../log
