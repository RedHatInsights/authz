package authzed

import (
	"context"
	"log"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
)

// Client - Authzed client interface
type Client interface {
	CheckPermission(checkReq *v1.CheckPermissionRequest) (*v1.CheckPermissionResponse, error)
	ReadSchema() (*v1.ReadSchemaResponse, error)
	WriteSchema(schemaToWrite string) (*v1.WriteSchemaResponse, error)
}

var _ Client = &Authzedclient{}

// Authzedclient - Authz client struct
type Authzedclient struct {
	authzed *authzed.Client
	ctx     context.Context
}

// CheckPermission - Checkpermission wrapper
func (a Authzedclient) CheckPermission(checkReq *v1.CheckPermissionRequest) (*v1.CheckPermissionResponse, error) {
	return a.authzed.CheckPermission(a.ctx, checkReq)
}

// ReadSchema - Read Schema wrapper
func (a Authzedclient) ReadSchema() (*v1.ReadSchemaResponse, error) {
	request := &v1.ReadSchemaRequest{}
	return a.authzed.ReadSchema(a.ctx, request)
}

// WriteSchema - Write Schema wrapper
func (a Authzedclient) WriteSchema(schemaToWrite string) (*v1.WriteSchemaResponse, error) {
	request := &v1.WriteSchemaRequest{Schema: schemaToWrite}
	return a.authzed.WriteSchema(a.ctx, request)
}

// NewAuthzedConnection - creates and returns a new AuthZ client
func NewAuthzedConnection(endpoint string, token string) *Authzedclient {
	//TODO: this might not be needed when calling the service from our cluster
	skipCA, _ := grpcutil.WithSystemCerts(grpcutil.SkipVerifyCA)
	client, err := authzed.NewClient(endpoint, grpcutil.WithBearerToken(token), skipCA)
	if err != nil {
		log.Fatalf("unable to initialize client: %s", err)
	}
	return &Authzedclient{
		authzed: client,
		ctx:     context.Background(),
	}
}
