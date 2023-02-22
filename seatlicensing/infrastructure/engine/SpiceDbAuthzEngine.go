package engine

import "authz/seatlicensing/domain/model"

type SpiceDbAuthzEngine struct{}

func (s SpiceDbAuthzEngine) CheckAccess(principal model.Principal, operation string, resource model.Resource) (bool, error) {
	//TODO: call spiceDB check using their api client
	return true, nil
}
