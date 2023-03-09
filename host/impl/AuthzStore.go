package impl

import (
	"authz/app"
	"authz/app/client/authzed"
)

// AuthzStore represents an spicedb authorization
type AuthzStore struct {
	Authzed authzed.Client
}

// CheckAccess returns false TODO Impl calling authz client.
func (s AuthzStore) CheckAccess(principal app.Principal, operation string, resource app.Resource) (bool, error) {
	//s.Authzed.CheckPermission()
	return false, nil
}

func (s AuthzStore) ReadSchema() {
	s.Authzed.ReadSchema()
	return
}
