package app

import (
	"authz/app/client/authzed"
)

// AuthzService
type AuthzService interface {
	CheckPermission()
}

type ServiceAuthz struct {
	authzed authzed.Client
}

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
