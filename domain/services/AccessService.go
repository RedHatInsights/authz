package services

import (
	"authz/domain/contracts"
	"authz/domain/model"
	"authz/infrastructure/server"
)

// AccessService is a domain service for abstract access management (ex: querying whether access has been granted.)
type AccessService struct {
	engine contracts.AuthzEngine
}

// NewAccessService constructs a new instance of the Access domain service
func NewAccessService(engine contracts.AuthzEngine) AccessService {
	return AccessService{engine}
}

// Check processes a CheckRequest and returns true or false if successful, otherwise error
func (a AccessService) Check(req model.CheckRequest) (bool, error) {
	if req.Requestor.IsAnonymous() {
		return false, server.ErrNotAuthenticated
	}

	accessResult, err := a.engine.CheckAccess(req.Requestor, "call", model.Resource{Type: "endpoint", ID: "checkaccess"})
	if err != nil {
		return false, err
	}

	if !accessResult {
		return false, server.ErrNotAuthorized
	}

	return a.engine.CheckAccess(req.Subject, req.Operation, req.Resource)
}
