package impl

import (
	"authz/app"
	"authz/app/client/authzed"
	"fmt"
)

// SpiceDBAuthzStore represents an spicedb authorization
type SpiceDBAuthzStore struct {
	Authzed authzed.Client
}

// CheckAccess returns false TODO Impl calling authz client.
func (s SpiceDBAuthzStore) CheckAccess(principal app.Principal, operation string, resource app.Resource) (bool, error) {
	//s.Authzed.CheckPermission()
	return false, nil
}

// ReadSchema - Implements Readschema
func (s SpiceDBAuthzStore) ReadSchema() {
	resp, err := s.Authzed.ReadSchema()
	fmt.Println("Response: ", resp, "Error:", err)

}
