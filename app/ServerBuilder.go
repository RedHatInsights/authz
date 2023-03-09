package app

import (
	"authz/domain/contracts"
	"authz/infrastructure/server"
)

// ServerBuilder is the builder containing the config for building technical implementations of the server
type ServerBuilder struct {
	framework string
}

// NewServerBuilder returns a new ServerBuilder instance
func NewServerBuilder() *ServerBuilder {
	return &ServerBuilder{}
}

// WithFramework sets the actual technical impl
func (s *ServerBuilder) WithFramework(fw string) *ServerBuilder {
	s.framework = fw
	return s
}

// Build builds an implementation based on the given param
func (s *ServerBuilder) Build() (contracts.Server, error) {
	switch s.framework {
	case "gin":
		return server.GinServer{}.NewServer(), nil
	case "echo":
		return server.EchoServer{}.NewServer(), nil
	case "grpc":
		var srv = server.GrpcGatewayServer{}
		return srv.NewServer(), nil
	case "grpcweb":
		webServer := server.GrpcWebServer{}
		return webServer.NewServer(), nil
	default:
		return nil, nil
	}
}
