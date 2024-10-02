# Secureworks Unified Logging Library

`secureworks/logger` is a unified interface that wraps popular logging
libraries such as [Logrus][logrus] and [Zerolog][zerolog]: _and that is just
the beginning!_

This is the logging library used in
[SecureWorks Taegis™ XDR (Extended Detection and Response)][taegis-xdr] Cloud
Platform, and is suggested for use with the [Taegis Golang SDK][taegis-sdk].

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
    - We have broken the packages up in order to keep dependencies in line with the log implementations. If you want 
      `zerolog` you shouldn't also need `logrus`; if you want to write code that consumes the shared interface you 
      shouldn't need to depend on either implementation. 
- There are some packages with "safe" and "unsafe" versions of code. Why is this?
    - *unsafe* refers to using [the Go standard library `unsafe`][unsafe], which allows us to step outside of Go's type-safety rules. This code is no more "not safe" than a typical C program.
    - While we use the unsafe code (less type-safe) by default, this can be disabled by adding a `safe` or `!unsafe` build tag. This may be useful if you are building for an environment that does not allow unsafe (less type-safe) code.
    - For `zerolog` and `logrus` the unsafe code is used for a big performance boost.
    - For `zerolog` it also addresses a small behavior change in the `zerolog.Hook` interface. **[See this issue for more.](https://github.com/rs/zerolog/issues/408)**

## License

This library is distributed under the [Apache-2.0 license][apache-2] found in
the [LICENSE](./LICENSE) file.

### Runtime Dependencies

| Library                                                                    | Purpose                         | License                                                          |
|----------------------------------------------------------------------------|---------------------------------|------------------------------------------------------------------|
| [`github.com/secureworks/errors`](https://github.com/secureworks/errors)   | Extracts error stack traces.    | [BSD 2-Clause](https://choosealicense.com/licenses/bsd-2-clause) |
| [`github.com/rs/zerolog`](https://github.com/rs/zerolog)                   | Logger.                         | [MIT](https://choosealicense.com/licenses/mit/)                  |
| [`github.com/sirupsen/logrus`](https://github.com/sirupsen/logrus)         | Logger.                         | [MIT](https://choosealicense.com/licenses/mit/)                  |

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
>   - [`github.com/secureworks/errors`](https://github.com/secureworks/errors)
>   - [`github.com/sirupsen/logrus`](https://github.com/sirupsen/logrus)
> - `github.com/secureworks/logger/zerolog`:
>   - [`github.com/rs/zerolog`](https://github.com/rs/zerolog)

### Test Dependencies

| Library                                                  | Purpose                         | License                                                          |
|----------------------------------------------------------|---------------------------------|------------------------------------------------------------------|
| [`github.com/pkg/errors`](https://github.com/pkg/errors) | Extracts error stack traces.    | [BSD 2-Clause](https://choosealicense.com/licenses/bsd-2-clause) |

<!-- Links -->

[taegis-xdr]: https://www.secureworks.com/products/taegis/xdr
[taegis-sdk]: https://github.com/secureworks/taegis-sdk-go
[godocs]: https://pkg.go.dev/github.com/secureworks/logger
[logrus]: https://github.com/sirupsen/logrus
[zerolog]: https://github.com/rs/zerolog
[apache-2]: https://choosealicense.com/licenses/apache-2.0/
[unsafe]: https://pkg.go.dev/unsafe
