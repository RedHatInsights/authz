package bootstrap

import (
	"authz/domain/contracts"
	"authz/infrastructure/repository/mock"
)

// PrincipalRepositoryBuilder constructs a PrincipalRepository based on the given configuration
type PrincipalRepositoryBuilder struct {
	store string
}

// NewPrincipalRepositoryBuilder constructs a new PrincipalRepositoryBuilder
func NewPrincipalRepositoryBuilder() *PrincipalRepositoryBuilder {
	return &PrincipalRepositoryBuilder{}
}

// WithStore specifies the type of backend in use by the application (ex: spicedb or stub)
func (b *PrincipalRepositoryBuilder) WithStore(store string) *PrincipalRepositoryBuilder {
	b.store = store
	return b
}

// Build constructs the repository
func (b *PrincipalRepositoryBuilder) Build() contracts.PrincipalRepository {
	switch b.store {
	case "stub":
		return &mock.StubPrincipalRepository{Principals: mock.GetMockPrincipalData(), DefaultOrg: "o1"}
	default:
		return &mock.StubPrincipalRepository{Principals: mock.GetMockPrincipalData(), DefaultOrg: "o1"}
	}
}
