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

// Unary impl of the Unary interceptor, returning the value of the authorization header directly.
func (authnInterceptor *PassthroughAuthnInterceptor) Unary() grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {

		return handler(context.WithValue(ctx, RequestorContextKey, getBearerTokenFromContext(ctx)), req)
	})
}
