// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package interceptors

import (
	"context"
	"time"

	"github.com/bufbuild/connect-go"
)

type ConnectType string

const (
	Unary        ConnectType = "unary"
	ClientStream ConnectType = "client_stream"
	ServerStream ConnectType = "server_stream"
	BidiStream   ConnectType = "bidi_stream"
)

var (
	AllCodes = []connect.Code{
		connect.CodeCanceled,
		connect.CodeUnknown,
		connect.CodeInvalidArgument,
		connect.CodeDeadlineExceeded,
		connect.CodeNotFound,
		connect.CodeAlreadyExists,
		connect.CodePermissionDenied,
		connect.CodeResourceExhausted,
		connect.CodeFailedPrecondition,
		connect.CodeAborted,
		connect.CodeOutOfRange,
		connect.CodeUnimplemented,
		connect.CodeInternal,
		connect.CodeUnavailable,
		connect.CodeDataLoss,
		connect.CodeUnauthenticated,
	}
)

type ClientReportable interface {
	ClientReporter(context.Context, CallMeta) (Reporter, context.Context)
}

type ServerReportable interface {
	ServerReporter(context.Context, CallMeta) (Reporter, context.Context)
}

// CommonReportableFunc helper allows an easy way to implement reporter with common client and server logic.
type CommonReportableFunc func(ctx context.Context, c CallMeta) (Reporter, context.Context)

func (f CommonReportableFunc) ClientReporter(ctx context.Context, c CallMeta) (Reporter, context.Context) {
	return f(ctx, c)
}

func (f CommonReportableFunc) ServerReporter(ctx context.Context, c CallMeta) (Reporter, context.Context) {
	return f(ctx, c)
}

type Reporter interface {
	PostCall(err error, rpcDuration time.Duration)
	PostMsgSend(res connect.AnyResponse, err error, sendDuration time.Duration)
	PostMsgReceive(req connect.AnyRequest, err error, recvDuration time.Duration)
}

var _ Reporter = NoopReporter{}

type NoopReporter struct{}

func (NoopReporter) PostCall(error, time.Duration)                           {}
func (NoopReporter) PostMsgSend(connect.AnyResponse, error, time.Duration)   {}
func (NoopReporter) PostMsgReceive(connect.AnyRequest, error, time.Duration) {}

type report struct {
	callMeta  CallMeta
	startTime time.Time
}

func newReport(callMeta CallMeta) report {
	r := report{
		startTime: time.Now(),
		callMeta:  callMeta,
	}
	return r
}
