// Package app contains the application glue.
package app

import (
	"authz/domain/contracts"
	"authz/infrastructure/engine"
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
		return &engine.StubAuthzEngine{Data: getMockData()}, nil
	case "spicedb":
		return &engine.SpiceDbAuthzEngine{}, nil
	default:
		return &engine.StubAuthzEngine{Data: getMockData()}, nil
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
