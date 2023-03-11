// Package mock implements a stub authzengine
package mock

import (
	"authz/domain/model"
)

// StubAuthzEngine represents an in-memory authorization system with a fixed state
type StubAuthzEngine struct {
	//The internal authorization state. The keys are subject IDs, and the values are the results. The results are the same per subject regardless of operation and resource.
	Data map[string]bool
}

// NewConnection Stub impl
func (s *StubAuthzEngine) NewConnection(endpoint string, token string) {
	// NOT USED IN STUBSTORE
}

// CheckAccess returns true if the subject has been specified to have access, otherwise false.
func (s *StubAuthzEngine) CheckAccess(principal model.Principal, operation string, resource model.Resource) (bool, error) {
	if authz, ok := s.Data[principal.ID]; ok {
		return authz, nil
	}

	return false, nil
}
