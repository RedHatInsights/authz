package impl

import (
	"authz/app"
	"authz/app/client/authzed"
)

// AuthzStore represents an spicedb authorization
type AuthzStore struct {
	Authzed authzed.Client
}

// CheckPermission - Check permission TODO
// CheckAccess returns false TODO implementation.
func (s AuthzStore) CheckAccess(principal app.Principal, operation string, resource app.Resource) (bool, error) {
	//s.Authzed.CheckPermission()
	return false, nil
}
