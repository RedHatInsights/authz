// Package services contains domain services. Only usage of other domain packages allowed.
package services

import (
	"authz/domain"
	"authz/domain/contracts"
)

// AccessService is a domain service for abstract access management (ex: querying whether access has been granted.)
type AccessService struct {
	accessRepository contracts.AccessRepository
}

// NewAccessService constructs a new instance of the Access domain service
func NewAccessService(accessRepository contracts.AccessRepository) AccessService {
	return AccessService{accessRepository}
}

// Check processes a CheckEvent and returns true or false if successful, otherwise error
func (a AccessService) Check(req domain.CheckEvent) (domain.AccessDecision, error) {
	if !req.Requestor.HasIdentity() {
		return false, domain.ErrNotAuthenticated
	}

	accessResult, err := true, error(nil) //a.accessRepository.CheckAccess(req.Requestor, "call", model.Resource{Type: "endpoint", ID: "checkaccess"}) //TODO: implement actual meta-authz
	if err != nil {
		return false, err
	}

	if !accessResult {
		return false, domain.ErrNotAuthorized
	}

	return a.accessRepository.CheckAccess(req.SubjectID, req.Operation, req.Resource)
}
