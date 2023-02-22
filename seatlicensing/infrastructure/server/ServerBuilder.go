package server

import (
	"authz/seatlicensing/domain/contracts"
)

type ServerBuilder struct {
	framework string
}

func NewBuilder() *ServerBuilder {
	return &ServerBuilder{}
}

// WithFramework sets the actual technical impl
func (s *ServerBuilder) WithFramework(fw string) *ServerBuilder {
	s.framework = fw
	return s
}

func (s *ServerBuilder) Build() (contracts.Server, error) {
	switch s.framework {
	case "gin":
		return new(GinServer), nil
	case "echo":
		return new(EchoServer), nil
	}

	return nil, nil
}
