//go:build !local
// +build !local

package grpc

import (
	"fmt"

	"google.golang.org/grpc"
)

// createNoAuthnProvidersHandler is a compile-time variable method. The default version always returns an error.
func createNoAuthnProvidersHandler() (grpc.ServerOption, error) {
	return nil, fmt.Errorf("no active authentication configurations")
}
