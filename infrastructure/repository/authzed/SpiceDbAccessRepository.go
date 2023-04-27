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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

// AssignSeats adds a range of assigned relations atomically.
func (s *SpiceDbAccessRepository) AssignSeats(subjectIDs []domain.SubjectID, orgID string, license *domain.License, svc domain.Service) error {

	// eventually below call is generic for assign/unassign, hence the separate function
	return s.assignSeatsAtomic(subjectIDs, orgID, license, svc)
}

func (s *SpiceDbAccessRepository) assignSeatsAtomic(subjectIDs []domain.SubjectID, orgID string, license *domain.License, svc domain.Service) error {
	//Step1 - Read the current License version
	resp, err := s.client.ReadRelationships(s.ctx, &v1.ReadRelationshipsRequest{
		Consistency: &v1.Consistency{Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true}},
		RelationshipFilter: &v1.RelationshipFilter{
			ResourceType:       LicenseObjectType,
			OptionalResourceId: fmt.Sprintf("%s/%s", orgID, svc.ID),
		},
	})

	if err != nil {
		glog.Errorf("Failed to read License relation :%v", err.Error())
		return err
	}

	var assignedCount int
	var currentLicenseVersion string
	for {
		v, err := resp.Recv()
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			glog.Errorf("Failed iterate License read response :%v", err.Error())
			return err
		}
		// The version is of the form: <Versionstring>/currentassignedseatscount
		if v.Relationship.Relation == "version" {
			glog.Infof("License - Version : %v", v.Relationship.Subject.Object.ObjectId)
			//spilt with "/" and the second part of the string is the current assigned count
			versionStrArr := strings.Split(v.Relationship.Subject.Object.ObjectId, "/")
			if len(versionStrArr) != 2 {
				return fmt.Errorf("invalid license version %s", v.Relationship.Subject.Object.ObjectId)
			}
			assignedCount, err = strconv.Atoi(versionStrArr[1])
			if err != nil {
				return err
			}
			currentLicenseVersion = versionStrArr[0]
		}
	}

	//prepare updates
	var relationshipUpdates []*v1.RelationshipUpdate

	//fill updates
	for _, subj := range subjectIDs {
		subject, object := createSubjectObjectTuple(SubjectType, string(subj), LicenseSeatObjectType, fmt.Sprintf("%s/%s", orgID, svc.ID))
		relationshipUpdates = append(relationshipUpdates, &v1.RelationshipUpdate{
			Operation: v1.RelationshipUpdate_OPERATION_CREATE, Relationship: &v1.Relationship{
				Subject:  subject,
				Resource: object,
				Relation: "assigned",
			}})
	}

	// Step 2 Create seat assignment relationships
	result, err := s.client.WriteRelationships(s.ctx, &v1.WriteRelationshipsRequest{
		Updates: relationshipUpdates,
	})

	if err != nil {
		glog.Errorf("Failed to assign relation :%v", err.Error())
		return err
	}

	glog.Infof("Assigned operation :%v", result)

	// Step 3 Delete the existing License - Version relationship
	_, err = s.client.DeleteRelationships(s.ctx, &v1.DeleteRelationshipsRequest{
		RelationshipFilter: &v1.RelationshipFilter{
			ResourceType:     LicenseObjectType,
			OptionalRelation: LicenseVersionStr,
			OptionalSubjectFilter: &v1.SubjectFilter{
				SubjectType:       LicenseVersionStr,
				OptionalSubjectId: fmt.Sprintf("%s/%d", currentLicenseVersion, assignedCount),
			},
		},
	})
	if err != nil {
		glog.Errorf("Failed to delete License Version relation :%v", err.Error())
		return err
	}
	glog.Infof("Deleted license version relation :%v", resp)

	//Step 4 - Write the new License - Version relationship
	//Get the old Data and perform the modification
	increment := true
	count := len(relationshipUpdates)
	if increment {
		assignedCount = assignedCount + count
	} else {
		assignedCount = assignedCount - count
	}
	err = s.writeLicenseVersionRelation(orgID, svc.ID, currentLicenseVersion, assignedCount)

	if err != nil {
		glog.Errorf("Failed to write new License version relation :%v", err.Error())
		return err
	}

	return nil
}

// AssignSeat create the relation
func (s *SpiceDbAccessRepository) AssignSeat(subjectID domain.SubjectID, orgID string, svc domain.Service) error {
	subject, object := createSubjectObjectTuple(SubjectType, string(subjectID), LicenseSeatObjectType, fmt.Sprintf("%s/%s", orgID, svc.ID))
	var relationshipUpdates = []*v1.RelationshipUpdate{
		{Operation: v1.RelationshipUpdate_OPERATION_CREATE, Relationship: &v1.Relationship{
			Subject:  subject,
			Resource: object,
			Relation: "assigned",
		}},
	}

	result, err := s.client.WriteRelationships(s.ctx, &v1.WriteRelationshipsRequest{
		Updates: relationshipUpdates,
	})

	if err != nil {
		glog.Errorf("Failed to assign relation :%v", err.Error())
		return err
	}

	glog.Infof("Assigned operation :%v", result)

	//Update the license version count - increment
	err = s.modifyLicenseSeatsVersionCount(orgID, svc.ID, 1, true)
	if err != nil {
		glog.Errorf("Failed to update license version relation :%v", err.Error())
		return err
	}

	return nil
}

// UnAssignSeats deletes a set of relations atomically, using preconditions for OCC
func (s *SpiceDbAccessRepository) UnAssignSeats(subjectIDs []domain.SubjectID, orgID string, license *domain.License, svc domain.Service) error {
	//prepare updates
	var relationshipUpdates []*v1.RelationshipUpdate

	//fill unassign updates
	for _, subj := range subjectIDs {
		subject, object := createSubjectObjectTuple(SubjectType, string(subj), LicenseSeatObjectType, fmt.Sprintf("%s/%s", orgID, svc.ID))
		relationshipUpdates = append(relationshipUpdates, &v1.RelationshipUpdate{
			Operation: v1.RelationshipUpdate_OPERATION_DELETE, Relationship: &v1.Relationship{
				Subject:  subject,
				Resource: object,
				Relation: "assigned",
			}})
	}

	//now that we have all at once, send updates
	result, err := s.client.WriteRelationships(s.ctx, &v1.WriteRelationshipsRequest{
		Updates: relationshipUpdates,
	})

	glog.Infof("Deleted relation :%v", result)

	if err != nil {
		glog.Errorf("Failed to delete relation :%v", err.Error())
		return err
	}

	//Update the license version count - decrement
	err = s.modifyLicenseSeatsVersionCount(orgID, svc.ID, 1, false)
	if err != nil {
		glog.Errorf("Failed to update license version relation :%v", err.Error())
		return err
	}
	return nil
}

// UnAssignSeat delete the relation
func (s *SpiceDbAccessRepository) UnAssignSeat(subjectID domain.SubjectID, orgID string, svc domain.Service) error {
	filter := &v1.RelationshipFilter{
		ResourceType:     LicenseSeatObjectType,
		OptionalRelation: "assigned",
		OptionalSubjectFilter: &v1.SubjectFilter{
			SubjectType:       SubjectType,
			OptionalSubjectId: string(subjectID),
		},
	}

	result, err := s.client.DeleteRelationships(s.ctx, &v1.DeleteRelationshipsRequest{
		RelationshipFilter: filter,
		OptionalPreconditions: []*v1.Precondition{
			{
				Operation: v1.Precondition_OPERATION_MUST_MATCH,
				Filter:    filter,
			},
		},
	})

	glog.Infof("Deleted relation :%v", result)

	if err != nil {
		glog.Errorf("Failed to delete relation :%v", err.Error())
		return err
	}

	//Update the license version count - decrement
	err = s.modifyLicenseSeatsVersionCount(orgID, svc.ID, 1, false)
	if err != nil {
		glog.Errorf("Failed to update license version relation :%v", err.Error())
		return err
	}
	return nil
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
			glog.Infof("License - Version : %v", v.Relationship.Subject.Object.ObjectId)
			//spilt with "/" and the second part of the string is the current assigned count
			versionStrArr := strings.Split(v.Relationship.Subject.Object.ObjectId, "/")
			if len(versionStrArr) != 2 {
				return nil, fmt.Errorf("invalid license version %s", v.Relationship.Subject.Object.ObjectId)
			}
			currentAssignedCount, err := strconv.Atoi(versionStrArr[1])
			if err != nil {
				return nil, err
			}
			license.InUse = currentAssignedCount
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

		ids = append(ids, domain.SubjectID(next.SubjectObjectId))
	}
	return ids, nil
}

func (s *SpiceDbAccessRepository) modifyLicenseSeatsVersionCount(orgID, serviceID string, count int, increment bool) error {
	//Step1 - Read the current License version
	resp, err := s.client.ReadRelationships(s.ctx, &v1.ReadRelationshipsRequest{
		Consistency: &v1.Consistency{Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true}},
		RelationshipFilter: &v1.RelationshipFilter{
			ResourceType:       LicenseObjectType,
			OptionalResourceId: fmt.Sprintf("%s/%s", orgID, serviceID),
		},
	})

	if err != nil {
		glog.Errorf("Failed to read License relation :%v", err.Error())
		return err
	}

	var assignedCount int
	var currentLicenseVersion string
	for {
		v, err := resp.Recv()
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			glog.Errorf("Failed iterate License read response :%v", err.Error())
			return err
		}
		// The version is of the form: <Versionstring>/currentassignedseatscount
		if v.Relationship.Relation == "version" {
			glog.Infof("License - Version : %v", v.Relationship.Subject.Object.ObjectId)
			//spilt with "/" and the second part of the string is the current assigned count
			versionStrArr := strings.Split(v.Relationship.Subject.Object.ObjectId, "/")
			if len(versionStrArr) != 2 {
				return fmt.Errorf("invalid license version %s", v.Relationship.Subject.Object.ObjectId)
			}
			assignedCount, err = strconv.Atoi(versionStrArr[1])
			if err != nil {
				return err
			}
			currentLicenseVersion = versionStrArr[0]
		}
	}

	// Step 2 Delete the existing License - Version relationship
	err = s.deleteLicenseVersionRelation(orgID, serviceID, currentLicenseVersion, assignedCount)
	if err != nil {
		glog.Errorf("Failed to delete old License version relation :%v", err.Error())
		return err
	}

	//Step 3 - Write the new License - Version relationship
	//Get the old Data and perform the modification
	if increment {
		assignedCount = assignedCount + count
	} else {
		assignedCount = assignedCount - count
	}
	err = s.writeLicenseVersionRelation(orgID, serviceID, currentLicenseVersion, assignedCount)

	if err != nil {
		glog.Errorf("Failed to write new License version relation :%v", err.Error())
		return err
	}
	return nil
}

func (s *SpiceDbAccessRepository) deleteLicenseVersionRelation(_, _, versionStr string, count int) error {
	resp, err := s.client.DeleteRelationships(s.ctx, &v1.DeleteRelationshipsRequest{
		RelationshipFilter: &v1.RelationshipFilter{
			ResourceType:     LicenseObjectType,
			OptionalRelation: LicenseVersionStr,
			OptionalSubjectFilter: &v1.SubjectFilter{
				SubjectType:       LicenseVersionStr,
				OptionalSubjectId: fmt.Sprintf("%s/%d", versionStr, count),
			},
		},
	})
	if err != nil {
		glog.Errorf("Failed to delete License Version relation :%v", err.Error())
		return err
	}
	glog.Infof("Deleted license version relation :%v", resp)
	return nil
}

func (s *SpiceDbAccessRepository) writeLicenseVersionRelation(orgID, srvcID, versionStr string, count int) error {

	subject, object := createSubjectObjectTuple(LicenseVersionStr, fmt.Sprintf("%s/%d", versionStr, count),
		LicenseObjectType, fmt.Sprintf("%s/%s", orgID, srvcID))
	var relationshipUpdates = []*v1.RelationshipUpdate{
		{Operation: v1.RelationshipUpdate_OPERATION_CREATE, Relationship: &v1.Relationship{
			Subject:  subject,
			Resource: object,
			Relation: LicenseVersionStr,
		}},
	}
	result, err := s.client.WriteRelationships(s.ctx, &v1.WriteRelationshipsRequest{
		Updates: relationshipUpdates,
	})
	if err != nil {
		glog.Errorf("Failed to create license version relation :%v", err.Error())
		return err
	}
	glog.Infof("License Version create operation :%v", result)
	return nil
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
