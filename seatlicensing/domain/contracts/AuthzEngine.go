package contracts

import "authz/app"

//TODO: Add these from wills draft.
type AuthzEngine interface {
	CheckAccess(principal app.Principal, operation string, resource app.Resource) (bool, error)
}
