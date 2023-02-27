package dependencies

import (
	"authz/app"
)

//AuthzStore represents an abstraction for systems that track authorization data and can respond to authorization queries.
type AuthzStore interface {
	CheckAccess(principal app.Principal, operation string, resource app.Resource) (bool, error)
}
