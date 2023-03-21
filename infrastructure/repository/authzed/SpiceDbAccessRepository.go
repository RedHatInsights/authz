// Package authzed contains the technical implementations for the accessRepo from authzed spicedb
package authzed

import (
	"authz/domain/model"
	vo "authz/domain/valueobjects"
	"context"
	"log"

	"github.com/golang/glog"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// SpiceDbAccessRepository -
type SpiceDbAccessRepository struct{}

// authzedClient - Authz client struct
type authzedClient struct {
	client *authzed.Client
	ctx    context.Context
}

var authzedConn *authzedClient

// CheckAccess -
func (s *SpiceDbAccessRepository) CheckAccess(principal model.Principal, operation string, resource model.Resource) (vo.AccessDecision, error) {
	s2, o2 := createSubjectObjectTuple("user", principal.ID, resource.Type, resource.ID)

	//TODO remove me, just for checking it is used (will return an err)
	_, err := authzedConn.client.ReadSchema(authzedConn.ctx, &v1.ReadSchemaRequest{})
	if err != nil {
		glog.Infof("Could not reach spiceDB! Error: %v", err)
	}

	//TODO: Implement.
	r, err := authzedConn.client.CheckPermission(authzedConn.ctx, &v1.CheckPermissionRequest{
		Resource:   o2,
		Permission: operation,
		Subject:    s2,
	})

	if err != nil {
		return false, err
	}

	if r.Permissionship != v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
		return false, nil
	}

	return true, nil
}

// NewConnection creates a new connection to an underlying SpiceDB store and saves it to the package variable conn
func (s *SpiceDbAccessRepository) NewConnection(spiceDbEndpoint string, token string, isBlocking bool) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpcutil.WithInsecureBearerToken(token),
	}

	if isBlocking {
		opts = append(opts, grpc.WithBlock())
	}

	client, err := authzed.NewClient(
		spiceDbEndpoint,
		opts...,
	)

	if err != nil {
		log.Fatalf("unable to initialize client: %s", err)
	}

	authzedConn = &authzedClient{
		client: client,
		ctx:    context.Background(),
	}
}

func createSubjectObjectTuple(subjectType string, subjectValue string, objectType string, objectValue string) (*v1.SubjectReference, *v1.ObjectReference) {
	subject := &v1.SubjectReference{Object: &v1.ObjectReference{
		ObjectType: subjectType,
		ObjectId:   subjectValue,
	}}

	t1 := &v1.ObjectReference{
		ObjectType: objectType,
		ObjectId:   objectValue,
	}
	return subject, t1
}
