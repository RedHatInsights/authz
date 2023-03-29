package bootstrap

import (
	"authz/domain/contracts"
	"authz/domain/model"
	vo "authz/domain/valueobjects"
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
		return &mock.StubPrincipalRepository{Principals: getMockPrincipalData()}
	default:
		return &mock.StubPrincipalRepository{Principals: getMockPrincipalData()}
	}
}

func getMockPrincipalData() map[vo.SubjectID]model.Principal {
	return map[vo.SubjectID]model.Principal{
		"token": model.NewPrincipal("token"),
		"alice": model.NewPrincipal("alice"),
		"bob":   model.NewPrincipal("bob"),
		"chuck": model.NewPrincipal("chuck"),
	}
}
