package mock

import (
	"authz/domain/model"
)

// StubPrincipalRepository represents an in-memory store of principal data
type StubPrincipalRepository struct {
	Principals map[string]model.Principal
}

// GetByID retrieves a principal for the given ID. If no ID is provided (ex: empty string), it returns an anonymous principal. If any error occurs, it's returned.
func (s *StubPrincipalRepository) GetByID(id string) (model.Principal, error) {
	if id == "" {
		return model.NewAnonymousPrincipal(), nil
	}

	principal, ok := s.Principals[id]
	if ok {
		return principal, nil
	}

	return s.createAndAddMissingPrincipal(id)
}

// GetByIDs is a bulk version of GetByID to allow the underlying implementation to optimize access to sets of principals and should otherwise have the same behavior.
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

func (s *StubPrincipalRepository) createAndAddMissingPrincipal(id string) (model.Principal, error) {
	p := model.Principal{ID: id}

	s.Principals[id] = p

	return p, nil
}
