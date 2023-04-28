package services

import (
	"authz/domain"
	"authz/domain/contracts"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicensingModifySeatsErrorsWhenNotAuthenticated(t *testing.T) {
	req := modifyLicRequestFromVars("",
		"o1",
		[]string{"u2"},
		[]string{})

	store := mockAuthzRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	err := lic.ModifySeats(req)

	assert.ErrorIs(t, err, domain.ErrNotAuthenticated)
}

func TestSeatLicenseOverAssignment(t *testing.T) {
	//given
	store := mockAuthzRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)
	err := fillUpLicense(lic)
	assert.NoError(t, err)

	//when
	req := modifyLicRequestFromVars("okay", "o1", []string{"usernext"}, []string{})
	err = lic.ModifySeats(req)

	//then
	assert.ErrorIs(t, err, domain.ErrLicenseLimitExceeded)
	license, err := lic.GetLicense(domain.GetLicenseEvent{
		Requestor: "okay",
		OrgID:     "o1",
		ServiceID: "smarts",
	})

	assert.NoError(t, err)
	assert.Equal(t, 0, license.GetAvailableSeats())
}

func TestCanSwapUsersWhenLicenseFullyAllocated(t *testing.T) {
	//given
	store := mockAuthzRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	err := fillUpLicense(lic)
	assert.NoError(t, err)
	spicedbContainer.WaitForQuantizationInterval()
	//when
	req := modifyLicRequestFromVars("okay", "o1", []string{"usernext"}, []string{"user0"})
	err = lic.ModifySeats(req)

	//then
	spicedbContainer.WaitForQuantizationInterval()
	assert.NoError(t, err)

	getevt := domain.GetLicenseEvent{
		Requestor: "okay",
		OrgID:     "o1",
		ServiceID: "smarts",
	}
	license, err := lic.GetLicense(getevt)
	assert.NoError(t, err)
	assert.Equal(t, 0, license.GetAvailableSeats())

	//Flaky due to read-after-write consistency- this call may not reflect changes
	seats, err := lic.GetAssignedSeats(getevt)
	assert.NoError(t, err)
	assert.Contains(t, seats, domain.SubjectID("usernext"))
	assert.NotContains(t, seats, domain.SubjectID("user0"))
}

func TestCantUnassignSeatThatWasNotAssigned(t *testing.T) {
	//given
	store := mockAuthzRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	// when
	req := modifyLicRequestFromVars("okay", "o1", []string{}, []string{"not_assigned"})
	err := lic.ModifySeats(req)

	// then
	assert.Error(t, err)
	license, err := lic.GetLicense(domain.GetLicenseEvent{
		Requestor: "okay",
		OrgID:     "o1",
		ServiceID: "smarts",
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, license.InUse)
}

func fillUpLicense(lic *SeatLicenseService) error {
	toAssign := make([]string, 9) //9 free slots in seed data
	for i := range toAssign {
		toAssign[i] = "user" + strconv.Itoa(i)
	}

	req := modifyLicRequestFromVars("okay", "o1", toAssign, []string{})
	err := lic.ModifySeats(req)

	return err
}

func TestLicensingModifySeatsErrorsWhenNotAuthorized(t *testing.T) {
	t.SkipNow() //Skip until meta-authz is in place
	req := modifyLicRequestFromVars("bad",
		"o1",
		[]string{"okay"},
		[]string{})

	store := mockAuthzRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	err := lic.ModifySeats(req)

	assert.ErrorIs(t, err, domain.ErrNotAuthorized)
}

func TestLicensingAssignUnassignRoundTrip(t *testing.T) {
	addReq := modifyLicRequestFromVars("okay",
		"o1",
		[]string{"okay"},
		[]string{})

	store := mockAuthzRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)
	license := domain.Resource{Type: "license", ID: "o1/smarts"}

	authz, err := store.CheckAccess(addReq.Assign[0], "access", license)
	assert.NoError(t, err)
	assert.False(t, bool(authz), "Should not have been authorized without license.")

	err = lic.ModifySeats(addReq)
	assert.NoError(t, err)

	spicedbContainer.WaitForQuantizationInterval()

	authz, err = store.CheckAccess(addReq.Assign[0], "access", license)
	assert.NoError(t, err)
	assert.True(t, bool(authz), "Should have been authorized with license.")

	remReq := modifyLicRequestFromVars("okay",
		"o1",
		[]string{},
		[]string{"okay"})

	err = lic.ModifySeats(remReq)
	assert.NoError(t, err)

	spicedbContainer.WaitForQuantizationInterval()

	authz, err = store.CheckAccess(addReq.Assign[0], "access", license)
	assert.NoError(t, err)
	assert.False(t, bool(authz), "Should not have been authorized without license.")
}

func modifyLicRequestFromVars(requestorID string, subjectOrg string, assign []string, unassign []string) domain.ModifySeatAssignmentEvent {
	evt := domain.ModifySeatAssignmentEvent{
		Request: domain.Request{
			Requestor: domain.SubjectID(requestorID),
		},
		Org:     domain.Organization{ID: subjectOrg},
		Service: domain.Service{ID: "smarts"},
	}

	evt.Assign = make([]domain.SubjectID, len(assign))
	for i, id := range assign {
		evt.Assign[i] = domain.SubjectID(id)
	}

	evt.UnAssign = make([]domain.SubjectID, len(unassign))
	for i, id := range unassign {
		evt.UnAssign[i] = domain.SubjectID(id)
	}

	return evt
}
