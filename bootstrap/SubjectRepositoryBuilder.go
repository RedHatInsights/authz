package bootstrap

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"authz/infrastructure/repository/userservice"
	"crypto/x509"
)

type SubjectRepositoryBuilder struct {
	config *serviceconfig.ServiceConfig
}

func NewSubjectRepositoryBuilder() *SubjectRepositoryBuilder {
	return &SubjectRepositoryBuilder{}
}

func (s *SubjectRepositoryBuilder) WithConfig(config *serviceconfig.ServiceConfig) *SubjectRepositoryBuilder {
	s.config = config
	return s
}

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
