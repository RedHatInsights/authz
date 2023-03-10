// Package contracts contains the contracts that can be used in the domain
package contracts

import (
	"authz/domain/model"
)

// AuthzEngine - the contract for the engine
type AuthzEngine interface {
	CheckAccess(principal model.Principal, operation string, resource model.Resource) (bool, error)
	NewConnection(endpoint string, token string)
}
