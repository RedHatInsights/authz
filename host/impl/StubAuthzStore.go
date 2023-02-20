package impl

import "authz/app"

type StubAuthzStore struct {
}

func (s StubAuthzStore) CheckAccess(principal app.Principal, operation string, resource app.Resource) (bool, error) {
	return true, nil
}
