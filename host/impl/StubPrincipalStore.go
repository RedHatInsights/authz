package impl

import (
	"authz/app"
	"errors"
)

type StubPrincipalStore struct {
	Principals map[string]app.Principal
}

func (s StubPrincipalStore) GetByID(id string) (app.Principal, error) {
	if principal, ok := s.Principals[id]; ok {
		return principal, nil
	} else {
		return principal, errors.New("not found") //Nil instead of error?
	}
}

func (s StubPrincipalStore) GetByIDs(ids []string) ([]app.Principal, error) {
	principals := make([]app.Principal, 0, len(ids))

	for i, id := range ids {
		var err error
		if principals[i], err = s.GetByID(id); err != nil {
			return nil, err
		}
	}

	return principals, nil
}

func (s StubPrincipalStore) GetByToken(token string) (app.Principal, error) {
	return s.GetByID(token)
}
