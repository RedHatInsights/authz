package engine

import (
	model2 "authz/domain/model"
)

// StubAuthzStore -
type StubAuthzStore struct{}

// CheckAccess -
func (s StubAuthzStore) CheckAccess(principal model2.Principal, operation string, resource model2.Resource) (bool, error) {
	return true, nil
}
