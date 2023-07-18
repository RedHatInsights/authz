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
		ID:        id,
		FirstName: "User",
		LastName:  string(id),
		Username:  fmt.Sprintf("user_%s", id),
		OrgID:     s.DefaultOrg,
	}

	s.Principals[id] = p

	return p, nil
}

// GetMockPrincipalData returns generated fake principal data to cover a variety of cases
func GetMockPrincipalData() map[domain.SubjectID]domain.Principal {
	return map[domain.SubjectID]domain.Principal{
		"token": domain.NewPrincipal("token", "System", "User", "system_user", "smarts"),
		"u1":    domain.NewPrincipal("u1", "O1", "User 1", "user_1", "o1"),
		"u2":    domain.NewPrincipal("u2", "O1", "User 2", "user_2", "o1"),
		"u3":    domain.NewPrincipal("u3", "O1", "User 3", "user_3", "o1"),
		"u4":    domain.NewPrincipal("u4", "O1", "User 4", "user_4", "o1"),
		"u5":    domain.NewPrincipal("u5", "O1", "User 5", "user_5", "o1"),
		"u6":    domain.NewPrincipal("u6", "O1", "User 6", "user_6", "o1"),
		"u7":    domain.NewPrincipal("u7", "O1", "User 7", "user_7", "o1"),
		"u8":    domain.NewPrincipal("u8", "O1", "User 8", "user_8", "o1"),
		"u9":    domain.NewPrincipal("u9", "O1", "User 9", "user_9", "o1"),
		"u10":   domain.NewPrincipal("u10", "O1", "User 10", "user_10", "o1"),
		"u11":   domain.NewPrincipal("u11", "O1", "User 11", "user_11", "o1"),
		"u12":   domain.NewPrincipal("u12", "O1", "User 12", "user_12", "o1"),
		"u13":   domain.NewPrincipal("u13", "O1", "User 13", "user_13", "o1"),
		"u14":   domain.NewPrincipal("u14", "O1", "User 14", "user_14", "o1"),
		"u15":   domain.NewPrincipal("u15", "O1", "User 15", "user_15", "o1"),
		"u16":   domain.NewPrincipal("u16", "O1", "User 16", "user_16", "o1"),
		"u17":   domain.NewPrincipal("u17", "O1", "User 17", "user_17", "o1"),
		"u18":   domain.NewPrincipal("u18", "O1", "User 18", "user_18", "o1"),
		"u19":   domain.NewPrincipal("u19", "O1", "User 19", "user_19", "o1"),
		"u20":   domain.NewPrincipal("u20", "O1", "User 20", "user_20", "o1"),
		"u21":   domain.NewPrincipal("u1", "O2", "User 1", "user_1", "o2"),
		"u22":   domain.NewPrincipal("u2", "O2", "User 2", "user_2", "o2"),
		"u23":   domain.NewPrincipal("u3", "O2", "User 3", "user_3", "o2"),
		"u24":   domain.NewPrincipal("u4", "O2", "User 4", "user_4", "o2"),
		"u25":   domain.NewPrincipal("u5", "O2", "User 5", "user_5", "o2"),
		"u26":   domain.NewPrincipal("u6", "O2", "User 6", "user_6", "o2"),
		"u27":   domain.NewPrincipal("u7", "O2", "User 7", "user_7", "o2"),
		"u28":   domain.NewPrincipal("u8", "O2", "User 8", "user_8", "o2"),
		"u29":   domain.NewPrincipal("u9", "O2", "User 9", "user_9", "o2"),
		"u30":   domain.NewPrincipal("u10", "O2", "User 10", "user_10", "o2"),
		"u31":   domain.NewPrincipal("u11", "O2", "User 11", "user_11", "o2"),
		"u32":   domain.NewPrincipal("u12", "O2", "User 12", "user_12", "o2"),
		"u33":   domain.NewPrincipal("u13", "O2", "User 13", "user_13", "o2"),
		"u34":   domain.NewPrincipal("u14", "O2", "User 14", "user_14", "o2"),
		"u35":   domain.NewPrincipal("u15", "O2", "User 15", "user_15", "o2"),
		"u36":   domain.NewPrincipal("u16", "O2", "User 16", "user_16", "o2"),
		"u37":   domain.NewPrincipal("u17", "O2", "User 17", "user_17", "o2"),
		"u38":   domain.NewPrincipal("u18", "O2", "User 18", "user_18", "o2"),
		"u39":   domain.NewPrincipal("u19", "O2", "User 19", "user_19", "o2"),
		"u40":   domain.NewPrincipal("u20", "O2", "User 20", "user_20", "o2"),
	}
}
