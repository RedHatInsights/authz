package contracts

import "authz/app"

type AuthzEngine interface {
	CheckAccess(principal app.Principal, operation string, resource app.Resource) (bool, error)
}
