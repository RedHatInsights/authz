package engine

import "authz/app"

type SpiceDbAuthzEngine struct{}

func (s SpiceDbAuthzEngine) CheckAccess(principal app.Principal, operation string, resource app.Resource) (bool, error) {
	//TODO: call spiceDB check using their api client
	return true, nil
}
