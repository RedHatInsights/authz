package bootstrap

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain"
	"authz/domain/contracts"
	"authz/infrastructure/repository/authzed"
	"authz/infrastructure/repository/mock"

	"github.com/golang/glog"
)

// AccessRepositoryBuilder is the builder containing the config for building technical implementations of the server
type AccessRepositoryBuilder struct {
	config *serviceconfig.ServiceConfig
}

// NewAccessRepositoryBuilder returns a new AccessRepositoryBuilder instance
func NewAccessRepositoryBuilder() *AccessRepositoryBuilder {
	return &AccessRepositoryBuilder{}
}

// WithConfig supplies a ServiceConfig struct to be used as-needed for building objects
func (e *AccessRepositoryBuilder) WithConfig(config *serviceconfig.ServiceConfig) *AccessRepositoryBuilder {
	e.config = config
	return e
}

// Build builds an implementation based on the given param
func (e *AccessRepositoryBuilder) Build() (contracts.AccessRepository, error) {
	config := e.config.StoreConfig
	switch config.Kind {
	case "stub":
		glog.Warning("Stub store implementation used. Do not use in production use cases!")
		return &mock.StubAccessRepository{Data: getMockData(), LicensedSeats: map[string]map[domain.SubjectID]bool{}, Licenses: getMockLicenseData()}, nil
	case "spicedb":
		spicedb := &authzed.SpiceDbAccessRepository{}
		token, err := config.ReadToken()
		if err != nil {
			return nil, err
		}
		err = spicedb.NewConnection(config.Endpoint, token, true, config.UseTLS)
		return spicedb, err
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
