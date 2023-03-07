package app

import (
	"authz/app/client/authzed"

	"github.com/golang/glog"
)

// AuthzService - AuthZ service interface
type AuthzService interface {
	CheckPermission()
	WriteSchema(req string)
	ReadSchema()
}

// ServiceAuthz - ServiceAuthz struct definition
type ServiceAuthz struct {
	authzed authzed.Client
}

// CheckPermission - Check permission TODO
func (s ServiceAuthz) CheckPermission() {
	glog.Errorf("Method not implemented yet - implement me")
}

// ReadSchema - ReadSchema method
func (s ServiceAuthz) ReadSchema() {
	resp, err := s.authzed.ReadSchema()
	if err != nil {
		glog.Errorf("Error in reading schema: %v", err)
	}
	glog.Infof("Received Schema text: %v", resp.SchemaText)
}

// WriteSchema - WriteSchema method
func (s ServiceAuthz) WriteSchema(writeReq string) {
	resp, err := s.authzed.WriteSchema(writeReq)
	if err != nil {
		glog.Errorf("Error in writing schema: %v", err)
	}
	glog.Infof("Received writeschema resp: %v", resp)
}

var _ AuthzService = &ServiceAuthz{}

// NewAuthzService - returns a new AuthZ service
func NewAuthzService(endpoint string, token string) *ServiceAuthz {
	az := authzed.NewAuthzedConnection(endpoint, token)
	return &ServiceAuthz{
		authzed: az,
	}
}
