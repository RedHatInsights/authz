package bootstrap

import (
	"authz/api"
	"authz/api/grpc"
	"authz/api/http"
	"authz/application"
	"authz/domain/contracts"
)

// ServerBuilder is the builder containing the config for building technical implementations of the server
type ServerBuilder struct {
	PrincipalRepository contracts.PrincipalRepository
	AccessAppService    *application.AccessAppService
	LicenseAppService   *application.LicenseAppService
	ServerConfig        *api.ServerConfig
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

// WithLicenseAppService sets the LicenseAppService for the server
func (s *ServerBuilder) WithLicenseAppService(sh *application.LicenseAppService) *ServerBuilder {
	s.LicenseAppService = sh
	return s
}

// WithServerConfig sets the ServerConfig configuration for the used server.
func (s *ServerBuilder) WithServerConfig(c *api.ServerConfig) *ServerBuilder {
	s.ServerConfig = c
	return s
}

// BuildGrpc builds the grpc-server of the grpc gateway
func (s *ServerBuilder) BuildGrpc() (srv *grpc.Server, err error) {
	return grpc.NewServer(*s.AccessAppService, *s.LicenseAppService, *s.ServerConfig), nil
}

// BuildHTTP builds the HTTP Server of the grpc gateway
func (s *ServerBuilder) BuildHTTP() (srv *http.Server, err error) {
	return http.NewServer(*s.ServerConfig), nil
}
