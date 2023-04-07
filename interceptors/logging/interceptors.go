// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logging

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/bufbuild/connect-go"
	"google.golang.org/grpc/peer"

	"github.com/svrana/go-connect-middleware/interceptors"
)

type reporter struct {
	interceptors.CallMeta

	ctx             context.Context
	kind            string
	startCallLogged bool

	opts   *options
	fields Fields
	logger Logger
}

func (c *reporter) PostCall(err error, duration time.Duration) {
	if !has(c.opts.loggableEvents, FinishCall) {
		return
	}
	if err == io.EOF {
		err = nil
	}

	code := c.opts.codeFunc(err)
	fields := c.fields.WithUnique(ExtractFields(c.ctx))
	fields = fields.AppendUnique(Fields{"connect.code", code.String()})
	if err != nil {
		fields = fields.AppendUnique(Fields{"connect.error", fmt.Sprintf("%v", err)})
	}
	c.logger.Log(c.ctx, c.opts.levelFunc(code), "finished call", fields.AppendUnique(c.opts.durationFieldFunc(duration))...)
}

func (c *reporter) PostMsgSend(res connect.AnyResponse, err error, duration time.Duration) {
	logLvl := c.opts.levelFunc(c.opts.codeFunc(err))
	fields := c.fields.WithUnique(ExtractFields(c.ctx))
	if err != nil {
		fields = fields.AppendUnique(Fields{"connect.error", fmt.Sprintf("%v", err)})
	}
	if !c.startCallLogged && has(c.opts.loggableEvents, StartCall) {
		c.startCallLogged = true
		c.logger.Log(c.ctx, logLvl, "started call", fields.AppendUnique(c.opts.durationFieldFunc(duration))...)
	}

	if err != nil || !has(c.opts.loggableEvents, PayloadSent) {
		return
	}
	if c.CallMeta.IsClient {
		fields = fields.AppendUnique(Fields{"connect.send.duration", duration.String(), "connect.request.content", res})
		c.logger.Log(c.ctx, logLvl, "request sent", fields...)
	} else {
		fields = fields.AppendUnique(Fields{"connect.send.duration", duration.String(), "connect.response.content", res})
		c.logger.Log(c.ctx, logLvl, "response sent", fields...)
	}
}

func (c *reporter) PostMsgReceive(req connect.AnyRequest, err error, duration time.Duration) {
	logLvl := c.opts.levelFunc(c.opts.codeFunc(err))
	fields := c.fields.WithUnique(ExtractFields(c.ctx))
	if err != nil {
		fields = fields.AppendUnique(Fields{"connect.error", fmt.Sprintf("%v", err)})
	}
	if !c.startCallLogged && has(c.opts.loggableEvents, StartCall) {
		c.startCallLogged = true
		c.logger.Log(c.ctx, logLvl, "started call", fields.AppendUnique(c.opts.durationFieldFunc(duration))...)
	}

	if err != nil || !has(c.opts.loggableEvents, PayloadReceived) {
		return
	}
	if !c.CallMeta.IsClient {
		fields = fields.AppendUnique(Fields{"connect.recv.duration", duration.String(), "connect.request.content", req})
		c.logger.Log(c.ctx, logLvl, "request received", fields...)
	} else {
		fields = fields.AppendUnique(Fields{"connect.recv.duration", duration.String(), "connect.response.content", req})
		c.logger.Log(c.ctx, logLvl, "response received", fields...)
	}
}

func reportable(logger Logger, opts *options) interceptors.CommonReportableFunc {
	return func(ctx context.Context, c interceptors.CallMeta) (interceptors.Reporter, context.Context) {
		kind := KindServerFieldValue
		if c.IsClient {
			kind = KindClientFieldValue
		}

		fields := ExtractFields(ctx).WithUnique(newCommonFields(kind, c))
		if !c.IsClient {
			// FIXME
			if peer, ok := peer.FromContext(ctx); ok {
				fields = append(fields, "peer.address", peer.Addr.String())
			}
		}
		if opts.fieldsFromCtxFn != nil {
			fields = fields.AppendUnique(opts.fieldsFromCtxFn(ctx))
		}

		singleUseFields := Fields{"connect.start_time", time.Now().Format(opts.timestampFormat)}
		if d, ok := ctx.Deadline(); ok {
			singleUseFields = singleUseFields.AppendUnique(Fields{"connect.request.deadline", d.Format(opts.timestampFormat)})
		}
		return &reporter{
			CallMeta:        c,
			ctx:             ctx,
			startCallLogged: false,
			opts:            opts,
			fields:          fields.WithUnique(singleUseFields),
			logger:          logger,
			kind:            kind,
		}, InjectFields(ctx, fields)
	}
}

// UnaryServerInterceptor returns a new unary server interceptors that optionally logs endpoint handling.
// Logger will read existing and write new logging.Fields available in current context.
// See `ExtractFields` and `InjectFields` for details.
func UnaryServerInterceptor(logger Logger, opts ...Option) connect.UnaryInterceptorFunc {
	o := evaluateServerOpt(opts)
	return interceptors.UnaryServerInterceptor(reportable(logger, o))
}
