// Package mock implements a stub accessRepository
package mock

import (
	"authz/domain/model"
	vo "authz/domain/valueobjects"
)

// StubAccessRepository represents an in-memory authorization system with a fixed state
type StubAccessRepository struct {
	//The internal authorization state. The keys are subject IDs, and the values are the results. The results are the same per subject regardless of operation and resource.
	Data          map[vo.SubjectID]bool
	LicensedSeats map[string]map[vo.SubjectID]bool
	Licenses      map[string]model.License
}

// NewConnection Stub impl
func (s *StubAccessRepository) NewConnection(_ string, _ string, _ bool) {
	// NOT USED IN STUB
}

// CheckAccess returns true if the subject has been specified to have access, otherwise false.
func (s *StubAccessRepository) CheckAccess(subjectID vo.SubjectID, operation string, resource model.Resource) (vo.AccessDecision, error) {
	if authz, ok := s.Data[subjectID]; ok {
		if authz && operation == "use" {
			return vo.AccessDecision(s.LicensedSeats[resource.ID][subjectID]), nil //Authorized, so return license status
		}
		return vo.AccessDecision(authz), nil //No licensing required, passthrough authz status
	}

	return vo.AccessDecision(false), nil //Unknown principal, implicitly not authorized
}

// GetLicense retrieves the stored license for the given organization and service, if any.
func (s *StubAccessRepository) GetLicense(_ string, serviceID string) (*model.License, error) {
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

// GetAssigned retrieves the IDs of the subjects assigned seats in the current license
func (s *StubAccessRepository) GetAssigned(_ string, serviceID string) ([]vo.SubjectID, error) {
	subjects := make([]vo.SubjectID, 0)
	if assignments, ok := s.LicensedSeats[serviceID]; ok {
		for id, assigned := range assignments {
			if assigned {
				subjects = append(subjects, id)
			}
		}
	}

	return subjects, nil
}

// AssignSeat assigns the given principal a seat for the given service
func (s *StubAccessRepository) AssignSeat(subjectID vo.SubjectID, _ string, svc model.Service) error {
	if lics, ok := s.LicensedSeats[svc.ID]; ok {
		lics[subjectID] = true
	} else {
		s.LicensedSeats[svc.ID] = map[vo.SubjectID]bool{subjectID: true}
	}
	return nil
}

// UnAssignSeat removes the seat assignment for the given principal for the given service
func (s *StubAccessRepository) UnAssignSeat(subjectID vo.SubjectID, _ string, svc model.Service) error {
	if lics, ok := s.LicensedSeats[svc.ID]; ok {
		lics[subjectID] = false
	}

	return nil
}
