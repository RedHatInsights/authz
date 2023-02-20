package dependencies

import (
	"authz/app"
)

type AuthzStore interface {
	CheckAccess(principal app.Principal, operation string, resource app.Resource) (bool, error)
}
