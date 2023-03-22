package mock

import (
	"authz/domain/model"
	"errors"
)

type StubPrincipalRepository struct {
	Principals map[string]model.Principal
}

func (s *StubPrincipalRepository) GetByID(id string) (model.Principal, error) {
	if principal, ok := s.Principals[id]; ok {
		return principal, nil
	} else {
		return principal, errors.New("not found") //Nil instead of error?
	}
}

func (s *StubPrincipalRepository) GetByIDs(ids []string) ([]model.Principal, error) {
	principals := make([]model.Principal, len(ids))

	for i, id := range ids {
		var err error
		if principals[i], err = s.GetByID(id); err != nil {
			return nil, err
		}
	}

	return principals, nil
}

func (s *StubPrincipalRepository) GetByToken(token string) (model.Principal, error) { //TODO: this may be more of an application-layer concern
	return s.GetByID(token)
}
