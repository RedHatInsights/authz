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

// SubjectType user
const SubjectType = "user"

// SpiceDbAccessRepository -
type SpiceDbAccessRepository struct{}

// authzedClient - Authz client struct
type authzedClient struct {
	client *authzed.Client
	ctx    context.Context
}

var authzedConn *authzedClient

// CheckAccess - verify permission with subject type "user"
func (s *SpiceDbAccessRepository) CheckAccess(subjectID vo.SubjectID, operation string, resource model.Resource) (vo.AccessDecision, error) {
	subject, object := createSubjectObjectTuple(SubjectType, string(subjectID), resource.Type, resource.ID)

	result, err := authzedConn.client.CheckPermission(authzedConn.ctx, &v1.CheckPermissionRequest{
		Resource:   object,
		Permission: operation,
		Subject:    subject,
	})

	if err != nil {
		glog.Errorf("Failed to check permission :%v", err.Error())
		return false, err
	}

	if result.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION {
		return true, nil
	}

	//DENIED BY DEFAULT
	return false, nil
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

	object := &v1.ObjectReference{
		ObjectType: objectType,
		ObjectId:   objectValue,
	}
	return subject, object
}
