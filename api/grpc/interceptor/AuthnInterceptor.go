// Package interceptor contains middleware interceptors for unary and stream. Interceptors are applied to calls from HTTP and GRPC
package interceptor

import (
	"authz/api"
	"context"
	"strings"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthnInterceptor - Middleware to validate incoming bearer tokens
type AuthnInterceptor struct {
	provider providerJSON
}

// ContextKey Type to hold Keys that are applied to the request context
type ContextKey string

const (
	// RequestorContextKey Key for the Requestor value
	RequestorContextKey ContextKey = ContextKey("Requestor")
)

type providerJSON struct {
	Issuer  string `json:"issuer"`
	JwksUrl string `json:"jwks_uri"`
}

// NewAuthnInterceptor constructor
func NewAuthnInterceptor(config api.AuthConfig) *AuthnInterceptor {
	//req, err := http.NewRequest("GET", config.DiscoveryEndpoint, nil)
	//resp, err := http.DefaultClient.Do(req)

	// TODO: jwks etc. initialization goes here
	return &AuthnInterceptor{}
}

// Unary impl of the Unary interceptor
func (r *AuthnInterceptor) Unary() grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		glog.Info("Hello from AuthnInterceptor %v: ", req)
		token := getBearerTokenFromContext(ctx)

		if token != "" {
			glog.Infof("Received placeholder token: %s", token) //Obvs remove
		} else {
			glog.Info("No bearer token received")
		}

		return handler(context.WithValue(ctx, RequestorContextKey, token), req)
	})
}

func getBearerTokenFromContext(ctx context.Context) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for _, name := range []string{"grpcgateway-authorization", "authorization"} {
			headers := md.Get(name)
			if len(headers) > 0 {
				value := headers[0]
				parts := strings.Split(value, " ")

				if len(parts) > 1 {
					return parts[1]
				}

				return parts[0]
			}
		}
	}
	return ""
}

func validateBearerToken() error {

	//Token validation

	//Token expiry check

	// JWKS - issuer verification

	// extract needed info

	// return

	return nil
}
