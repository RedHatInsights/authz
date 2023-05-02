package services

import (
	"authz/domain"
	"authz/domain/contracts"
	"errors"
	"strconv"
	"sync"
	"testing"

	"github.com/google/uuid"
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

func TestConcurrentIndividualRequestsCannotExceedLimit(t *testing.T) {
	//given
	store := mockAuthzRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	//when
	runCount := 5
	wait := &sync.WaitGroup{}
	errs := make(chan error, runCount)
	wait.Add(runCount)
	for i := 0; i < runCount; i++ {
		go func(run int) {
			req := modifyLicRequestFromVars("okay", "o1", []string{"user-" + strconv.Itoa(run)}, []string{})
			errs <- lic.ModifySeats(req)
			wait.Done()
		}(i)
	}
	wait.Wait()
	close(errs)
	spicedbContainer.WaitForQuantizationInterval()

	//then
	for err := range errs {
		if errors.Is(err, domain.ErrConflict) {
			continue
		}

		assert.NoError(t, err)
	}

	getevt := domain.GetLicenseEvent{
		Requestor: "okay",
		OrgID:     "o1",
		ServiceID: "smarts",
	}
	license, err := lic.GetLicense(getevt)
	assert.NoError(t, err)

	seats, err := lic.GetAssignedSeats(getevt)
	assert.NoError(t, err)
	assert.Equal(t, license.InUse, len(seats), "Expected is the number of seats allocated on the license, actual is the number of seats actually assigned.") //Ensure license count is accurate
}

func TestConcurrentRequestsCannotExceedLimit(t *testing.T) {
	//given
	store := mockAuthzRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	//when
	runCount := 5
	wait := &sync.WaitGroup{}
	errs := make(chan error, runCount)
	wait.Add(runCount)
	for i := 0; i < runCount; i++ {
		go func(run int) {
			subjects := make([]string, run+5) //Each run should have a different number of users so they don't compute the same license version
			for j := 0; j < len(subjects); j++ {
				subjects[j] = "user-" + uuid.NewString()
			}

			req := modifyLicRequestFromVars("okay", "o1", subjects, []string{})
			errs <- lic.ModifySeats(req)
			wait.Done()
		}(i)
	}
	wait.Wait()
	close(errs)
	spicedbContainer.WaitForQuantizationInterval()

	//then
	for err := range errs {
		if errors.Is(err, domain.ErrConflict) {
			continue
		}

		assert.NoError(t, err)
	}

	getevt := domain.GetLicenseEvent{
		Requestor: "okay",
		OrgID:     "o1",
		ServiceID: "smarts",
	}
	license, err := lic.GetLicense(getevt)
	assert.NoError(t, err)

	seats, err := lic.GetAssignedSeats(getevt)
	assert.NoError(t, err)
	assert.Equal(t, license.InUse, len(seats), "Expected is the number of seats allocated on the license, actual is the number of seats actually assigned.") //Ensure license count is accurate
}

func TestCanSwapWhenLicenseFull(t *testing.T) {
	//given
	store := mockAuthzRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	err := fillUpLicense(lic)
	assert.NoError(t, err)
	spicedbContainer.WaitForQuantizationInterval()
	//when
	req := modifyLicRequestFromVars("okay", "o1", []string{"usernext"}, []string{"user0"})
	err = lic.ModifySeats(req)
	assert.NoError(t, err)

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
	assert.ErrorIs(t, err, domain.ErrConflict)
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
