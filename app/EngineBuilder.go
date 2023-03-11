package app

import (
	"authz/domain/contracts"
	"authz/infrastructure/engine/authzed"
	"authz/infrastructure/engine/mock"
	"authz/infrastructure/engine/openfga"
)

// AuthzEngineBuilder is the builder containing the config for building technical implementations of the server
type AuthzEngineBuilder struct {
	engine string
}

// NewAuthzEngineBuilder returns a new AuthzEngineBuilder instance
func NewAuthzEngineBuilder() *AuthzEngineBuilder {
	return &AuthzEngineBuilder{}
}

// WithEngine defines the impl of the authzengine to use
func (e *AuthzEngineBuilder) WithEngine(engine string) *AuthzEngineBuilder {
	e.engine = engine
	return e
}

// Build builds an implementation based on the given param
func (e *AuthzEngineBuilder) Build() (contracts.AuthzEngine, error) {
	switch e.engine {
	case "stub":
		return &mock.StubAuthzEngine{Data: getMockData()}, nil
	case "spicedb":
		return &authzed.SpiceDbAuthzEngine{}, nil
	case "openfga":
		return &openfga.FgaAuthzEngine{}, nil
	default:
		return &mock.StubAuthzEngine{Data: getMockData()}, nil
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
