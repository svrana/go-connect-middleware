// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package zap_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	"github.com/svrana/go-connect-middleware/interceptors/logging"
	"go.uber.org/zap"
)

// InterceptorLogger adapts zap logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func InterceptorLogger(l *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)
		i := logging.Fields(fields).Iterator()
		for i.Next() {
			k, v := i.At()
			f = append(f, zap.Any(k, v))
		}
		l = l.WithOptions(zap.AddCallerSkip(1)).With(f...)

		switch lvl {
		case logging.LevelDebug:
			l.Debug(msg)
		case logging.LevelInfo:
			l.Info(msg)
		case logging.LevelWarn:
			l.Warn(msg)
		case logging.LevelError:
			l.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func ExampleInterceptorLogger() {
	logger := zap.NewExample()

	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
		// Add any other option (check functions starting with logging.With).
	}

	interceptors := connect.WithInterceptors(
		logging.UnaryServerInterceptor(InterceptorLogger(logger), opts...),
		// And any other interceptors you want
	)

	mux := http.NewServeMux()
	mux.Handle(
		// imaginary user service
		pingv1connect.NewPingServiceHandler(
			pingServer{},
			interceptors, // your middleware here
		),
	)
}
