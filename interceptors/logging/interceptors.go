// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logging

import (
	"context"
	"fmt"
	"io"
	"time"

	"connectrpc.com/connect"
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

	fields := c.fields.WithUnique(ExtractFields(c.ctx))

	var level Level
	if err != nil {
		code := c.opts.codeFunc(err)
		fields = fields.AppendUnique(Fields{"code", code.String()})
		level = c.opts.levelFunc(code)
	} else {
		level = LevelInfo
	}
	if err != nil {
		fields = fields.AppendUnique(Fields{"error", fmt.Sprintf("%v", err)})
	}
	c.logger.Log(c.ctx, level, "finished call", fields.AppendUnique(c.opts.durationFieldFunc(duration))...)
}

func (c *reporter) errorToLevel(err error) Level {
	if err != nil {
		return c.opts.levelFunc(c.opts.codeFunc(err))
	}
	return LevelInfo
}

func (c *reporter) PostMsgSend(res connect.AnyResponse, err error, duration time.Duration) {
	logLvl := c.errorToLevel(err)
	fields := c.fields.WithUnique(ExtractFields(c.ctx))
	if err != nil {
		fields = fields.AppendUnique(Fields{"error", fmt.Sprintf("%v", err)})
	}
	if !c.startCallLogged && has(c.opts.loggableEvents, StartCall) {
		c.startCallLogged = true
		c.logger.Log(c.ctx, logLvl, "started call", fields.AppendUnique(c.opts.durationFieldFunc(duration))...)
	}

	if err != nil || !has(c.opts.loggableEvents, PayloadSent) {
		return
	}
	if c.CallMeta.IsClient {
		fields = fields.AppendUnique(Fields{"send.duration", duration.String(), "request.content", res})
		c.logger.Log(c.ctx, logLvl, "request sent", fields...)
	} else {
		fields = fields.AppendUnique(Fields{"send.duration", duration.String(), "response.content", res})
		c.logger.Log(c.ctx, logLvl, "response sent", fields...)
	}
}

func (c *reporter) PostMsgReceive(req connect.AnyRequest, err error, duration time.Duration) {
	var logLvl Level
	if err != nil {
		logLvl = c.opts.levelFunc(c.opts.codeFunc(err))
	} else {
		logLvl = LevelInfo
	}
	fields := c.fields.WithUnique(ExtractFields(c.ctx))
	if err != nil {
		fields = fields.AppendUnique(Fields{"error", fmt.Sprintf("%v", err)})
	}
	if !c.startCallLogged && has(c.opts.loggableEvents, StartCall) {
		c.startCallLogged = true
		c.logger.Log(c.ctx, logLvl, "started call", fields.AppendUnique(c.opts.durationFieldFunc(duration))...)
	}

	if err != nil || !has(c.opts.loggableEvents, PayloadReceived) {
		return
	}
	if !c.CallMeta.IsClient {
		fields = fields.AppendUnique(Fields{"recv.duration", duration.String(), "request.content", req})
		c.logger.Log(c.ctx, logLvl, "request received", fields...)
	} else {
		fields = fields.AppendUnique(Fields{"recv.duration", duration.String(), "response.content", req})
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

		singleUseFields := Fields{"start_time", time.Now().Format(opts.timestampFormat)}
		if d, ok := ctx.Deadline(); ok {
			singleUseFields = singleUseFields.AppendUnique(Fields{"request.deadline", d.Format(opts.timestampFormat)})
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
