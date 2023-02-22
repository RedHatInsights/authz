package engine

import "authz/seatlicensing/domain/model"

// StubAuthzStore -
type StubAuthzStore struct{}

// CheckAccess -
func (s StubAuthzStore) CheckAccess(principal model.Principal, operation string, resource model.Resource) (bool, error) {
	return true, nil
}
