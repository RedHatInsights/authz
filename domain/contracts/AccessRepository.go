// Package contracts inside the domain package contains the contracts that can be used in the domain
// without being coupled to a technical implementation
package contracts

import (
	"authz/domain/model"
)

// AccessRepository - the contract for the access repository
type AccessRepository interface {
	CheckAccess(principal model.Principal, operation string, resource model.Resource) (bool, error)
	NewConnection(endpoint string, token string)
}
