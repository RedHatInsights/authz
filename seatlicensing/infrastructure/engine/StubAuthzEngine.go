package engine

import "authz/seatlicensing/domain/model"

type StubAuthzStore struct{}

func (s StubAuthzStore) CheckAccess(principal model.Principal, operation string, resource model.Resource) (bool, error) {
	return true, nil
}
