package engine

import (
	model2 "authz/domain/model"
)

// SpiceDbAuthzEngine -
type SpiceDbAuthzEngine struct{}

// CheckAccess -
func (s SpiceDbAuthzEngine) CheckAccess(principal model2.Principal, operation string, resource model2.Resource) (bool, error) {
	//TODO: call spiceDB check using their api client
	return true, nil
}
