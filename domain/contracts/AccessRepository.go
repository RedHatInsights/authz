// Package contracts inside the domain package contains the contracts that can be used in the domain
// without being coupled to a technical implementation
package contracts

import (
	"authz/domain/model"
	vo "authz/domain/valueobjects"
)

// AccessRepository - the contract for the access repository
type AccessRepository interface {
	CheckAccess(principal model.Principal, operation string, resource model.Resource) (vo.AccessDecision, error)
	NewConnection(endpoint string, token string, isBlocking bool) //TODO: Remove from interface.don't think it is needed here.
}
