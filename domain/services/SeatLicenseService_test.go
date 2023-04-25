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
		"aspian",
		[]string{"okay"},
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
	req := modifyLicRequestFromVars("okay", "aspian", []string{"usernext"}, []string{})
	err = lic.ModifySeats(req)

	//then
	var limitExceededErr domain.ErrLicenseLimitExceeded
	assert.ErrorAs(t, err, &limitExceededErr)
	assert.Equal(t, 5, limitExceededErr.MaxSeats)
	assert.Equal(t, 0, limitExceededErr.AvailableSeats)

	license, err := lic.GetLicense(domain.GetLicenseEvent{
		Requestor: "okay",
		OrgID:     "aspian",
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

	//when
	req := modifyLicRequestFromVars("okay", "aspian", []string{"usernext"}, []string{"user0"})
	err = lic.ModifySeats(req)

	//then
	assert.NoError(t, err)

	getevt := domain.GetLicenseEvent{
		Requestor: "okay",
		OrgID:     "aspian",
		ServiceID: "smarts",
	}
	license, err := lic.GetLicense(getevt)
	assert.NoError(t, err)
	assert.Equal(t, 0, license.GetAvailableSeats())

	seats, err := lic.GetAssignedSeats(getevt)
	assert.NoError(t, err)
	assert.Contains(t, seats, domain.SubjectID("usernext"))
	assert.NotContains(t, seats, domain.SubjectID("user0"))
}

func fillUpLicense(lic *SeatLicenseService) error {
	toAssign := make([]string, 5)
	for i := range toAssign {
		toAssign[i] = "user" + strconv.Itoa(i)
	}

	req := modifyLicRequestFromVars("okay", "aspian", toAssign, []string{})
	err := lic.ModifySeats(req)

	return err
}

func TestLicensingModifySeatsErrorsWhenNotAuthorized(t *testing.T) {
	t.SkipNow()
	req := modifyLicRequestFromVars("bad",
		"aspian",
		[]string{"okay"},
		[]string{})

	store := mockAuthzRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	err := lic.ModifySeats(req)

	assert.ErrorIs(t, err, domain.ErrNotAuthorized)
}

func TestLicensingAssignUnassignRoundTrip(t *testing.T) {
	addReq := modifyLicRequestFromVars("okay",
		"aspian",
		[]string{"okay"},
		[]string{})

	store := mockAuthzRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	authz, err := store.CheckAccess(addReq.Assign[0], "use", addReq.Service.AsResource())
	assert.NoError(t, err)
	assert.False(t, bool(authz), "Should not have been authorized without license.")

	err = lic.ModifySeats(addReq)
	assert.NoError(t, err)

	authz, err = store.CheckAccess(addReq.Assign[0], "use", addReq.Service.AsResource())
	assert.NoError(t, err)
	assert.True(t, bool(authz), "Should have been authorized with license.")

	remReq := modifyLicRequestFromVars("okay",
		"aspian",
		[]string{},
		[]string{"okay"})

	err = lic.ModifySeats(remReq)
	assert.NoError(t, err)

	authz, err = store.CheckAccess(remReq.UnAssign[0], "use", remReq.Service.AsResource())
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
