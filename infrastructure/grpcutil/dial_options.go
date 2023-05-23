// Package grpcutil implements various utilities to simplify common gRPC APIs.
package grpcutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/certifi/gocertifi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type verification int

const (
	// SkipVerifyCA is a constant that improves the readability of functions
	// with the insecureSkipVerify parameter.
	SkipVerifyCA verification = iota

	// VerifyCA is a constant that improves the readability of functions
	// with the insecureSkipVerify parameter.
	VerifyCA
)

func (v verification) asInsecureSkipVerify() bool {
	switch v {
	case SkipVerifyCA:
		return true
	case VerifyCA:
		return false
	default:
		panic("unknown verification")
	}
}

// WithSystemCerts returns a grpc.DialOption that uses the system-provided
// certificate authority chain to verify the connection.
//
// If one cannot be found, this falls back to using a vendored version of
// Mozilla's collection of root certificate authorities.
func WithSystemCerts(v verification) (grpc.DialOption, error) {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		// Fall back to Mozilla collection of root CAs.
		certPool, err = gocertifi.CACerts()
		if err != nil {
			// This library promises that this should never occur.
			return nil, fmt.Errorf("gocertifi returned an error: %w", err)
		}
	}

	return grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		RootCAs:            certPool,
		InsecureSkipVerify: v.asInsecureSkipVerify(),
	})), nil
}

type secureMetadataCreds map[string]string

func (c secureMetadataCreds) RequireTransportSecurity() bool { return true }
func (c secureMetadataCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return c, nil
}

// WithBearerToken returns a grpc.DialOption that adds a standard HTTP Bearer
// token to all requests sent from a client.
func WithBearerToken(token string) grpc.DialOption {
	return grpc.WithPerRPCCredentials(secureMetadataCreds{"authorization": "Bearer " + token})
}

type insecureMetadataCreds map[string]string

func (c insecureMetadataCreds) RequireTransportSecurity() bool { return false }
func (c insecureMetadataCreds) GetRequestMetadata(_ context.Context, _ ...string) (map[string]string, error) {
	return c, nil
}

// WithInsecureBearerToken returns a grpc.DialOption that adds a standard HTTP
// Bearer token to all requests sent from an insecure client.
//
// Must be used in conjunction with `insecure.NewCredentials()`.
func WithInsecureBearerToken(token string) grpc.DialOption {
	return grpc.WithPerRPCCredentials(insecureMetadataCreds{"authorization": "Bearer " + token})
}
