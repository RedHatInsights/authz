package bootstrap

import (
	"authz/api/grpc"
	"authz/api/http"
	"authz/application"
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
)

// ServerBuilder is the builder containing the config for building technical implementations of the server
type ServerBuilder struct {
	PrincipalRepository contracts.PrincipalRepository
	AccessAppService    *application.AccessAppService
	ServiceConfig       *serviceconfig.ServiceConfig
}

// NewServerBuilder returns a new ServerBuilder instance
func NewServerBuilder() *ServerBuilder {
	return &ServerBuilder{}
}

// WithAccessAppService sets the AccessAppService for the server
func (s *ServerBuilder) WithAccessAppService(ph *application.AccessAppService) *ServerBuilder {
	s.AccessAppService = ph
	return s
}

// WithServiceConfig sets the ServiceConfig configuration for the used server.
func (s *ServerBuilder) WithServiceConfig(c *serviceconfig.ServiceConfig) *ServerBuilder {
	s.ServiceConfig = c
	return s
}

// BuildGrpc builds the grpc-server of the grpc gateway
func (s *ServerBuilder) BuildGrpc() (srv *grpc.Server, err error) {
	return grpc.NewServer(*s.AccessAppService, *s.ServiceConfig), nil
}

// BuildHTTP builds the HTTP Server of the grpc gateway
func (s *ServerBuilder) BuildHTTP() (srv *http.Server, err error) {
	return http.NewServer(*s.ServiceConfig), nil
}
