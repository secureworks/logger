# Secureworks Unified Logging Library

[![logger release (latest SemVer)](https://img.shields.io/github/v/release/secureworks/taegis-sdk-go?sort=semver)](https://github.com/secureworks/taegis-sdk-go/releases)
[![Test Status](https://github.com/secureworks/taegis-sdk-go/workflows/gitleaks/badge.svg)](https://github.com/secureworks/taegis-sdk-go/actions?query=workflow%3Agitleaks)
[![Coverage Status](https://coveralls.io/repos/github/secureworks/taegis-sdk-go/badge.svg?branch=master)](https://coveralls.io/github/secureworks/taegis-sdk-go?branch=master)

`secureworks/logger` is a unified interface that wraps popular logging libraries such as [Logrus][logrus] and [Zerolog][zerolog], and can instrument automatic reporting to services such as [Sentry][sentry]... _And that is just the beginning!_

This is the official Golang logging library used in such projects as the [SecureWorks Taegis™ XDR (Extended Detection and Response)][taegis-xdr] Cloud Platform, and is suggested for use with the [Taegis Golang SDK][taegis-sdk].

## Installation

Run the command:

```
$ go get -u github.com/secureworks/logger/log
```

Alternatively, if your project is using Go Modules then, reference `logger` in a file with import:

```go
import "github.com/secureworks/logger/log"
```

Run any Go command and the toolchain will resolve and fetch the `logger` module automatically.

## Usage

[Documentation is available on pkg.go.dev][godocs]. You should also look at the examples in the main package.

## FAQ
- Why are there so many go.mods?
    - In order to keep dependencies in line with the log implementations. If you want zerolog you shouldn't also need logrus.
- There are some packages with safe and unsafe versions of code. Why is this?
    - For logrus, the unsafe code is mostly a performance trick, to keep it from generating even more garbage than normal.
    - For zerolog it is the same as logrus, but it also addresses a small behavior change in the zerolog.Hook interface. See this [issue](https://github.com/rs/zerolog/issues/408) for more.
    - All unsafe code can be disabled by adding a `safe` or `!unsafe` build tag. This may be useful if you do not desire the changes the unsafe code brings or because you are building for an environment that does not allow unsafe code.

## License

This library is distributed under the [Apache-2.0 license][apache-2] found in the [LICENSE](./LICENSE) file.

### Runtime Dependencies
Note these are dependent on what mods you import. Importing `log` by itself will only yield `testify` for testing.
Importing `logrus` will only yield its dependencies and not `zerolog`'s for example. 

| Library                                                                    | Purpose                         | License                                                          |
| -------------------------------------------------------------------------- | ------------------------------- | ---------------------------------------------------------------- |
| [`github.com/pkg/errors`](https://github.com/pkg/errors)                   | Extracts error stack traces.    | [BSD 2-Clause](https://choosealicense.com/licenses/bsd-2-clause) |
| [`github.com/rs/zerolog`](https://github.com/rs/zerolog)                   | Logger.                         | [MIT](https://choosealicense.com/licenses/mit/)                  |
| [`github.com/sirupsen/logrus`](https://github.com/sirupsen/logrus)         | Logger.                         | [MIT](https://choosealicense.com/licenses/mit/)                  |
| [`github.com/getsentry/sentry-go`](https://github.com/getsentry/sentry-go) | Sentry SDK for error reporting. | [BSD 2-Clause](https://choosealicense.com/licenses/bsd-2-clause) |
| [`github.com/makasim/sentryhook`](https://github.com/makasim/sentryhook)   | Sentry hook for Logrus          | [MIT](https://choosealicense.com/licenses/mit/)                  |

### Test Dependencies

| Library                                                                          | Purpose                   | License                                         |
| -------------------------------------------------------------------------------- | ------------------------- | ----------------------------------------------- |
| [`github.com/stretchr/testify`](https://github.com/stretchr/testify)             | Test tooling.             | [MIT](https://choosealicense.com/licenses/mit/) |
| [`github.com/VerticalOps/fakesentry`](https://github.com/VerticalOps/fakesentry) | Run a fake Sentry server. | [MIT](https://choosealicense.com/licenses/mit/) |

<!-- Links -->

[taegis-xdr]: https://www.secureworks.com/products/taegis/xdr
[taegis-sdk]: https://github.com/secureworks/taegis-sdk-go
[godocs]: https://pkg.go.dev/github.com/secureworks/logger
[logrus]: https://github.com/sirupsen/logrus
[zerolog]: https://github.com/rs/zerolog
[sentry]: https://docs.sentry.io/platforms/go/
[apache-2]: https://choosealicense.com/licenses/apache-2.0/
