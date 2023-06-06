package bootstrap

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"authz/infrastructure/repository/authzed"
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
	case "spicedb":
		return createSpiceDbRepository(config)
	default:
		return createSpiceDbRepository(config)
	}
}

func createSpiceDbRepository(config serviceconfig.StoreConfig) (contracts.AccessRepository, error) {
	spicedb := &authzed.SpiceDbAccessRepository{}
	token, err := config.ReadToken()
	if err != nil {
		return nil, err
	}
	err = spicedb.NewConnection(config.Endpoint, token, true, config.UseTLS)
	return spicedb, err
}
