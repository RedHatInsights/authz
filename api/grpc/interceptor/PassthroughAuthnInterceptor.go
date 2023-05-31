package interceptor

import (
	"context"

	"google.golang.org/grpc"
)

// PassthroughAuthnInterceptor - passthrough AuthN Middleware for usage in local dev use cases where authN does not matter.
type PassthroughAuthnInterceptor struct{}

// NewPassthroughAuthnInterceptor creates a new PassthroughAuthnInterceptor
func NewPassthroughAuthnInterceptor() *PassthroughAuthnInterceptor {
	return &PassthroughAuthnInterceptor{}
}

// Unary impl of the Unary passthrough interceptor, returning a static value in the context..
func (authnInterceptor *PassthroughAuthnInterceptor) Unary() grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		return handler(context.WithValue(ctx, RequestorContextKey, "static-subject"), req)
	})
}

func (AuthnInterceptor *PassthroughAuthnInterceptor) Stream() grpc.ServerOption {
	return grpc.StreamInterceptor(func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrappedStream := AuthnServerStream{}
		wrappedStream.ServerStream = ss
		wrappedStream.ctx = context.WithValue(ss.Context(), RequestorContextKey, "static-subject")

		return handler(srv, wrappedStream)
	})
}
