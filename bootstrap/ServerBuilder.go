package bootstrap

import (
	"authz/api"
	"authz/api/grpc"
	"authz/api/http"
	"authz/application"
)

// ServerBuilder is the builder containing the config for building technical implementations of the server
type ServerBuilder struct {
	PermissionHandler *application.AccessAppService
	SeatHandler       *application.SeatAppService
	ServerConfig      *api.ServerConfig
}

// NewServerBuilder returns a new ServerBuilder instance
func NewServerBuilder() *ServerBuilder {
	return &ServerBuilder{}
}

// WithAccessAppService sets the PermissionHandler for the server
func (s *ServerBuilder) WithAccessAppService(ph *application.AccessAppService) *ServerBuilder {
	s.PermissionHandler = ph
	return s
}

// WithSeatAppService sets the SeatHandler for the server
func (s *ServerBuilder) WithSeatAppService(sh *application.SeatAppService) *ServerBuilder {
	s.SeatHandler = sh
	return s
}

// WithServerConfig sets the ServerConfig configuration for the used server.
func (s *ServerBuilder) WithServerConfig(c *api.ServerConfig) *ServerBuilder {
	s.ServerConfig = c
	return s
}

// BuildGrpc builds the grpc-server of the grpc gateway
func (s *ServerBuilder) BuildGrpc() (srv *grpc.Server, err error) {
	return grpc.NewServer(*s.PermissionHandler, *s.ServerConfig), nil
}

// BuildHTTP builds the HTTP Server of the grpc gateway
func (s *ServerBuilder) BuildHTTP() (srv *http.Server, err error) {
	return http.NewServer(*s.ServerConfig), nil
}
