package contracts

import (
	model2 "authz/domain/model"
)

// AuthzEngine - the contract for the engine
type AuthzEngine interface {
	CheckAccess(principal model2.Principal, operation string, resource model2.Resource) (bool, error)
}
