package mock

import (
	"authz/domain"
	"fmt"
)

// StubPrincipalRepository represents an in-memory store of principal data
type StubPrincipalRepository struct {
	DefaultOrg string
	Principals map[domain.SubjectID]domain.Principal
}

// GetByID retrieves a principal for the given ID. If no ID is provided (ex: empty string), it returns an anonymous principal. If any error occurs, it's returned.
func (s *StubPrincipalRepository) GetByID(id domain.SubjectID) (domain.Principal, error) {
	if id == "" {
		return domain.NewAnonymousPrincipal(), nil
	}

	principal, ok := s.Principals[id]
	if ok {
		return principal, nil
	}

	return s.createAnonPrincipal(id)
}

// GetByIDs is a bulk version of GetByID to allow the underlying implementation to optimize access to sets of principals and should otherwise have the same behavior.
func (s *StubPrincipalRepository) GetByIDs(ids []domain.SubjectID) ([]domain.Principal, error) {
	principals := make([]domain.Principal, len(ids))

	for i, id := range ids {
		var err error
		if principals[i], err = s.GetByID(id); err != nil {
			return nil, err
		}
	}

	return principals, nil
}

// GetByOrgID retrieves all members of the given organization
func (s *StubPrincipalRepository) GetByOrgID(orgID string) ([]domain.SubjectID, error) {
	ids := make([]domain.SubjectID, 0)
	for _, p := range s.Principals {
		if p.OrgID == orgID {
			ids = append(ids, p.ID)
		}
	}
	return ids, nil
}

func (s *StubPrincipalRepository) createAnonPrincipal(id domain.SubjectID) (domain.Principal, error) {
	p := domain.Principal{
		ID:          "anon",
		DisplayName: fmt.Sprintf("User %s", id),
		OrgID:       s.DefaultOrg,
	}

	s.Principals[id] = p

	return p, nil
}
