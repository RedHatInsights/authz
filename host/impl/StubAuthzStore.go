package impl

import (
	"authz/app"
)

// StubAuthzStore represents an in-memory authorization system with a fixed state
type StubAuthzStore struct {
	//The internal authorization state. The keys are subject IDs, and the values are the results. The results are the same per subject regardless of operation and resource.
	AuthzdUsers   map[string]bool
	LicensedSeats map[string]map[string]bool
}

// CheckAccess returns true if the subject has been specified to have access, otherwise false.
func (s *StubAuthzStore) CheckAccess(principal app.Principal, operation string, resource app.Resource) (bool, error) {
	if authz, ok := s.AuthzdUsers[principal.ID]; ok {
		if authz && operation == "use" {
			return s.LicensedSeats[principal.ID][resource.ID], nil //Authorized, so return license status
		}
		return authz, nil //No licensing required, passthrough authz status
	}

	return false, nil //Unknown principal, implicitly not authorized
}

func (s *StubAuthzStore) AssignSeat(principal app.Principal, svc app.Service) error {
	if lics, ok := s.LicensedSeats[principal.ID]; ok {
		lics[svc.Id] = true
	} else {
		s.LicensedSeats[principal.ID] = map[string]bool{svc.Id: true}
	}
	return nil
}

func (s *StubAuthzStore) UnAssignSeat(principal app.Principal, svc app.Service) error {
	if lics, ok := s.LicensedSeats[principal.ID]; ok {
		lics[svc.Id] = false
	}

	return nil
}
