# Secureworks Unified Logging Library

`secureworks/logger` is a unified interface that wraps popular logging 
libraries such as [Logrus][logrus] and [Zerolog][zerolog], and can instrument 
automatic reporting to [Sentry][sentry].

This is the logging library used in 
[SecureWorks Taegis™ XDR (Extended Detection and Response)][taegis-xdr] Cloud 
Platform, and is suggested for use with the [Taegis Golang SDK][taegis-sdk].

### Design Goals FIXME

The design of `secureworks/logger` is meant to solve a group of problems 
related to logging:

1. It provides a single interface that brings together both logging and error 
   reporting.
2. It allows teams to use the logger they believe works best for them, without 
   introducing issues for those whose expectations are built around a 
   different logger.
3. It provides a simple mechanism for testing an application's logging.
4. It offers a uniform interface that guides developers towards logging best 
   practices. This is done by simplifying the API for the `secureworks/logger`
   and implementing a more rigid logging model than the popular loggers.

### The `secureworks/logger` logging model FIXME

The logging model creates a clearer, smaller API for the two main objects: 
`Logger` and `Entry`:
   - fields (metadata set on a single log `Entry`) are never associated with 
     `Loggers`; // can only be associated with a logger when it is created...
   - `Entries` can not spawn or be cloned as new `Entries`, and they can not 
     be written out more than once;
   - `Entries` clearly distinguish between the act of setting fields as 
     opposed to writing out. 

The benefits of this model include:

- Fewer race conditions: It can eliminate, or drastically reduce, race 
conditions that arise from using the most popular logging libraries.
- Reduced log bloat: 
  - do not fill up a logging object with cruft that gets repeated with each cycle
  - use canonical log lines (a single entry per cycle) for web 
services and similar applications with the built in middleware that embeds an 
entry in a context instead of logger
- Controlled log writes: 
  - use async logs
  - can not send a given entry more than once

Understanding this model and using `secureworks/logger` instead of the more 
relaxed popular loggers makes a big difference for your teams logging story. 

## Installation

This library is broken into submodules that are linked together. You may 
download them separately, but the easiest thing to do is import whichever 
driver you want to use (`logrus`, `zerolog`, or `testlogger`), and these will 
include the dependencies you need:

```
$ go get -u github.com/secureworks/logger/logrus
```

If you want the middleware you would also need:

```
$ go get -u github.com/secureworks/logger/middleware
```

Alternatively, if your project is using Go modules then, reference the driver 
package(s) in a file's `import`:

```go
import (
	// ...
	"github.com/secureworks/logger/middleware"
	_ "github.com/secureworks/logger/zerolog"
)
```

You may run any Go command and the toolchain will resolve and fetch the 
required modules automatically.

## Usage

[Documentation is available on pkg.go.dev][godocs]. You may also look at the 
examples in the `logger` package.

## FAQ

- Why are there so many submodules / why do all the packages have `go.mod`s?
    - We have broken the packages up in order to keep dependencies in line 
      with the log implementations. If you want `zerolog` you shouldn't also 
      need `logrus`; if you want to write code that consumes the shared 
      interface you shouldn't need to depend on either implementation.  
- There are some packages with "safe" and "unsafe" versions of code. Why is this?
    - *unsafe* refers to using [the Go standard library `unsafe`][unsafe], 
      which allows us to step outside of Go's type-safety rules. This code is 
      no more "not safe" than a typical C program.
    - While we use the unsafe code (less type-safe) by default, this can be 
      disabled by adding a `safe` or `!unsafe` build tag. This may be useful 
      if you are building for an environment that does not allow unsafe (less 
      type-safe) code.
    - For `zerolog` and `logrus` the unsafe code is used for a big performance 
      boost.
    - For `zerolog` it also addresses a small behavior change in the 
      `zerolog.Hook` interface. 
      **[See this issue for more.](https://github.com/rs/zerolog/issues/408)** 

## License

This library is distributed under the [Apache-2.0 license][apache-2] found in 
the [LICENSE](./LICENSE) file.

### Runtime Dependencies

| Library                                                                    | Purpose                         | License                                                          |
|----------------------------------------------------------------------------|---------------------------------|------------------------------------------------------------------|
| [`github.com/pkg/errors`](https://github.com/pkg/errors)                   | Extracts error stack traces.    | [BSD 2-Clause](https://choosealicense.com/licenses/bsd-2-clause) |
| [`github.com/rs/zerolog`](https://github.com/rs/zerolog)                   | Logger.                         | [MIT](https://choosealicense.com/licenses/mit/)                  |
| [`github.com/sirupsen/logrus`](https://github.com/sirupsen/logrus)         | Logger.                         | [MIT](https://choosealicense.com/licenses/mit/)                  |
| [`github.com/getsentry/sentry-go`](https://github.com/getsentry/sentry-go) | Sentry SDK for error reporting. | [BSD 2-Clause](https://choosealicense.com/licenses/bsd-2-clause) |
| [`github.com/makasim/sentryhook`](https://github.com/makasim/sentryhook)   | Sentry hook for Logrus.         | [MIT](https://choosealicense.com/licenses/mit/)                  |

> _**Note:** these are different based on what you import._ Given which submodule(s) you use, the dependencies
> are included as follows:
>
> - `github.com/secureworks/logger/log`:
>   - *No external dependencies.*
> - `github.com/secureworks/logger/testlogger`: 
>   - *No external dependencies.*
> - `github.com/secureworks/logger/middleware`:
>   - [`github.com/pkg/errors`](https://github.com/pkg/errors)
> - `github.com/secureworks/logger/logrus`:
>   - [`github.com/pkg/errors`](https://github.com/pkg/errors)
>   - [`github.com/sirupsen/logrus`](https://github.com/sirupsen/logrus)
>   - [`github.com/getsentry/sentry-go`](https://github.com/getsentry/sentry-go)
>   - [`github.com/makasim/sentryhook`](https://github.com/makasim/sentryhook)
> - `github.com/secureworks/logger/zerolog`:
>   - [`github.com/rs/zerolog`](https://github.com/rs/zerolog)
>   - [`github.com/getsentry/sentry-go`](https://github.com/getsentry/sentry-go)

### Test Dependencies

| Library                                                                          | Purpose                         | License                                                          |
|----------------------------------------------------------------------------------|---------------------------------|------------------------------------------------------------------|
| [`github.com/pkg/errors`](https://github.com/pkg/errors)                         | Extracts error stack traces.    | [BSD 2-Clause](https://choosealicense.com/licenses/bsd-2-clause) |
| [`github.com/VerticalOps/fakesentry`](https://github.com/VerticalOps/fakesentry) | Run a fake Sentry server.       | [MIT](https://choosealicense.com/licenses/mit/)                  |
| [`github.com/getsentry/sentry-go`](https://github.com/getsentry/sentry-go)       | Sentry SDK for error reporting. | [BSD 2-Clause](https://choosealicense.com/licenses/bsd-2-clause) |

<!-- Links -->

[taegis-xdr]: https://www.secureworks.com/products/taegis/xdr
[taegis-sdk]: https://github.com/secureworks/taegis-sdk-go
[godocs]: https://pkg.go.dev/github.com/secureworks/logger
[logrus]: https://github.com/sirupsen/logrus
[zerolog]: https://github.com/rs/zerolog
[sentry]: https://docs.sentry.io/platforms/go/
[apache-2]: https://choosealicense.com/licenses/apache-2.0/
[unsafe]: https://pkg.go.dev/unsafe
