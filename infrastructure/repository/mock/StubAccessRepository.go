// Package mock implements a stub accessRepository
package mock

import (
	"authz/domain/model"
	vo "authz/domain/valueobjects"
)

// StubAccessRepository represents an in-memory authorization system with a fixed state
type StubAccessRepository struct {
	//The internal authorization state. The keys are subject IDs, and the values are the results. The results are the same per subject regardless of operation and resource.
	Data map[string]bool
}

// NewConnection Stub impl
func (s *StubAccessRepository) NewConnection(endpoint string, token string) {
	// NOT USED IN STUB
}

// CheckAccess returns true if the subject has been specified to have access, otherwise false.
func (s *StubAccessRepository) CheckAccess(principal model.Principal, operation string, resource model.Resource) (vo.AccessDecision, error) {
	if hasAccess, ok := s.Data[principal.ID]; ok {
		return vo.AccessDecision(hasAccess), nil
	}

	return false, nil
}
