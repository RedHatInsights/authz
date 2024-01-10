// Package authzed contains the technical implementations for the accessRepo from authzed spicedb
package authzed

import (
	"authz/domain"
	"authz/infrastructure/grpcutil"
	"context"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/golang/glog"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	// OrgType org relation
	OrgType = "org"
	// SubjectType user relation
	SubjectType = "user"
	// LicenseVersionStr - License Version relation
	LicenseVersionStr = "version"
)

// SpiceDbAccessRepository -
type SpiceDbAccessRepository struct {
	client    *authzed.Client
	ctx       context.Context
	CurrToken string
}

// CheckAccess - verify permission with subject type "user"
func (s *SpiceDbAccessRepository) CheckAccess(subjectID domain.SubjectID, operation string, resource domain.Resource) (domain.AccessDecision, error) {
	subject, object := createSubjectObjectTuple(SubjectType, string(subjectID), resource.Type, resource.ID)

	result, err := s.client.CheckPermission(s.ctx, &v1.CheckPermissionRequest{
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

// AddSubject stores a subject associated with an organization. If a subject is already found, it returns an error.
func (s *SpiceDbAccessRepository) AddSubject(orgID string, subject domain.Subject) error {
	relationshipUpdates := make([]*v1.RelationshipUpdate, 0, 2)

	orgResource := &v1.ObjectReference{
		ObjectType: "org",
		ObjectId:   orgID,
	}
	userSubject := &v1.SubjectReference{
		Object: &v1.ObjectReference{
			ObjectType: "user",
			ObjectId:   string(subject.SubjectID),
		},
	}

	relationshipUpdates = append(relationshipUpdates, &v1.RelationshipUpdate{
		Operation: v1.RelationshipUpdate_OPERATION_CREATE,
		Relationship: &v1.Relationship{
			Resource: orgResource,
			Relation: "member",
			Subject:  userSubject,
		},
	})

	if !subject.Enabled { //conditionally add tombstone
		relationshipUpdates = append(relationshipUpdates, &v1.RelationshipUpdate{
			Operation: v1.RelationshipUpdate_OPERATION_CREATE,
			Relationship: &v1.Relationship{
				Resource: orgResource,
				Relation: "disabled",
				Subject:  userSubject,
			},
		})
	}

	_, err := s.client.WriteRelationships(s.ctx, &v1.WriteRelationshipsRequest{
		Updates: relationshipUpdates,
		OptionalPreconditions: []*v1.Precondition{{
			Operation: v1.Precondition_OPERATION_MUST_NOT_MATCH,
			Filter: &v1.RelationshipFilter{
				ResourceType:       orgResource.ObjectType,
				OptionalResourceId: orgResource.ObjectId,
				OptionalRelation:   "member",
				OptionalSubjectFilter: &v1.SubjectFilter{
					SubjectType:       userSubject.Object.ObjectType,
					OptionalSubjectId: userSubject.Object.ObjectId,
				},
			},
		}},
	})

	err = spiceDbErrorToDomainError(err)

	if err == domain.ErrConflict {
		err = domain.ErrSubjectAlreadyExists
	}

	return err
}

// UpsertSubject stores a subject associated with an organization. If a subject is found, it gets updated. If it is not found, it gets created.
func (s *SpiceDbAccessRepository) UpsertSubject(orgID string, subject domain.Subject) error {
	relationshipUpdates := make([]*v1.RelationshipUpdate, 0, 2)

	orgResource := &v1.ObjectReference{
		ObjectType: "org",
		ObjectId:   orgID,
	}
	userSubject := &v1.SubjectReference{
		Object: &v1.ObjectReference{
			ObjectType: "user",
			ObjectId:   string(subject.SubjectID),
		},
	}

	relationshipUpdates = append(relationshipUpdates, &v1.RelationshipUpdate{
		Operation: v1.RelationshipUpdate_OPERATION_TOUCH,
		Relationship: &v1.Relationship{
			Resource: orgResource,
			Relation: "member",
			Subject:  userSubject,
		},
	})

	var tombstoneOperation v1.RelationshipUpdate_Operation

	// Four scenarios:
	if subject.Enabled {
		// Scenario 1: If previously enabled, delete nonexistent tombstone (no change)
		// Scenario 2: if previously disabled, delete tombstone, now enabled
		tombstoneOperation = v1.RelationshipUpdate_OPERATION_DELETE
	} else {
		// Scenario 3: if previously disabled, touch tombstone that's already there (no change)
		// Scenario 4: if previously enabled, add tombstone, now disabled
		tombstoneOperation = v1.RelationshipUpdate_OPERATION_TOUCH
	}

	relationshipUpdates = append(relationshipUpdates, &v1.RelationshipUpdate{
		Operation: tombstoneOperation,
		Relationship: &v1.Relationship{
			Resource: orgResource,
			Relation: "disabled",
			Subject:  userSubject,
		},
	})

	_, err := s.client.WriteRelationships(s.ctx, &v1.WriteRelationshipsRequest{
		Updates: relationshipUpdates,
	})

	err = spiceDbErrorToDomainError(err)

	return err
}

// NewConnection creates a new connection to an underlying SpiceDB store and saves it to the package variable conn
func (s *SpiceDbAccessRepository) NewConnection(spiceDbEndpoint string, token string, isBlocking, useTLS bool) error {

	var opts []grpc.DialOption

	if isBlocking {
		opts = append(opts, grpc.WithBlock())
	}

	if !useTLS {
		opts = append(opts, grpcutil.WithInsecureBearerToken(token))
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig, _ := grpcutil.WithSystemCerts(grpcutil.VerifyCA)
		opts = append(opts, grpcutil.WithBearerToken(token))
		opts = append(opts, tlsConfig)
	}

	client, err := authzed.NewClient(
		spiceDbEndpoint,
		opts...,
	)

	if err != nil {
		return err
	}

	s.client = client
	s.ctx = context.Background()
	return nil
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

func spiceDbErrorToDomainError(err error) error {
	if info, ok := unwrapSpiceDbError(err); ok {
		reasonValue := v1.ErrorReason_value[info.Reason]
		switch reasonValue {
		case int32(v1.ErrorReason_ERROR_REASON_WRITE_OR_DELETE_PRECONDITION_FAILURE):
			return domain.ErrConflict
		}
	}
	return err
}

func unwrapSpiceDbError(err error) (*errdetails.ErrorInfo, bool) {
	if s, ok := status.FromError(err); ok {
		if len(s.Details()) > 0 {
			if info := s.Details()[0].(*errdetails.ErrorInfo); ok {
				return info, true
			}
		}
	}

	return nil, false
}
