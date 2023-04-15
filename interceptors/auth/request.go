// Copyright (c) The go-grpc-middleware Authors.
// Licensed under the Apache License 2.0.

package auth

import (
	"errors"
	"strings"

	"github.com/bufbuild/connect-go"
)

var (
	headerAuthorize = "authorization"
)

// FromRequest is a helper function for extracting the :authorization header from the connect request.
//
// It expects the `:authorization` header to be of a certain scheme (e.g. `basic`, `bearer`), in a
// case-insensitive format (see rfc2617, sec 1.2). If no such authorization is found, or the token
// is of wrong scheme, an error with connect status `Unauthenticated` is returned.
func FromRequest(req connect.AnyRequest, expectedScheme string) (string, error) {
	authHeader := req.Header().Get(headerAuthorize)
	if authHeader == "" {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("request unauthenticated with "+expectedScheme))
	}
	splits := strings.SplitN(authHeader, " ", 2)
	if len(splits) < 2 {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("bad authorization string"))

	}
	if !strings.EqualFold(splits[0], expectedScheme) {
		return "", connect.NewError(connect.CodeUnauthenticated, errors.New("request unauthenticated with "+expectedScheme))
	}
	return splits[1], nil
}
