package application

import (
	"authz/infrastructure/repository/mock"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrgEnablement(t *testing.T) {
	//Given
	service := createService()
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
	service := createService()
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

func createService() *LicenseAppService {
	spiceDbRepo, err := spicedbContainer.CreateClient()
	if err != nil {
		panic(err)
	}

	principalRepo := &mock.StubPrincipalRepository{
		DefaultOrg: "o1",
		Principals: mock.GetMockPrincipalData(),
	}

	return NewLicenseAppService(spiceDbRepo, spiceDbRepo, principalRepo, principalRepo, spiceDbRepo)
}
