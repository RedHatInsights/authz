package bootstrap

import (
	"authz/domain/contracts"
	"authz/domain/model"
	vo "authz/domain/valueobjects"
	"authz/infrastructure/repository/authzed"
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
		return &mock.StubAccessRepository{Data: getMockData(), LicensedSeats: map[string]map[vo.SubjectID]bool{}, Licenses: getMockLicenseData()}, nil
	case "spicedb":
		return &authzed.SpiceDbAccessRepository{}, nil
	default:
		return &mock.StubAccessRepository{Data: getMockData(), LicensedSeats: map[string]map[vo.SubjectID]bool{}, Licenses: getMockLicenseData()}, nil
	}
}

func getMockData() map[vo.SubjectID]bool {
	return map[vo.SubjectID]bool{
		"token": true,
		"alice": true,
		"bob":   true,
		"chuck": false,
	}
}

func getMockLicenseData() map[string]model.License {
	return map[string]model.License{"smarts": *model.NewLicense("aspian", "smarts", 20, 0)}
}
