package app

import (
	"authz/app/client/authzed"
)

// AuthzService
type AuthzService interface {
	CheckPermission()
}

type serviceAuthz struct {
	authzed authzed.AuthzedClient
}

func (s serviceAuthz) CheckPermission() {
	//TODO implement me
	//s.authzed.CheckPermission()
	panic("implement me")
}

var _ AuthzService = &serviceAuthz{}

// NewAuthzService
func NewAuthzService(endpoint string, token string) *serviceAuthz {
	az := authzed.NewAuthzedConnection(endpoint, token)
	return &serviceAuthz{
		authzed: az,
	}
}
