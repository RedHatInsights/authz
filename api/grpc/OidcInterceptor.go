package grpc

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-jose/go-jose"
	jwt "github.com/go-jose/go-jose/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TODO: all the comments
const (
	RequestorKey string = "requestor"
)

type OidcInterceptor struct {
	jwks           *jose.JSONWebKeySet
	requiredScopes []string
}

type tokenData struct {
	Subject string   `json:"sub,omitempty"`
	Scopes  []string `json:"scopes,omitempty"`
}

func NewOidcInterceptor(jwks *jose.JSONWebKeySet, requiredScopes []string) *OidcInterceptor {
	return &OidcInterceptor{jwks: jwks, requiredScopes: requiredScopes}
}

func (i *OidcInterceptor) Unary() grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		requestor, err := i.getRequestingSubject(ctx)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, fmt.Errorf("Error processing bearer token: %w", err).Error())
		}

		return handler(context.WithValue(ctx, RequestorKey, requestor), req) //TODO: what type should the key be?
	})
}

func (i *OidcInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, ss)
	}
}

func (i *OidcInterceptor) getRequestingSubject(ctx context.Context) (string, error) {
	bearer := getBearerTokenFromContext(ctx)
	if bearer != "" {
		token, err := jwt.ParseSigned(bearer)
		if err != nil {
			return "", err
		}

		data := tokenData{}
		err = token.Claims(i.jwks, &data)
		if err != nil {
			return "", err
		}

		err = i.validateScopes(data)
		if err != nil {
			return "", err
		}

		return data.Subject, nil
	} else {
		return "", nil
	}
}

func (i *OidcInterceptor) validateScopes(data tokenData) error {
OUTER:
	for _, required := range i.requiredScopes { //TODO: consider set/subset for O(n) time vs O(n^2) inner loops. Though both slices should be very small..
		for _, provided := range data.Scopes {
			if required == provided {
				continue OUTER
			}
		}

		return fmt.Errorf("Required scope not found: %s", required)
	}

	return nil
}

func getBearerTokenFromContext(ctx context.Context) string {
	for _, name := range []string{"grpcgateway-authorization"} {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			headers := md.Get(name)
			if len(headers) > 0 {
				value := headers[0]
				parts := strings.Split(value, " ")

				if len(parts) > 1 {
					return parts[1]
				} else {
					return parts[0]
				}
			}
		}
	}

	return ""
}
