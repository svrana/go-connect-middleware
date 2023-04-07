// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package interceptors

import (
	"fmt"
	"strings"

	"github.com/bufbuild/connect-go"
)

func splitFullMethodName(fullMethod string) (string, string) {
	fullMethod = strings.TrimPrefix(fullMethod, "/") // remove leading slash
	if i := strings.Index(fullMethod, "/"); i >= 0 {
		return fullMethod[:i], fullMethod[i+1:]
	}
	return "unknown", "unknown"
}

type CallMeta struct {
	ReqOrNil any
	Typ      connect.StreamType
	Service  string
	Method   string
	IsClient bool
}

func NewServerCallMeta(spec connect.Spec, reqOrNil any) CallMeta {
	c := CallMeta{IsClient: spec.IsClient, ReqOrNil: reqOrNil, Typ: spec.StreamType}
	c.Service, c.Method = splitFullMethodName(spec.Procedure)
	return c
}
func (c CallMeta) FullMethod() string {
	return fmt.Sprintf("/%s/%s", c.Service, c.Method)
}
