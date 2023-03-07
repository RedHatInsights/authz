package authzed

import (
	"context"
	"log"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
)

// AuthzedClient - Authzed client interface
type AuthzedClient interface {
	CheckPermission(checkReq *v1.CheckPermissionRequest) (*v1.CheckPermissionResponse, error)
}

var _ AuthzedClient = &authzedclient{}

type authzedclient struct {
	authzed *authzed.Client
	ctx     context.Context
}

func (a authzedclient) CheckPermission(checkReq *v1.CheckPermissionRequest) (*v1.CheckPermissionResponse, error) {
	return a.authzed.CheckPermission(a.ctx, checkReq)
}

// NewAuthzedConnection - creates and returns a new AuthZ client
func NewAuthzedConnection(endpoint string, token string) *authzedclient {
	client, err := authzed.NewClient(endpoint, grpcutil.WithBearerToken(token))
	if err != nil {
		log.Fatalf("unable to initialize client: %s", err)
	}
	return &authzedclient{
		authzed: client,
		ctx:     context.Background(),
	}
}
