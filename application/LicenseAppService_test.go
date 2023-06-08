package application

import (
	"authz/domain"
	"authz/domain/contracts"
	"authz/infrastructure/repository/mock"
	"context"
	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrgEnablement(t *testing.T) {
	//Given
	service, _ := createService(nil)
	evt := OrgEntitledEvent{
		OrgID:     "o2",
		ServiceID: "smarts",
		MaxSeats:  2,
	}

	//When
	err := service.ProcessOrgEntitledEvent(evt)

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

func TestSameOrgAndServiceAddedTwiceNotPossible(t *testing.T) {
	//Given
	service, _ := createService(nil)
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
	err := service.ProcessOrgEntitledEvent(evt)
	assert.NoError(t, err)
	err = service.ProcessOrgEntitledEvent(evt2)
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

func createService(subjectRepositoryOverride contracts.SubjectRepository) (*LicenseAppService, *authzed.Client) {
	spiceDbRepo, authzedClient, err := spicedbContainer.CreateClient()
	if err != nil {
		panic(err)
	}

	principalRepo := &mock.StubPrincipalRepository{
		DefaultOrg: "o1",
		Principals: mock.GetMockPrincipalData(),
	}

	if subjectRepositoryOverride != nil {
		return NewLicenseAppService(spiceDbRepo, spiceDbRepo, principalRepo, subjectRepositoryOverride, spiceDbRepo), authzedClient
	}
	return NewLicenseAppService(spiceDbRepo, spiceDbRepo, principalRepo, principalRepo, spiceDbRepo), authzedClient
}

func TestBatchImportedDisabledUserDoesNotOverwriteEnabledUser(t *testing.T) {
	//Given
	mockRepo := InterruptableSubjectRepository{
		PreInterruptSubjects:  nil,
		PostInterruptSubjects: []domain.Subject{{SubjectID: "foo", Enabled: false}},
	}
	licenseAppService, spiceDbClient := createService(mockRepo)

	//When
	doneSignal := make(chan interface{})
	go func() {
		licenseAppService.ProcessOrgEntitledEvent(OrgEntitledEvent{
        		OrgID:     "myorg",
        		ServiceID: "myservice",
        		MaxSeats:  5,
        	})
		doneSignal <- "done"
		close(doneSignal)
	}()


	//now we wait for resume, and add the same subject manually
	spiceDbClient.PermissionsServiceClient.WriteRelationships(context.Background(), &v1.WriteRelationshipsRequest{
		//Updates: []v1.RelationshipUpdate{},
	}

	<- doneSignal
	//Then
}

type InterruptableSubjectRepository struct {
	StopSignal          chan interface{}
	resumeSignal          chan interface{}
	PreInterruptSubjects  []domain.Subject
	PostInterruptSubjects []domain.Subject
}

func (r InterruptableSubjectRepository) GetByOrgID(_ string) (chan domain.Subject, chan error) {
	subjects := make(chan domain.Subject)
	errors := make(chan error)

	go func() {
		if r.PreInterruptSubjects != nil {
			for _, s := range r.PreInterruptSubjects {
				subjects <- s
			}
		}
		r.StopSignal <- "stopped"
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
