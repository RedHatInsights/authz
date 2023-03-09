package server

import (
	"authz/domain/contracts"
)

// SrvBuilder is the builder containing the config for building technical implementations of the server
type SrvBuilder struct {
	framework string
}

// NewBuilder returns a new SrvBuilder instance
func NewBuilder() *SrvBuilder {
	return &SrvBuilder{}
}

// WithFramework sets the actual technical impl
func (s *SrvBuilder) WithFramework(fw string) *SrvBuilder {
	s.framework = fw
	return s
}

// Build builds an implementation based on the given param
func (s *SrvBuilder) Build() (contracts.Server, error) {
	switch s.framework {
	case "gin":
		return new(GinServer), nil
	case "echo":
		return new(EchoServer), nil
	}

	return nil, nil
}
