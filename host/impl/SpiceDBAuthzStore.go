package impl

import (
	"authz/app"
	"authz/app/client/authzed"
)

// AuthzStore represents an spicedb authorization
type SpiceDBAuthzStore struct {
	Authzed authzed.Client
}

// CheckAccess returns false TODO Impl calling authz client.
func (s SpiceDBAuthzStore) CheckAccess(principal app.Principal, operation string, resource app.Resource) (bool, error) {
	//s.Authzed.CheckPermission()
	return false, nil
}

func (s SpiceDBAuthzStore) ReadSchema() {
	s.Authzed.ReadSchema()
	return
}
