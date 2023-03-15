// Package handler contains all handlers untangled from the server implementation.
package application

import (
	"authz/domain/contracts"
	"authz/domain/model"
	"authz/domain/services"
	vo "authz/domain/valueobjects"
	"context"
)

// Empty interface tp use for Handlers. no idea if this is idiomatic.
type Handler interface{}

// AccessAppService the handler for permission related endpoints.
type AccessAppService struct {
	accessRepo *contracts.AccessRepository
	ctx        context.Context
}

// CheckRequest is an actual request to check for permissions.
type CheckRequest struct {
	Requestor    model.Principal
	Subject      string
	ResourceType string
	ResourceID   string
	Operation    string
}

// NewPermissionHandler returns a new instance of the permissionhandler.
func (p *AccessAppService) NewPermissionHandler(accessRepo *contracts.AccessRepository) *AccessAppService {
	return &AccessAppService{
		accessRepo: accessRepo,
		ctx:        context.Background(),
	}
}

// Check calls the domainservice using a CheckEvent and can be used with every server impl if wanted.
func (p *AccessAppService) Check(req CheckRequest) (vo.AccessDecision, error) {
	event := model.CheckEvent{
		Request: model.Request{
			Requestor: req.Requestor,
		},
		Subject:   model.Principal{ID: req.Subject},
		Operation: req.Operation,
		Resource:  model.Resource{Type: req.ResourceType, ID: req.ResourceID},
	}

	checkResult := services.NewAccessService(*p.accessRepo)

	result, err := checkResult.Check(event)
	return result, err
}
