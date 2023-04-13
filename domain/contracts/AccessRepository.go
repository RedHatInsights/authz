// Package contracts inside the domain package contains the contracts that can be used in the domain
// without being coupled to a technical implementation
package contracts

import (
	"authz/domain"
)

// AccessRepository - the contract for the access repository
type AccessRepository interface {
	CheckAccess(subjectID domain.SubjectID, operation string, resource domain.Resource) (domain.AccessDecision, error)
	NewConnection(endpoint string, token string, isBlocking, useTLS bool) //TODO: Remove from interface.don't think it is needed here.
}
