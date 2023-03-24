package bootstrap

import (
	"authz/domain/contracts"
	"authz/domain/model"
	"authz/infrastructure/repository/authzed"
	"authz/infrastructure/repository/fga"
	"authz/infrastructure/repository/mock"
)

// AccessRepositoryBuilder is the builder containing the config for building technical implementations of the server
type AccessRepositoryBuilder struct {
	impl string
}

// NewAccessRepositoryBuilder returns a new AccessRepositoryBuilder instance
func NewAccessRepositoryBuilder() *AccessRepositoryBuilder {
	return &AccessRepositoryBuilder{}
}

// WithImplementation defines the impl of the accessRepository to use
func (e *AccessRepositoryBuilder) WithImplementation(implID string) *AccessRepositoryBuilder {
	e.impl = implID
	return e
}

// Build builds an implementation based on the given param
func (e *AccessRepositoryBuilder) Build() (contracts.AccessRepository, error) {
	switch e.impl {
	case "stub":
		return &mock.StubAccessRepository{Data: getMockData(), LicensedSeats: map[string]map[string]model.License{}}, nil
	case "spicedb":
		return &authzed.SpiceDbAccessRepository{}, nil
	case "openfga":
		return &fga.OpenFgaAccessRepository{}, nil
	default:
		return &mock.StubAccessRepository{Data: getMockData(), LicensedSeats: map[string]map[string]model.License{}}, nil
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
