// Package authzed contains the technical implementations for the accessRepo from authzed spicedb
package authzed

import (
	"authz/domain"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/golang/glog"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// SubjectType user
const SubjectType = "user"

// LicenseSeatObjectType license_seats
const LicenseSeatObjectType = "license_seats"

// LicenseObjectType - License object
const LicenseObjectType = "license"

// LicenseVersionStr - License Version realation
const LicenseVersionStr = "version"

// SpiceDbAccessRepository -
type SpiceDbAccessRepository struct {
	authzedClient
}

// authzedClient - Authz client struct
type authzedClient struct {
	client *authzed.Client
	ctx    context.Context
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

// ModifySeats atomically persists changes to seat assignments for a license
func (s *SpiceDbAccessRepository) ModifySeats(assignedSubjectIDs []domain.SubjectID, removedSubjectIDs []domain.SubjectID, license *domain.License, orgID string, svc domain.Service) error {
	// Step 1 Add seat changes
	var relationshipUpdates []*v1.RelationshipUpdate

	var preconditions []*v1.Precondition
	assignedCount := license.InUse

	for _, subj := range assignedSubjectIDs {
		relationshipUpdates = append(relationshipUpdates, createUserSeatAssignmentRelationshipUpdate(
			v1.RelationshipUpdate_OPERATION_CREATE,
			subj,
			orgID,
			svc))

		assignedCount++
	}

	for _, subj := range removedSubjectIDs {
		relationshipUpdates = append(relationshipUpdates, createUserSeatAssignmentRelationshipUpdate(
			v1.RelationshipUpdate_OPERATION_DELETE,
			subj,
			orgID,
			svc))
		preconditions = append(preconditions, createSeatAssignedPrecondition(subj, orgID, svc))

		assignedCount--
	}

	// Step 2 Add license changes
	relationshipUpdates, preconditions = addLicenseVersionSwap(relationshipUpdates, preconditions, license, assignedCount)

	// Step 3 submit transaction
	result, err := s.client.WriteRelationships(s.ctx, &v1.WriteRelationshipsRequest{
		Updates:               relationshipUpdates,
		OptionalPreconditions: preconditions,
	})

	// Step 4 examine any errors
	if err != nil {
		glog.Errorf("Failed to write modify seats :%v", err.Error())

		return spiceDbErrorToDomainError(err)
	}

	glog.Infof("Assigned operation :%v", result)

	return nil
}

func createUserSeatAssignmentRelationshipUpdate(operation v1.RelationshipUpdate_Operation, subj domain.SubjectID, orgID string, svc domain.Service) *v1.RelationshipUpdate {
	subject, object := createSubjectObjectTuple(SubjectType, string(subj), LicenseSeatObjectType, fmt.Sprintf("%s/%s", orgID, svc.ID))
	return &v1.RelationshipUpdate{
		Operation: operation, Relationship: &v1.Relationship{
			Subject:  subject,
			Resource: object,
			Relation: "assigned",
		}}
}

func addLicenseVersionSwap(updates []*v1.RelationshipUpdate, conditions []*v1.Precondition, lic *domain.License, newCount int) ([]*v1.RelationshipUpdate, []*v1.Precondition) {
	licenseObj := createObjectFromLicense(lic)
	oldVersionSubj := createSubjectFromLicenseAndCount(lic, lic.InUse)

	conditions = append(conditions, &v1.Precondition{
		Operation: v1.Precondition_OPERATION_MUST_MATCH,
		Filter: &v1.RelationshipFilter{
			ResourceType:       licenseObj.ObjectType,
			OptionalResourceId: licenseObj.ObjectId,
			OptionalRelation:   LicenseVersionStr,
			OptionalSubjectFilter: &v1.SubjectFilter{
				SubjectType:       oldVersionSubj.Object.ObjectType,
				OptionalSubjectId: oldVersionSubj.Object.ObjectId,
			},
		},
	})

	if lic.InUse == newCount {
		return updates, conditions //The version is essentially just the count at the moment, so a swap isn't technically a change. This may change in the future!
	}

	updates = append(updates, &v1.RelationshipUpdate{
		Operation: v1.RelationshipUpdate_OPERATION_DELETE,
		Relationship: &v1.Relationship{
			Resource: licenseObj,
			Relation: LicenseVersionStr,
			Subject:  oldVersionSubj,
		},
	})

	updates = append(updates, &v1.RelationshipUpdate{
		Operation: v1.RelationshipUpdate_OPERATION_CREATE,
		Relationship: &v1.Relationship{
			Resource: licenseObj,
			Relation: LicenseVersionStr,
			Subject:  createSubjectFromLicenseAndCount(lic, newCount),
		},
	})

	return updates, conditions
}

func createSubjectFromLicenseAndCount(lic *domain.License, count int) *v1.SubjectReference {
	return &v1.SubjectReference{
		Object: &v1.ObjectReference{
			ObjectType: LicenseVersionStr,
			ObjectId:   fmt.Sprintf("%s/%d", lic.Version, count),
		},
	}
}

func createObjectFromLicense(lic *domain.License) *v1.ObjectReference {
	return &v1.ObjectReference{
		ObjectType: LicenseObjectType,
		ObjectId:   fmt.Sprintf("%s/%s", lic.OrgID, lic.ServiceID),
	}
}

func createSeatAssignedPrecondition(subj domain.SubjectID, orgID string, svc domain.Service) *v1.Precondition {
	return &v1.Precondition{
		Operation: v1.Precondition_OPERATION_MUST_MATCH,
		Filter: &v1.RelationshipFilter{
			ResourceType:       LicenseSeatObjectType,
			OptionalResourceId: fmt.Sprintf("%s/%s", orgID, svc.ID),
			OptionalRelation:   "assigned",
			OptionalSubjectFilter: &v1.SubjectFilter{
				SubjectType:       SubjectType,
				OptionalSubjectId: string(subj),
			},
		},
	}
}

// GetLicense - Get the current license infoarmation
func (s *SpiceDbAccessRepository) GetLicense(orgID string, serviceID string) (*domain.License, error) {
	var license domain.License
	resp, err := s.client.ReadRelationships(s.ctx, &v1.ReadRelationshipsRequest{
		Consistency: &v1.Consistency{Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true}},
		RelationshipFilter: &v1.RelationshipFilter{
			ResourceType:       LicenseObjectType,
			OptionalResourceId: fmt.Sprintf("%s/%s", orgID, serviceID),
		},
	})

	if err != nil {
		glog.Errorf("Failed to read License relation :%v", err.Error())
		return nil, err
	}

	for {
		v, err := resp.Recv()
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			glog.Errorf("Failed iterate License read response :%v", err.Error())
			return nil, err
		}
		// The Max relation is read to extract the MAx count of the license
		if v.Relationship.Relation == "max" {
			glog.Infof("License - Max count: %v", v.Relationship.Subject.Object.ObjectId)
			license.MaxSeats, err = strconv.Atoi(v.Relationship.Subject.Object.ObjectId)

			if err != nil {
				return nil, err
			}
		}
		// The version is of the form: <Versionstring>/currentassignedseatscount
		if v.Relationship.Relation == "version" {
			version := v.Relationship.Subject.Object.ObjectId
			glog.Infof("License - Version : %v", version)
			//spilt with "/" and the second part of the string is the current assigned count
			versionStrArr := strings.Split(version, "/")
			if len(versionStrArr) != 2 {
				return nil, fmt.Errorf("invalid license version %s", v.Relationship.Subject.Object.ObjectId)
			}
			currentAssignedCount, err := strconv.Atoi(versionStrArr[1])
			if err != nil {
				return nil, err
			}
			license.InUse = currentAssignedCount
			license.Version = versionStrArr[0]
		}
		license.OrgID = orgID
		license.ServiceID = serviceID
	}

	return &license, nil
}

// GetAssigned - todo implementation
func (s *SpiceDbAccessRepository) GetAssigned(orgID string, serviceID string) ([]domain.SubjectID, error) {
	result, err := s.client.LookupSubjects(s.ctx, &v1.LookupSubjectsRequest{
		Resource: &v1.ObjectReference{
			ObjectType: LicenseObjectType,
			ObjectId:   fmt.Sprintf("%s/%s", orgID, serviceID),
		},
		Permission:        "access",
		SubjectObjectType: SubjectType,
	})

	if err != nil {
		return nil, err
	}

	ids := make([]domain.SubjectID, 0)
	for {
		next, err := result.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		ids = append(ids, domain.SubjectID(next.Subject.SubjectObjectId))
	}
	return ids, nil
}

// NewConnection creates a new connection to an underlying SpiceDB store and saves it to the package variable conn
func (s *SpiceDbAccessRepository) NewConnection(spiceDbEndpoint string, token string, isBlocking, useTLS bool) {

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
		log.Fatalf("unable to initialize client: %s", err)
	}

	s.client = client
	s.ctx = context.Background()
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
		switch info.Reason {
		case "ERROR_REASON_WRITE_OR_DELETE_PRECONDITION_FAILURE":
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
