// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

// Go gRPC Middleware monitoring interceptors for server-side gRPC.

package interceptors

import (
	"context"
	"time"

	"connectrpc.com/connect"
)

func UnaryServerInterceptor(reportable ServerReportable) connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			r := newReport(NewServerCallMeta(req.Spec(), req))
			reporter, newCtx := reportable.ServerReporter(ctx, r.callMeta)

			reporter.PostMsgReceive(req, nil, time.Since(r.startTime))
			resp, err := next(newCtx, req)
			reporter.PostMsgSend(resp, err, time.Since(r.startTime))

			reporter.PostCall(err, time.Since(r.startTime))
			return resp, err
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
