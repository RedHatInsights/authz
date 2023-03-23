// Package mock implements a stub accessRepository
package mock

import (
	"authz/domain/model"
	vo "authz/domain/valueobjects"
)

// StubAccessRepository represents an in-memory authorization system with a fixed state
type StubAccessRepository struct {
	//The internal authorization state. The keys are subject IDs, and the values are the results. The results are the same per subject regardless of operation and resource.
	Data          map[string]bool
	LicensedSeats map[string]map[string]bool
}

// NewConnection Stub impl
func (s *StubAccessRepository) NewConnection(_ string, _ string, _ bool) {
	// NOT USED IN STUB
}

// CheckAccess returns true if the subject has been specified to have access, otherwise false.
func (s *StubAccessRepository) CheckAccess(principal model.Principal, operation string, resource model.Resource) (vo.AccessDecision, error) {
	if authz, ok := s.Data[principal.ID]; ok {
		if authz && operation == "use" {
			return vo.AccessDecision(s.LicensedSeats[principal.ID][resource.ID]), nil //Authorized, so return license status
		}
		return vo.AccessDecision(authz), nil //No licensing required, passthrough authz status
	}

	return vo.AccessDecision(false), nil //Unknown principal, implicitly not authorized
}

// AssignSeat assigns the given principal a seat for the given service
func (s *StubAccessRepository) AssignSeat(principal model.Principal, svc model.Service) error {
	if lics, ok := s.LicensedSeats[principal.ID]; ok {
		lics[svc.ID] = true
	} else {
		s.LicensedSeats[principal.ID] = map[string]bool{svc.ID: true}
	}
	return nil
}

// UnAssignSeat removes the seat assignment for the given principal for the given service
func (s *StubAccessRepository) UnAssignSeat(principal model.Principal, svc model.Service) error {
	if lics, ok := s.LicensedSeats[principal.ID]; ok {
		lics[svc.ID] = false
	}

	return nil
}
