package app

import (
	"authz/app/client/authzed"
)

// AuthzService - AuthZ service interface
type AuthzService interface {
	CheckPermission()
}

// ServiceAuthz - ServiceAuthz struct definition
type ServiceAuthz struct {
	authzed authzed.Client
}

// CheckPermission - Check permission TODO
func (s ServiceAuthz) CheckPermission() {
	//TODO implement me
	//s.authzed.CheckPermission()
	panic("implement me")
}

var _ AuthzService = &ServiceAuthz{}

// NewAuthzService - returns a new AuthZ service
func NewAuthzService(endpoint string, token string) *ServiceAuthz {
	az := authzed.NewAuthzedConnection(endpoint, token)
	return &ServiceAuthz{
		authzed: az,
	}
}
