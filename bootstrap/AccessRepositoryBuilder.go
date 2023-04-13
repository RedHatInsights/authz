package bootstrap

import (
	"authz/domain"
	"authz/domain/contracts"
	"authz/infrastructure/repository/authzed"
	"authz/infrastructure/repository/mock"
)

// AccessRepositoryBuilder is the builder containing the config for building technical implementations of the server
type AccessRepositoryBuilder struct {
	impl string

	endpoint  string
	authToken string
	useTLS    bool
}

// NewAccessRepositoryBuilder returns a new AccessRepositoryBuilder instance
func NewAccessRepositoryBuilder() *AccessRepositoryBuilder {
	return &AccessRepositoryBuilder{}
}

// WithConnectionInfo configures connection information to be used with applicable implementations
func (e *AccessRepositoryBuilder) WithConnectionInfo(endpoint string, authToken string, useTLS bool) *AccessRepositoryBuilder {
	e.endpoint = endpoint
	e.authToken = authToken
	e.useTLS = useTLS

	return e
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
		return &mock.StubAccessRepository{Data: getMockData(), LicensedSeats: map[string]map[domain.SubjectID]bool{}, Licenses: getMockLicenseData()}, nil
	case "spicedb":
		spicedb := &authzed.SpiceDbAccessRepository{}
		spicedb.NewConnection(e.endpoint, e.authToken, true, e.useTLS)
		return spicedb, nil
	default:
		return &mock.StubAccessRepository{Data: getMockData(), LicensedSeats: map[string]map[domain.SubjectID]bool{}, Licenses: getMockLicenseData()}, nil
	}
}

func getMockData() map[domain.SubjectID]bool {
	return map[domain.SubjectID]bool{
		"token": true,
		"alice": true,
		"bob":   true,
		"chuck": false,
	}
}

func getMockLicenseData() map[string]domain.License {
	return map[string]domain.License{"smarts": *domain.NewLicense("aspian", "smarts", 20, 0)}
}
