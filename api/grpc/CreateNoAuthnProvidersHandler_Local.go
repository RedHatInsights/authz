//go:build local
// +build local

package grpc

import (
	"authz/api/grpc/interceptor"
	"github.com/golang/glog"
	"google.golang.org/grpc"
)

// createNoAuthnProvidersHandler is a compile-time variable method. The local version injects a new passthrough interceptor.
func createNoAuthnProvidersHandler() (grpc.ServerOption, error) {
	// local dev: no authconfig given, so we enable a passthrough middleware to get the requestor from authorization header.
	glog.Warning("Client authorization disabled. Do not use in production use cases!")
	return interceptor.NewPassthroughAuthnInterceptor().Unary(), nil
}
