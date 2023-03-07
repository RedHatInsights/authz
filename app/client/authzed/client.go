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
}

var _ Client = &Authzedclient{}

// Authzedclient - Authz client struct
type Authzedclient struct {
	authzed *authzed.Client
	ctx     context.Context
}

func (a Authzedclient) CheckPermission(checkReq *v1.CheckPermissionRequest) (*v1.CheckPermissionResponse, error) {
	return a.authzed.CheckPermission(a.ctx, checkReq)
}

// NewAuthzedConnection - creates and returns a new AuthZ client
func NewAuthzedConnection(endpoint string, token string) *Authzedclient {
	client, err := authzed.NewClient(endpoint, grpcutil.WithBearerToken(token))
	if err != nil {
		log.Fatalf("unable to initialize client: %s", err)
	}
	return &Authzedclient{
		authzed: client,
		ctx:     context.Background(),
	}
}
