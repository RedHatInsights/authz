package bootstrap

import "authz/domain/contracts"

type SeatLicenseRepositoryBuilder struct {
	stub  contracts.SeatLicenseRepository
	store string
}

func NewSeatLicenseRepositoryBuilder() *SeatLicenseRepositoryBuilder {
	return &SeatLicenseRepositoryBuilder{}
}

func (b *SeatLicenseRepositoryBuilder) WithStub(stub contracts.SeatLicenseRepository) *SeatLicenseRepositoryBuilder {
	b.stub = stub
	return b
}

func (b *SeatLicenseRepositoryBuilder) WithStore(store string) *SeatLicenseRepositoryBuilder {
	b.store = store
	return b
}

func (b *SeatLicenseRepositoryBuilder) Build() contracts.SeatLicenseRepository {
	switch b.store {
	case "stub":
		return b.stub
	default:
		return b.stub
	}
}
