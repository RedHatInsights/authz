package bootstrap

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"authz/infrastructure/repository/userservice"
	"crypto/x509"
)

// SubjectRepositoryBuilder is a builder for SubjectRepository
type SubjectRepositoryBuilder struct {
	config *serviceconfig.ServiceConfig
}

// NewSubjectRepositoryBuilder returns a new NewSubjectRepositoryBuilder
func NewSubjectRepositoryBuilder() *SubjectRepositoryBuilder {
	return &SubjectRepositoryBuilder{}
}

// WithConfig allows SubjectRepository to be built with a given configuration
func (s *SubjectRepositoryBuilder) WithConfig(config *serviceconfig.ServiceConfig) *SubjectRepositoryBuilder {
	s.config = config
	return s
}

// Build uses the builder to create the SubjectRepository
func (s *SubjectRepositoryBuilder) Build() (contracts.SubjectRepository, error) {
	config := s.config.UserServiceConfig

	return createUserServiceSubjectRepository(config)
}

func createUserServiceSubjectRepository(config serviceconfig.UserServiceConfig) (contracts.SubjectRepository, error) {
	caCerts, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	return userservice.NewUserServiceSubjectRepositoryFromConfig(config, caCerts)
}
