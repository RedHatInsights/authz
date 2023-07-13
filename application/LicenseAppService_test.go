package application

import (
	"authz/domain"
	"authz/domain/contracts"
	spicedb "authz/infrastructure/repository/authzed"
	"authz/infrastructure/repository/mock"
	"context"
	"testing"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"

	"github.com/stretchr/testify/assert"
)

func TestOrgEnablement(t *testing.T) {
	//Given
	service, _ := createService(nil, nil)
	evt := OrgEntitledEvent{
		OrgID:     "o2",
		ServiceID: "smarts",
		MaxSeats:  20,
	}

	//When
	err := service.HandleOrgEntitledEvent(evt)

	//Then
	assert.NoError(t, err)

	spicedbContainer.WaitForQuantizationInterval()

	limit, available, err := service.GetSeatAssignmentCounts(GetSeatAssignmentCountsRequest{
		Requestor: "system",
		OrgID:     evt.OrgID,
		ServiceID: evt.ServiceID,
	})
	assert.NoError(t, err)
	assert.Equal(t, evt.MaxSeats, limit)
	assert.Equal(t, evt.MaxSeats, available) //None in use

	assignable, err := service.GetSeatAssignments(GetSeatAssignmentRequest{
		Requestor:    "system",
		OrgID:        evt.OrgID,
		ServiceID:    evt.ServiceID,
		IncludeUsers: false,
		Assigned:     false,
	})
	assert.NoError(t, err)
	assert.Equal(t, 20, len(assignable))
}

func TestOrgEnablementWithNegativeSeats(t *testing.T) {
	service, _ := createService(nil, nil)
	evt := OrgEntitledEvent{
		OrgID:     "o1",
		ServiceID: "smarts",
		MaxSeats:  -1,
	}

	err := service.HandleOrgEntitledEvent(evt)

	var validationErr domain.ErrInvalidRequest

	assert.ErrorAs(t, err, &validationErr)
}

func TestOrgImportWithImproperOrgID(t *testing.T) {
	service, _ := createService(nil, nil)
	evt := ImportOrgEvent{
		OrgID: "!!!",
	}

	_, err := service.ImportUsersForOrg(evt)
	var validationErr domain.ErrInvalidRequest
	assert.ErrorAs(t, err, &validationErr)

}

func TestSameOrgAndServiceAddedTwiceNotPossible(t *testing.T) {
	//Given
	service, _ := createService(nil, nil)
	evt := OrgEntitledEvent{
		OrgID:     "o2",
		ServiceID: "smarts",
		MaxSeats:  2,
	}

	evt2 := OrgEntitledEvent{
		OrgID:     "o2",
		ServiceID: "smarts",
		MaxSeats:  1,
	}

	//When
	err := service.HandleOrgEntitledEvent(evt)
	assert.NoError(t, err)
	err = service.HandleOrgEntitledEvent(evt2)
	assert.Error(t, err)

	spicedbContainer.WaitForQuantizationInterval()

	limit, available, err := service.GetSeatAssignmentCounts(GetSeatAssignmentCountsRequest{
		Requestor: "system",
		OrgID:     evt.OrgID,
		ServiceID: evt.ServiceID,
	})
	assert.NoError(t, err)
	assert.Equal(t, evt.MaxSeats, limit)
	assert.Equal(t, evt.MaxSeats, available) //None in use

	assignable, err := service.GetSeatAssignments(GetSeatAssignmentRequest{
		Requestor:    "system",
		OrgID:        evt.OrgID,
		ServiceID:    evt.ServiceID,
		IncludeUsers: false,
		Assigned:     false,
	})
	assert.NoError(t, err)
	assert.Equal(t, 20, len(assignable))
}

func TestSubjectChangeEventForLicensedOrg(t *testing.T) {
	service, client := createService(nil, nil)

	subjectID := domain.SubjectID("new-subject")

	err := service.HandleSubjectAddOrUpdateEvent(contracts.SubjectAddOrUpdateEvent{
		SubjectID: string(subjectID),
		OrgID:     "o1",
		Active:    true,
	})

	assert.NoError(t, err)
	assert.True(t, spicedb.CheckForSubjectRelationship(client, subjectID, "member", "org", "o1"))
}

func TestSubjectChangeEventForUnlicensedOrg(t *testing.T) {
	service, client := createService(nil, nil)

	subjectID := domain.SubjectID("new-subject")

	err := service.HandleSubjectAddOrUpdateEvent(contracts.SubjectAddOrUpdateEvent{
		SubjectID: string(subjectID),
		OrgID:     "new-org",
		Active:    true,
	})

	assert.NoError(t, err)
	assert.False(t, spicedb.CheckForSubjectRelationship(client, subjectID, "member", "org", "new-org"))
}

func createService(subjectRepositoryOverride contracts.SubjectRepository, orgRepositoryOverride contracts.OrganizationRepository) (*LicenseAppService, *authzed.Client) {
	spiceDbRepo, authzedClient, err := spicedbContainer.CreateClient()
	if err != nil {
		panic(err)
	}

	principalRepo := &mock.StubPrincipalRepository{
		DefaultOrg: "o1",
		Principals: mock.GetMockPrincipalData(),
	}

	var subjectRepo contracts.SubjectRepository
	subjectRepo = principalRepo

	if subjectRepositoryOverride != nil {
		subjectRepo = subjectRepositoryOverride
	}

	var orgRepo contracts.OrganizationRepository
	orgRepo = spiceDbRepo

	if orgRepositoryOverride != nil {
		orgRepo = orgRepositoryOverride
	}

	return NewLicenseAppService(spiceDbRepo, spiceDbRepo, principalRepo, subjectRepo, orgRepo), authzedClient
}

func TestBatchImportedDisabledUserDoesNotOverwriteEnabledUser(t *testing.T) {
	//Given
	mockSubjectRepo := &InterruptableSubjectRepository{
		PreInterruptSubjects:  nil,
		PostInterruptSubjects: []domain.Subject{{SubjectID: "foo", Enabled: false}},
		resumeSignal:          make(chan interface{}),
		StoppedSignal:         make(chan interface{}),
	}
	licenseAppService, spiceDbClient := createService(mockSubjectRepo, nil)

	//When
	doneSignal := make(chan interface{})
	go func() {
		err := licenseAppService.HandleOrgEntitledEvent(OrgEntitledEvent{
			OrgID:     "myorg",
			ServiceID: "myservice",
			MaxSeats:  5,
		})
		assert.NoError(t, err)
		doneSignal <- "done"
		close(doneSignal)
	}()

	<-mockSubjectRepo.StoppedSignal //Wait for the import to reach the pause

	//Add the user directly to SpiceDB as enabled
	_, err := spiceDbClient.WriteRelationships(context.Background(), &v1.WriteRelationshipsRequest{
		Updates: []*v1.RelationshipUpdate{{
			Operation: v1.RelationshipUpdate_OPERATION_TOUCH,
			Relationship: &v1.Relationship{
				Resource: &v1.ObjectReference{
					ObjectType: "org",
					ObjectId:   "myorg",
				},
				Relation: "member",
				Subject: &v1.SubjectReference{
					Object: &v1.ObjectReference{
						ObjectType: "user",
						ObjectId:   "foo",
					},
				},
			}},
		}})
	assert.NoError(t, err)

	mockSubjectRepo.Resume() //Allow import to continue

	<-doneSignal //Wait for import to finish
	//Then
	assert.True(t, getEnabled(spiceDbClient, "foo", "myorg")) //Assert user is still enabled

}

func getEnabled(client *authzed.Client, subjectID string, orgID string) bool {
	resp, err := client.CheckPermission(context.Background(), &v1.CheckPermissionRequest{
		Consistency: &v1.Consistency{Requirement: &v1.Consistency_FullyConsistent{FullyConsistent: true}},
		Resource: &v1.ObjectReference{
			ObjectType: "org",
			ObjectId:   orgID,
		},
		Permission: "disabled",
		Subject: &v1.SubjectReference{
			Object: &v1.ObjectReference{
				ObjectType: "user",
				ObjectId:   subjectID,
			},
		},
	})

	if err != nil {
		panic(err)
	}

	if resp.Permissionship == v1.CheckPermissionResponse_PERMISSIONSHIP_NO_PERMISSION {
		return true //Not disabled
	}
	return false
}

type InterruptableSubjectRepository struct {
	StoppedSignal         chan interface{}
	resumeSignal          chan interface{}
	PreInterruptSubjects  []domain.Subject
	PostInterruptSubjects []domain.Subject
}

func (r *InterruptableSubjectRepository) GetByOrgID(_ string) (chan domain.Subject, chan error) {
	subjects := make(chan domain.Subject)
	errors := make(chan error)

	go func() {
		if r.PreInterruptSubjects != nil {
			for _, s := range r.PreInterruptSubjects {
				subjects <- s
			}
		}
		r.StoppedSignal <- "stopped"
		<-r.resumeSignal
		if r.PostInterruptSubjects != nil {
			for _, s := range r.PostInterruptSubjects {
				subjects <- s
			}
		}
		close(subjects)
		close(errors)
	}()

	return subjects, errors
}

func (r InterruptableSubjectRepository) Resume() {
	r.resumeSignal <- "go resume it!"
}
