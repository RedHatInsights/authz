package controllers

import (
	"authz/app"
	"authz/app/contracts"
	"authz/app/dependencies"
)

// Access is an domain service for abstract access management (ex: querying whether access has been granted.)
type Access struct {
	store dependencies.AuthzStore
}

// NewAccess constructs a new instance of the Access domain service
func NewAccess(store dependencies.AuthzStore) Access {
	return Access{store}
}

// Check processes a CheckRequest and returns true or false if successful, otherwise error
func (a Access) Check(req contracts.CheckRequest) (bool, error) {
	if req.Requestor.IsAnonymous() {
		return false, app.ErrNotAuthenticated
	}

	authzed, err := a.store.CheckAccess(req.Requestor, "call", app.Resource{Type: "endpoint", ID: "checkaccess"})
	if err != nil {
		return false, err
	}

	if !authzed {
		return false, app.ErrNotAuthorized
	}

	return a.store.CheckAccess(req.Subject, req.Operation, req.Resource)
}
