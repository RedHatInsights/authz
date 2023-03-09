package app

import (
	"authz/domain/contracts"
	"authz/infrastructure/engine"
	"authz/infrastructure/server"
)

// AppBuilder is the builder containing the config for building technical implementations of the server
type AppBuilder struct {
	framework string
	engine    string
}

// NewAppBuilder returns a new AppBuilder instance
func NewAppBuilder() *AppBuilder {
	return &AppBuilder{}
}

// WithFramework sets the actual technical impl
func (s *AppBuilder) WithFramework(fw string) *AppBuilder {
	s.framework = fw
	return s
}

func (s *AppBuilder) WithEngine(e string) *AppBuilder {
	s.engine = e
	return s
}

// Build builds an implementation based on the given param
func (s *AppBuilder) Build() (contracts.Server, error) {
	switch s.framework {
	case "gin":
		return server.GinServer{}.NewServer(), nil
	case "echo":
		return server.EchoServer{}.NewServer(), nil
	case "grpc":
		var srv = server.GrpcGatewayServer{}
		return srv.NewServer(), nil
	default:
		return nil, nil
	}
}

func getEngine(e string) contracts.AuthzEngine {
	switch e {
	case "stub":
		return engine.StubAuthzEngine{Data: getMockData()}.NewEngine()
	case "spicedb":
		return engine.SpiceDbAuthzEngine{}.NewEngine()
	default:
		return engine.StubAuthzEngine{Data: getMockData()}.NewEngine()
	}
}

func getMockData() map[string]bool {
	return map[string]bool{
		"token": true,
		"alice": true,
		"bob":   true,
		"chuck": false,
	}
}
