# Go Connect Middleware

This reports holds [connect Go](https://github.com/bufbuild/connect-go) Middlewares: interceptors, helpers and utilities.

This is a derived work of [go gRPC middleware](https://github.com/grpc-ecosystem/go-grpc-middleware).

## Interceptors

#### Observability

- Logging with [`github.com/svrana/go-connect-middleware//interceptors/logging`](interceptors/logging) - a customizable logging middleware offering extended per request logging. It requires logging adapter, see examples in [`interceptors/logging/examples`](interceptors/logging/examples) for `go-kit`, `log`, `logr`, `logrus`, `slog`, `zap` and `zerolog`.
  (Only unary server interceptor for now)

## Prerequisites

- **[Go](https://golang.org)**: Any one of the **three latest major** [releases](https://golang.org/doc/devel/release.html) are supported.

## License

`go-grpc-middleware` is released under the Apache 2.0 license. See the [LICENSE](LICENSE) file for details.
