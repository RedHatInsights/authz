package bootstrap

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"authz/infrastructure/repository/authzed"
)

// SeatLicenseRepositoryBuilder constructs SeatLicenseRepositories based on the provided configuration
type SeatLicenseRepositoryBuilder struct {
	config *serviceconfig.ServiceConfig
}

// NewSeatLicenseRepositoryBuilder constructs a new SeatLicenseRepositoryBuilder
func NewSeatLicenseRepositoryBuilder() *SeatLicenseRepositoryBuilder {
	return &SeatLicenseRepositoryBuilder{}
}

// WithConfig supplies a ServiceConfig struct to be used as-needed for building objects
func (b *SeatLicenseRepositoryBuilder) WithConfig(config *serviceconfig.ServiceConfig) *SeatLicenseRepositoryBuilder {
	b.config = config
	return b
}

// Build constructs the repository
func (b *SeatLicenseRepositoryBuilder) Build() (contracts.SeatLicenseRepository, error) {
	config := b.config.StoreConfig
	switch config.Kind {
	case "spicedb":
		return createSeatLicenseRepository(config)
	default:
		return createSeatLicenseRepository(config)
	}
}

func createSeatLicenseRepository(config serviceconfig.StoreConfig) (contracts.SeatLicenseRepository, error) {
	spicedb := authzed.SpiceDbAccessRepository{}
	token, err := config.ReadToken()
	if err != nil {
		return nil, err
	}
	err = spicedb.NewConnection(config.Endpoint, token, true, config.UseTLS)
	return &spicedb, err
}
