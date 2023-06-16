// Package authzed contains the technical implementations for the accessRepo from authzed spicedb
package authzed

import (
	"authz/domain"
	"authz/infrastructure/grpcutil"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/golang/glog"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
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
	// LicenseSeatObjectType license_seats relation
	LicenseSeatObjectType = "license_seats"
	// LicenseObjectType - License relation
	LicenseObjectType = "license"
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
		preconditions = append(preconditions, createUserNotDisabledPrecondition(subj, orgID), createUserIsMemberOfOrgPrecondition(subj, orgID))

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

func createUserIsMemberOfOrgPrecondition(subj domain.SubjectID, orgID string) *v1.Precondition {
	return &v1.Precondition{
		Operation: v1.Precondition_OPERATION_MUST_MATCH,
		Filter: &v1.RelationshipFilter{
			ResourceType:       OrgType,
			OptionalResourceId: orgID,
			OptionalRelation:   "member",
			OptionalSubjectFilter: &v1.SubjectFilter{
				SubjectType:       SubjectType,
				OptionalSubjectId: string(subj),
			},
		},
	}
}

func createUserNotDisabledPrecondition(subj domain.SubjectID, orgID string) *v1.Precondition {
	return &v1.Precondition{
		Operation: v1.Precondition_OPERATION_MUST_NOT_MATCH,
		Filter: &v1.RelationshipFilter{
			ResourceType:       OrgType,
			OptionalResourceId: orgID,
			OptionalRelation:   "disabled",
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

// GetAssignable returns assignable seats for a given organization ID and service ID (which are not already assigned)
func (s *SpiceDbAccessRepository) GetAssignable(orgID string, serviceID string) ([]domain.SubjectID, error) {
	result, err := s.client.LookupSubjects(s.ctx, &v1.LookupSubjectsRequest{
		Resource: &v1.ObjectReference{
			ObjectType: LicenseObjectType,
			ObjectId:   fmt.Sprintf("%s/%s", orgID, serviceID),
		},
		Permission:        "assignable",
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

// GetAssigned returns assigned seats for a given organization ID and service ID
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

// ApplyLicense stores the given license associated with its service and organization
func (s *SpiceDbAccessRepository) ApplyLicense(license *domain.License) error {
	licenseID := fmt.Sprintf("%s/%s", license.OrgID, license.ServiceID)
	licenseResource := &v1.ObjectReference{
		ObjectType: LicenseObjectType,
		ObjectId:   licenseID,
	}

	_, err := s.client.WriteRelationships(s.ctx, &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{
			{
				Operation: v1.RelationshipUpdate_OPERATION_CREATE,
				Relationship: &v1.Relationship{
					Resource: licenseResource,
					Relation: "max",
					Subject: &v1.SubjectReference{
						Object: &v1.ObjectReference{
							ObjectType: "max",
							ObjectId:   strconv.Itoa(license.MaxSeats),
						},
					},
				},
			},
			{
				Operation: v1.RelationshipUpdate_OPERATION_CREATE,
				Relationship: &v1.Relationship{
					Resource: licenseResource,
					Relation: "seats",
					Subject: &v1.SubjectReference{
						Object: &v1.ObjectReference{
							ObjectType: LicenseSeatObjectType,
							ObjectId:   licenseID,
						},
					},
				},
			},
			{
				Operation: v1.RelationshipUpdate_OPERATION_CREATE,
				Relationship: &v1.Relationship{
					Resource: licenseResource,
					Relation: "version",
					Subject: &v1.SubjectReference{
						Object: &v1.ObjectReference{
							ObjectType: LicenseVersionStr,
							ObjectId:   fmt.Sprintf("%s/%d", license.Version, license.InUse),
						},
					},
				},
			},
			{
				Operation: v1.RelationshipUpdate_OPERATION_CREATE,
				Relationship: &v1.Relationship{
					Resource: licenseResource,
					Relation: "org",
					Subject: &v1.SubjectReference{
						Object: &v1.ObjectReference{
							ObjectType: OrgType,
							ObjectId:   license.OrgID,
						},
					},
				},
			},
		},
		OptionalPreconditions: []*v1.Precondition{{
			Operation: v1.Precondition_OPERATION_MUST_NOT_MATCH,
			Filter: &v1.RelationshipFilter{
				ResourceType:       licenseResource.ObjectType,
				OptionalResourceId: licenseResource.ObjectId,
			},
		}},
	})

	return err
}

// AddSubject stores a subject associated with an organization
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

// IsImported returns true if an org has at least one persisted, existing license or at least one member.
func (s *SpiceDbAccessRepository) IsImported(orgID string) (bool, error) {
	resp, err := s.client.LookupResources(s.ctx, &v1.LookupResourcesRequest{
		Consistency:        useFullConsistency(),
		ResourceObjectType: LicenseObjectType,
		Permission:         "org",
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: "org",
				ObjectId:   orgID,
			},
		},
		OptionalLimit: 1,
	})

	if err != nil {
		glog.Errorf("Error checking if an org has a license. Could not call spiceDB! %v", err)
		return false, err
	}

	_, err = resp.Recv()

	if errors.Is(err, io.EOF) {
		//if no license found, check if an org with members exists inside the schema
		result, err := hasOrgMembers(s.ctx, s.client, orgID)
		return result, err
	}
	if err != nil {
		return false, err
	}
	//if a license is found, assume the org is already imported. return true
	return true, nil
}

// hasOrgMembers returns true if at least one member exists for an org in spiceDB
func hasOrgMembers(ctx context.Context, client *authzed.Client, orgID string) (bool, error) {
	// zed lookup-subjects org:o2 member user
	resp, err := client.LookupSubjects(ctx, &v1.LookupSubjectsRequest{
		Consistency: useFullConsistency(),
		Resource: &v1.ObjectReference{
			ObjectType: OrgType,
			ObjectId:   orgID,
		},
		Permission:        "member",
		SubjectObjectType: "user",
	})

	if err != nil {
		glog.Errorf("Error checking if an org has members. Could not call spiceDB! %v", err)
		return false, err
	}

	_, e := resp.Recv()

	// if stream ends immediately, no member found -> return false
	if errors.Is(e, io.EOF) {
		return false, nil
	}

	if e != nil {
		return false, e
	}
	// else member found, return true
	return true, nil
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

func useFullConsistency() *v1.Consistency {
	return &v1.Consistency{Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true}}
}
