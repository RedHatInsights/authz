// Package services contains domain services. Only usage of other domain packages allowed.
package services

import (
	"authz/domain/contracts"
	"authz/domain/model"
	vo "authz/domain/valueobjects"
)

// AccessService is a domain service for abstract access management (ex: querying whether access has been granted.)
type AccessService struct {
	accessRepository contracts.AccessRepository
}

// NewAccessService constructs a new instance of the Access domain service
func NewAccessService(accessRepository contracts.AccessRepository) AccessService {
	return AccessService{accessRepository}
}

// Check processes a CheckRequest and returns true or false if successful, otherwise error
func (a AccessService) Check(req model.CheckRequest) (vo.AccessDecision, error) {
	if req.Requestor.IsAnonymous() {
		return false, model.ErrNotAuthenticated
	}

	accessResult, err := a.accessRepository.CheckAccess(req.Requestor, "call", model.Resource{Type: "endpoint", ID: "checkaccess"})
	if err != nil {
		return false, err
	}

	if !accessResult {
		return false, model.ErrNotAuthorized
	}

	return a.accessRepository.CheckAccess(req.Subject, req.Operation, req.Resource)
}
