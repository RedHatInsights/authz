// Package mock contains mock implementations for the store.
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

	return s.createAndAddMissingPrincipal(id)
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
func (s *StubPrincipalRepository) GetByOrgID(orgID string) (chan domain.Subject, chan error) {
	subjects := make(chan domain.Subject)
	errors := make(chan error)

	go func() {
		for _, p := range s.Principals {
			if p.OrgID == orgID {
				subjects <- domain.Subject{
					SubjectID: p.ID,
					Enabled:   true,
				}
			}
		}
		close(subjects)
		close(errors)
	}()

	return subjects, errors
}

func (s *StubPrincipalRepository) createAndAddMissingPrincipal(id domain.SubjectID) (domain.Principal, error) {
	p := domain.Principal{
		ID:          id,
		DisplayName: fmt.Sprintf("User %s", id),
		OrgID:       s.DefaultOrg,
	}

	s.Principals[id] = p

	return p, nil
}
