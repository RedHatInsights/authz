package bootstrap

import (
	"authz/api"
	"authz/domain"
	"authz/domain/contracts"
	"authz/infrastructure/repository/authzed"
	"authz/infrastructure/repository/mock"
)

var singletonSpiceDbRepository *authzed.SpiceDbAccessRepository

// AccessRepositoryBuilder is the builder containing the config for building technical implementations of the server
type AccessRepositoryBuilder struct {
	config *api.ServerConfig
}

// NewAccessRepositoryBuilder returns a new AccessRepositoryBuilder instance
func NewAccessRepositoryBuilder() *AccessRepositoryBuilder {
	return &AccessRepositoryBuilder{}
}

// WithConfig supplies a ServerConfig struct to be used as-needed for building objects
func (e *AccessRepositoryBuilder) WithConfig(config *api.ServerConfig) *AccessRepositoryBuilder {
	e.config = config
	return e
}

// Build builds an implementation based on the given param
func (e *AccessRepositoryBuilder) Build() (contracts.AccessRepository, error) {
	config := e.config.StoreConfig
	switch config.Store {
	case "stub":
		return &mock.StubAccessRepository{Data: getMockData(), LicensedSeats: map[string]map[domain.SubjectID]bool{}, Licenses: getMockLicenseData()}, nil
	case "spicedb":
		spicedb := &authzed.SpiceDbAccessRepository{}
		spicedb.NewConnection(config.Endpoint, config.AuthToken, true, config.UseTLS)
		singletonSpiceDbRepository = spicedb
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
