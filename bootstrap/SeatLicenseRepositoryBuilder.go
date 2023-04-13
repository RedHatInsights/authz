package bootstrap

import (
	"authz/domain/contracts"
	"authz/infrastructure/repository/authzed"
)

// SeatLicenseRepositoryBuilder constructs SeatLicenseRepositories based on the provided configuration
type SeatLicenseRepositoryBuilder struct {
	stub  contracts.SeatLicenseRepository
	store string

	endpoint  string
	authToken string
	useTLS    bool
}

// NewSeatLicenseRepositoryBuilder constructs a new SeatLicenseRepositoryBuilder
func NewSeatLicenseRepositoryBuilder() *SeatLicenseRepositoryBuilder {
	return &SeatLicenseRepositoryBuilder{}
}

// WithConnectionInfo configures connection information to be used with applicable implementations
func (b *SeatLicenseRepositoryBuilder) WithConnectionInfo(endpoint string, authToken string, useTLS bool) *SeatLicenseRepositoryBuilder {
	b.endpoint = endpoint
	b.authToken = authToken
	b.useTLS = useTLS

	return b
}

// WithStub provides a stub implementation. This enables a different object to be reused as a stub implementation of SeatLicenseRepository, ex: if the same object implements both seat licensing and access checks.
func (b *SeatLicenseRepositoryBuilder) WithStub(stub contracts.SeatLicenseRepository) *SeatLicenseRepositoryBuilder {
	b.stub = stub
	return b
}

// WithStore specifies the application back-end (ex: stub or spicedb)
func (b *SeatLicenseRepositoryBuilder) WithStore(store string) *SeatLicenseRepositoryBuilder {
	b.store = store
	return b
}

// Build constructs the repository
func (b *SeatLicenseRepositoryBuilder) Build() contracts.SeatLicenseRepository {
	switch b.store {
	case "spicedb":
		spicedb := authzed.SpiceDbAccessRepository{}
		spicedb.NewConnection(b.endpoint, b.authToken, true, b.useTLS)
		return &spicedb
	case "stub":
		return b.stub
	default:
		return b.stub
	}
}
