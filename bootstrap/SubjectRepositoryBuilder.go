package bootstrap

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"authz/infrastructure/repository/userservice"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
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
	url, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	caCert, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	cert, err := tls.LoadX509KeyPair(config.UserServiceClientCertFile, config.UserServiceClientKeyFile)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCert,
				Certificates: []tls.Certificate{cert},
			},
		},
	}

	return userservice.NewUserServiceSubjectRepository(*url, client), nil
}
