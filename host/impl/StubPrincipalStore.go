package impl

import (
	"authz/app"
)

type StubPrincipalStore struct {
}

func (s StubPrincipalStore) GetByID(id string) app.Principal {
	return app.NewPrincipal(id, id)
}

func (s StubPrincipalStore) GetByIDs(ids []string) []app.Principal {
	principals := make([]app.Principal, 0, len(ids))

	for i, id := range ids {
		principals[i] = s.GetByID(id)
	}

	return principals
}

func (s StubPrincipalStore) GetByToken(token string) app.Principal {
	return s.GetByID(token)
}
