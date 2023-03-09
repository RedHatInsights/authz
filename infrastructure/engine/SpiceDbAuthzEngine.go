package engine

import (
	"authz/domain/model"
)

// SpiceDbAuthzEngine -
type SpiceDbAuthzEngine struct{}

// CheckAccess -
func (s SpiceDbAuthzEngine) CheckAccess(principal model.Principal, operation string, resource model.Resource) (bool, error) {
	//TODO: call spiceDB check using their api client
	return true, nil
}
