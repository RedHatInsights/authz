package app

import (
	"authz/api/contracts"
	"authz/api/handler"
	"authz/api/server"
	"authz/app/config"
)

// ServerBuilder is the builder containing the config for building technical implementations of the server
type ServerBuilder struct {
	framework         string
	PermissionHandler *handler.PermissionHandler
	ServerConfig      *config.ServerConfig
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

// WithPermissionHandler sets the PermissionHandler for the server
func (s *ServerBuilder) WithPermissionHandler(h *handler.PermissionHandler) *ServerBuilder {
	s.PermissionHandler = h
	return s
}

// WithServerConfig sets the ServerConfig configuration for the used server.
func (s *ServerBuilder) WithServerConfig(c *config.ServerConfig) *ServerBuilder {
	s.ServerConfig = c
	return s
}

// Build builds an implementation based on the given param
func (s *ServerBuilder) Build() (contracts.Server, error) {
	switch s.framework {
	case "gin":
		var srv = server.GinServer{}
		return srv.NewServer(*s.PermissionHandler, *s.ServerConfig), nil
	case "echo":
		var srv = server.EchoServer{}
		return srv.NewServer(*s.PermissionHandler, *s.ServerConfig), nil
	case "grpc":
		var srv = server.GrpcGatewayServer{}
		return srv.NewServer(*s.PermissionHandler, *s.ServerConfig), nil
	case "grpcweb":
		webServer := server.GrpcWebServer{}
		return webServer.NewServer(*s.PermissionHandler, *s.ServerConfig), nil
	default:
		return nil, nil
	}
}
