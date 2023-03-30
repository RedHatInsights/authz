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
		return &mock.StubPrincipalRepository{Principals: getMockPrincipalData(), DefaultOrg: "o1"}
	default:
		return &mock.StubPrincipalRepository{Principals: getMockPrincipalData(), DefaultOrg: "o1"}
	}
}

func getMockPrincipalData() map[vo.SubjectID]model.Principal {
	return map[vo.SubjectID]model.Principal{
		"token": model.NewPrincipal("token", "System User", "smarts"),
		"u1":    model.NewPrincipal("u1", "O1 User 1", "o1"),
		"u2":    model.NewPrincipal("u2", "O1 User 2", "o1"),
		"u3":    model.NewPrincipal("u3", "O1 User 3", "o1"),
		"u4":    model.NewPrincipal("u4", "O1 User 4", "o1"),
		"u5":    model.NewPrincipal("u5", "O1 User 5", "o1"),
		"u6":    model.NewPrincipal("u6", "O1 User 6", "o1"),
		"u7":    model.NewPrincipal("u7", "O1 User 7", "o1"),
		"u8":    model.NewPrincipal("u8", "O1 User 8", "o1"),
		"u9":    model.NewPrincipal("u9", "O1 User 9", "o1"),
		"u10":   model.NewPrincipal("u10", "O1 User 10", "o1"),
		"u11":   model.NewPrincipal("u11", "O1 User 11", "o1"),
		"u12":   model.NewPrincipal("u12", "O1 User 12", "o1"),
		"u13":   model.NewPrincipal("u13", "O1 User 13", "o1"),
		"u14":   model.NewPrincipal("u14", "O1 User 14", "o1"),
		"u15":   model.NewPrincipal("u15", "O1 User 15", "o1"),
		"u16":   model.NewPrincipal("u16", "O1 User 16", "o1"),
		"u17":   model.NewPrincipal("u17", "O1 User 17", "o1"),
		"u18":   model.NewPrincipal("u18", "O1 User 18", "o1"),
		"u19":   model.NewPrincipal("u19", "O1 User 19", "o1"),
		"u20":   model.NewPrincipal("u20", "O1 User 20", "o1"),

		"u21": model.NewPrincipal("u1", "O2 User 1", "o2"),
		"u22": model.NewPrincipal("u2", "O2 User 2", "o2"),
		"u23": model.NewPrincipal("u3", "O2 User 3", "o2"),
		"u24": model.NewPrincipal("u4", "O2 User 4", "o2"),
		"u25": model.NewPrincipal("u5", "O2 User 5", "o2"),
		"u26": model.NewPrincipal("u6", "O2 User 6", "o2"),
		"u27": model.NewPrincipal("u7", "O2 User 7", "o2"),
		"u28": model.NewPrincipal("u8", "O2 User 8", "o2"),
		"u29": model.NewPrincipal("u9", "O2 User 9", "o2"),
		"u30": model.NewPrincipal("u10", "O2 User 10", "o2"),
		"u31": model.NewPrincipal("u11", "O2 User 11", "o2"),
		"u32": model.NewPrincipal("u12", "O2 User 12", "o2"),
		"u33": model.NewPrincipal("u13", "O2 User 13", "o2"),
		"u34": model.NewPrincipal("u14", "O2 User 14", "o2"),
		"u35": model.NewPrincipal("u15", "O2 User 15", "o2"),
		"u36": model.NewPrincipal("u16", "O2 User 16", "o2"),
		"u37": model.NewPrincipal("u17", "O2 User 17", "o2"),
		"u38": model.NewPrincipal("u18", "O2 User 18", "o2"),
		"u39": model.NewPrincipal("u19", "O2 User 19", "o2"),
		"u40": model.NewPrincipal("u20", "O2 User 20", "o2"),
	}
}
