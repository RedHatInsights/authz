package bootstrap

import (
	"authz/domain/contracts"
	"authz/domain/model"
	"authz/infrastructure/repository/mock"
)

type PrincipalRepositoryBuilder struct {
	store string
}

func NewPrincipalRepositoryBuilder() *PrincipalRepositoryBuilder {
	return &PrincipalRepositoryBuilder{}
}

func (b *PrincipalRepositoryBuilder) WithStore(store string) *PrincipalRepositoryBuilder {
	b.store = store
	return b
}

func (b *PrincipalRepositoryBuilder) Build() contracts.PrincipalRepository {
	switch b.store {
	case "stub":
		return &mock.StubPrincipalRepository{Principals: getMockPrincipalData()}
	default:
		return &mock.StubPrincipalRepository{Principals: getMockPrincipalData()}
	}
}

func getMockPrincipalData() map[string]model.Principal {
	return map[string]model.Principal{
		"token": model.NewPrincipal("token", "aspian"),
		"alice": model.NewPrincipal("alice", "aspian"),
		"bob":   model.NewPrincipal("bob", "aspian"),
		"chuck": model.NewPrincipal("chuck", "aspian"),
	}
}
