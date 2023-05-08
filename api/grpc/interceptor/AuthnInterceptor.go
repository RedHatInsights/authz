package interceptor

import (
	"context"
	"fmt"
	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strings"
)

type AuthnInterceptor struct{}

type RequestorContextKey string

// NewAuthnInterceptor -
func NewAuthnInterceptor() *AuthnInterceptor {
	return &AuthnInterceptor{}
}

// Unary -
func (r *AuthnInterceptor) Unary() grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		glog.Info("Hello from AuthnInterceptor %v: ", req)
		token, err := getBearerTokenFromContext(ctx)

		if err != nil {
			glog.Error(err)
			return nil, err
		}

		if token != "" {
			glog.Infof("Received placeholder token: %s", token) //Obvs remove
		} else {
			glog.Info("No bearer token received")
		}

		key := RequestorContextKey("Requestor")
		return handler(context.WithValue(ctx, key, token), req)
	})
}

func getBearerTokenFromContext(ctx context.Context) (string, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for _, name := range []string{"grpcgateway-authorization", "authorization"} {
			headers := md.Get(name)
			if len(headers) > 0 {
				value := headers[0]
				parts := strings.Split(value, " ")

				if len(parts) > 1 {
					return parts[1], nil
				} else {
					return parts[0], nil
				}
			}
		}
	}
	return "", fmt.Errorf("bearer token not found")
}

func validateBearerToken() error {

	//Token validation

	//Token expiry check

	// JWKS - issuer verificaiton

	// extract needed info

	// return

	return nil
}
