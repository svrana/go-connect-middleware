// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package logging

import (
	"context"
	"fmt"
	"time"

	"github.com/bufbuild/connect-go"
)

// LoggableEvent defines the events a log line can be added on.
type LoggableEvent uint

const (
	// StartCall is a loggable event representing start of the connect call.
	StartCall LoggableEvent = iota
	// FinishCall is a loggable event representing finish of the connect call.
	FinishCall
	// PayloadReceived is a loggable event representing received request (server) or response (client).
	// Log line for this event also includes (potentially big) proto.Message of that payload in
	// "grpc.request.content" (server) or "grpc.response.content" (client) field.
	// NOTE: This can get quite verbose, especially for streaming calls, use with caution (e.g. debug only purposes).
	PayloadReceived
	// PayloadSent is a loggable event representing sent response (server) or request (client).
	// Log line for this event also includes (potentially big) proto.Message of that payload in
	// "grpc.response.content" (server) or "grpc.request.content" (client) field.
	// NOTE: This can get quite verbose, especially for streaming calls, use with caution (e.g. debug only purposes).
	PayloadSent
)

func has(events []LoggableEvent, event LoggableEvent) bool {
	for _, e := range events {
		if e == event {
			return true
		}
	}
	return false
}

var (
	defaultOptions = &options{
		loggableEvents:    []LoggableEvent{StartCall, FinishCall},
		codeFunc:          DefaultErrorToCode,
		durationFieldFunc: DefaultDurationToFields,
		// levelFunc depends if it's client or server.
		levelFunc:       nil,
		timestampFormat: time.RFC3339,
	}
)

type options struct {
	levelFunc         CodeToLevel
	loggableEvents    []LoggableEvent
	codeFunc          ErrorToCode
	durationFieldFunc DurationToFields
	timestampFormat   string
	fieldsFromCtxFn   fieldsFromCtxFn
}

type Option func(*options)

func evaluateServerOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultServerCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func evaluateClientOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultClientCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

// DurationToFields function defines how to produce duration fields for logging.
type DurationToFields func(duration time.Duration) Fields

// ErrorToCode function determines the error code of an error.
// This makes using custom errors with grpc middleware easier.
type ErrorToCode func(err error) connect.Code

func DefaultErrorToCode(err error) connect.Code {
	return connect.CodeOf(err)
}

// CodeToLevel function defines the mapping between gRPC return connect and interceptor log level.
type CodeToLevel func(code connect.Code) Level

// DefaultServerCodeToLevel is the helper mapper that maps gRPC return connect to log levels for server side.
func DefaultServerCodeToLevel(code connect.Code) Level {
	switch code {
	case connect.CodeNotFound, connect.CodeCanceled, connect.CodeAlreadyExists, connect.CodeInvalidArgument, connect.CodeUnauthenticated:
		return LevelInfo

	case connect.CodeDeadlineExceeded, connect.CodePermissionDenied, connect.CodeResourceExhausted, connect.CodeFailedPrecondition, connect.CodeAborted,
		connect.CodeOutOfRange, connect.CodeUnavailable:
		return LevelWarn

	case connect.CodeUnknown, connect.CodeUnimplemented, connect.CodeInternal, connect.CodeDataLoss:
		return LevelError

	default:
		return LevelError
	}
}

// DefaultClientCodeToLevel is the helper mapper that maps gRPC return connect to log levels for client side.
func DefaultClientCodeToLevel(code connect.Code) Level {
	switch code {
	case connect.CodeCanceled, connect.CodeInvalidArgument, connect.CodeNotFound, connect.CodeAlreadyExists, connect.CodeResourceExhausted,
		connect.CodeFailedPrecondition, connect.CodeAborted, connect.CodeOutOfRange:
		return LevelDebug
	case connect.CodeUnknown, connect.CodeDeadlineExceeded, connect.CodePermissionDenied, connect.CodeUnauthenticated:
		return LevelInfo
	case connect.CodeUnimplemented, connect.CodeInternal, connect.CodeUnavailable, connect.CodeDataLoss:
		return LevelWarn
	default:
		return LevelInfo
	}
}

type fieldsFromCtxFn func(ctx context.Context) Fields

// WithFieldsFromContext allows adding extra fields to all log messages per given request.
func WithFieldsFromContext(f fieldsFromCtxFn) Option {
	return func(o *options) {
		o.fieldsFromCtxFn = f
	}
}

// WithLogOnEvents customizes on what events the gRPC interceptor should log on.
func WithLogOnEvents(events ...LoggableEvent) Option {
	return func(o *options) {
		o.loggableEvents = events
	}
}

// WithLevels customizes the function for mapping gRPC return connect and interceptor log level statements.
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.levelFunc = f
	}
}

// Withconnect customizes the function for mapping errors to error connect.
func Withconnect(f ErrorToCode) Option {
	return func(o *options) {
		o.codeFunc = f
	}
}

// WithDurationField customizes the function for mapping request durations to log fields.
func WithDurationField(f DurationToFields) Option {
	return func(o *options) {
		o.durationFieldFunc = f
	}
}

// DefaultDurationToFields is the default implementation of converting request duration to a field.
var DefaultDurationToFields = DurationToTimeMillisFields

// DurationToTimeMillisFields converts the duration to milliseconds and uses the key `grpc.time_ms`.
func DurationToTimeMillisFields(duration time.Duration) Fields {
	return Fields{"grpc.time_ms", fmt.Sprintf("%v", durationToMilliseconds(duration))}
}

// DurationToDurationField uses a Duration field to log the request duration
// and leaves it up to Log's encoder settings to determine how that is output.
func DurationToDurationField(duration time.Duration) Fields {
	return Fields{"grpc.duration", duration.String()}
}

func durationToMilliseconds(duration time.Duration) float32 {
	return float32(duration.Nanoseconds()/1000) / 1000
}

// WithTimestampFormat customizes the timestamps emitted in the log fields.
func WithTimestampFormat(format string) Option {
	return func(o *options) {
		o.timestampFormat = format
	}
}
