package impl

import "authz/app"

// StubAuthzStore represents an in-memory authorization system with a fixed state
type StubAuthzStore struct {
	//The internal authorization state. The keys are subject IDs, and the values are the results. The results are the same per subject regardless of operation and resource.
	Data map[string]bool
}

// CheckAccess returns true if the subject has been specified to have access, otherwise false.
func (s StubAuthzStore) CheckAccess(principal app.Principal, operation string, resource app.Resource) (bool, error) {
	if authz, ok := s.Data[principal.ID]; ok {
		return authz, nil
	}

	return false, nil
}
