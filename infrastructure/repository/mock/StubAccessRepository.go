// Package mock implements a stub accessRepository
package mock

import (
	"authz/domain"
)

// StubAccessRepository represents an in-memory authorization system with a fixed state
type StubAccessRepository struct {
	//The internal authorization state. The keys are subject IDs, and the values are the results. The results are the same per subject regardless of operation and resource.
	Data          map[domain.SubjectID]bool
	LicensedSeats map[string]map[domain.SubjectID]bool
	Licenses      map[string]domain.License
}

// NewConnection Stub impl
func (s *StubAccessRepository) NewConnection(_ string, _ string, _ bool, _ bool) {
	// NOT USED IN STUB
}

// CheckAccess returns true if the subject has been specified to have access, otherwise false.
func (s *StubAccessRepository) CheckAccess(subjectID domain.SubjectID, operation string, resource domain.Resource) (domain.AccessDecision, error) {
	if authz, ok := s.Data[subjectID]; ok {
		if authz && operation == "use" {
			return domain.AccessDecision(s.LicensedSeats[resource.ID][subjectID]), nil //Authorized, so return license status
		}
		return domain.AccessDecision(authz), nil //No licensing required, passthrough authz status
	}

	return domain.AccessDecision(false), nil //Unknown principal, implicitly not authorized
}

// GetLicense retrieves the stored license for the given organization and service, if any.
func (s *StubAccessRepository) GetLicense(_ string, serviceID string) (*domain.License, error) {
	lic := s.Licenses[serviceID]
	inuse := 0

	if assignments, ok := s.LicensedSeats[serviceID]; ok {
		for _, v := range assignments {
			if v {
				inuse++
			}
		}
	}

	lic.InUse = inuse
	return &lic, nil
}

// GetAssignable retrieves the IDs of the subjects who are assignable to seats in the current license
func (s *StubAccessRepository) GetAssignable(_ string, serviceID string) ([]domain.SubjectID, error) {
	// TODO: GetAssignable(orgID string, serviceID string) ([]domain.SubjectID, error) to maintain the contract

	return nil, nil
}

// GetAssigned retrieves the IDs of the subjects assigned seats in the current license
func (s *StubAccessRepository) GetAssigned(_ string, serviceID string) ([]domain.SubjectID, error) {
	subjects := make([]domain.SubjectID, 0)
	if assignments, ok := s.LicensedSeats[serviceID]; ok {
		for id, assigned := range assignments {
			if assigned {
				subjects = append(subjects, id)
			}
		}
	}

	return subjects, nil
}

// ModifySeats atomically persists changes to seat assignments for a license
func (s *StubAccessRepository) ModifySeats(_ []domain.SubjectID, _ []domain.SubjectID, _ *domain.License, _ string, _ domain.Service) error {
	panic("Not implemented")
}
