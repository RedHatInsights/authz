package impl

import "authz/app"

type StubAuthzStore struct {
	Data map[string]bool
}

func (s StubAuthzStore) CheckAccess(principal app.Principal, operation string, resource app.Resource) (bool, error) {
	if authz, ok := s.Data[principal.Id]; ok {
		return authz, nil
	} else {
		return false, nil
	}
}
