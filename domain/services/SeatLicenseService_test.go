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

	store := spicedbSeatLicenseRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	err := lic.ModifySeats(req)

	assert.ErrorIs(t, err, domain.ErrNotAuthenticated)
}

func TestNewAssignedUserNotAssignableButNewUnassignedUserIs(t *testing.T) {
	//given (see schema/spicedb_bootstrap_relations.yaml for initial seed data)
	store := spicedbSeatLicenseRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	// when
	req := modifyLicRequestFromVars("okay", "o1", []string{"u2"}, []string{"u1"})
	err := lic.ModifySeats(req)
	assert.NoError(t, err)

	// then
	getevt := domain.GetLicenseEvent{
		Requestor: "okay",
		OrgID:     "o1",
		ServiceID: "smarts",
	}
	assignable, err := lic.GetAssignableSeats(getevt)
	assert.NoError(t, err)
	expectedAssignableUsers := []domain.SubjectID{"u1", "u5", "u6", "u7", "u8", "u9", "u10", "u11", "u12", "u13", "u14", "u15", "u16", "u17", "u18", "u19", "u20"}
	assert.ElementsMatch(t, expectedAssignableUsers, assignable)

	assigned, err := lic.GetAssignedSeats(getevt)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []domain.SubjectID{"u2", "u3"}, assigned)
}

func TestDisabledUserNotAssignable(t *testing.T) {
	//given
	store := spicedbSeatLicenseRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	// when
	req := modifyLicRequestFromVars("okay", "o1", []string{"u3"}, nil)
	err := lic.ModifySeats(req)

	// then
	assert.Error(t, err)
}

func TestNonExistentUserNotAssignable(t *testing.T) {
	//given
	store := spicedbSeatLicenseRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	// when
	req := modifyLicRequestFromVars("okay", "o1", []string{"u777"}, nil)
	err := lic.ModifySeats(req)

	// then
	assert.Error(t, err)
}

func TestDuplicateUserAssignmentNotAllowed(t *testing.T) {
	//given
	store := spicedbSeatLicenseRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	//when
	req := modifyLicRequestFromVars("okay", "o1", []string{"u1"}, nil)
	err := lic.ModifySeats(req)

	//then
	assert.Error(t, err)
}

func TestSeatLicenseOverAssignment(t *testing.T) {
	//given
	store := spicedbSeatLicenseRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)
	err := fillUpLicense(lic)
	assert.NoError(t, err)

	//when
	req := modifyLicRequestFromVars("okay", "o1", []string{"u5"}, []string{})
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

func TestConcurrentSwapsCannotReplaceTheSameUser(t *testing.T) {
	//given
	store := spicedbSeatLicenseRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	//when
	runCount := 5
	wait := &sync.WaitGroup{}
	errs := make(chan error, runCount)
	wait.Add(runCount)
	for i := 0; i < runCount; i++ {
		go func(run int) {
			req := modifyLicRequestFromVars("okay", "o1", []string{"user-" + strconv.Itoa(run)}, []string{"u1"})
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

	assertLicenseCountIsCorrect(t, lic)
}

func TestDisabledAssignedUsersCanBeUnassigned(t *testing.T) {
	//given
	store := spicedbSeatLicenseRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)
	req := modifyLicRequestFromVars("okay", "o1", []string{}, []string{"u3"}) //u3 is assigned and disabled

	//when
	err := lic.ModifySeats(req)

	//then
	assert.NoError(t, err)
	spicedbContainer.WaitForQuantizationInterval()

	assigned, err := lic.GetAssignedSeats(domain.GetLicenseEvent{
		Requestor: "okay",
		OrgID:     "o1",
		ServiceID: "smarts",
	})
	assert.NoError(t, err)
	assert.NotContains(t, assigned, domain.SubjectID("u3"))
}

func TestConcurrentIndividualRequestsCannotExceedLimit(t *testing.T) {
	//given
	store := spicedbSeatLicenseRepository()
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

	assertLicenseCountIsCorrect(t, lic)
}

func TestConcurrentRequestsCannotExceedLimit(t *testing.T) {
	//given
	store := spicedbSeatLicenseRepository()
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
		if errors.Is(err, domain.ErrConflict) || errors.Is(err, domain.ErrLicenseLimitExceeded) {
			continue
		}

		assert.NoError(t, err)
	}

	assertLicenseCountIsCorrect(t, lic)
}

func assertLicenseCountIsCorrect(t *testing.T, lic *SeatLicenseService) {
	getevt := domain.GetLicenseEvent{
		Requestor: "okay",
		OrgID:     "o1",
		ServiceID: "smarts",
	}
	license, err := lic.GetLicense(getevt)
	assert.NoError(t, err)

	seats, err := lic.GetAssignedSeats(getevt)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, license.InUse, 0)
	assert.LessOrEqual(t, license.InUse, license.MaxSeats)
	assert.Equal(t, license.InUse, len(seats), "Expected is the number of seats allocated on the license, actual is the number of seats actually assigned.") //Ensure license count is accurate
}

// Added in place of an input validation rule that at least one of the slices has content
func TestModifySeatsRequestWithNoSubjectsIsNoOp(t *testing.T) {
	store := spicedbSeatLicenseRepository()
	service := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	getevt := domain.GetLicenseEvent{
		Requestor: "system",
		OrgID:     "o1",
		ServiceID: "smarts",
	}
	licBefore, err := service.GetLicense(getevt)
	assert.NoError(t, err)
	seatsBefore, err := service.GetAssignedSeats(getevt)
	assert.NoError(t, err)

	err = service.ModifySeats(modifyLicRequestFromVars("system", "o1", []string{}, []string{}))
	assert.NoError(t, err)

	licAfter, err := service.GetLicense(getevt)
	assert.NoError(t, err)
	seatsAfter, err := service.GetAssignedSeats(getevt)
	assert.NoError(t, err)

	assert.Equal(t, licBefore.InUse, licAfter.InUse)
	assert.ElementsMatch(t, seatsBefore, seatsAfter)
}

// Added in place of an input validation rule that the slices do not overlap
func TestModifySeatsRequestWithSwappingSameUser(t *testing.T) {
	store := spicedbSeatLicenseRepository()
	service := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	getevt := domain.GetLicenseEvent{
		Requestor: "system",
		OrgID:     "o1",
		ServiceID: "smarts",
	}
	licBefore, err := service.GetLicense(getevt)
	assert.NoError(t, err)
	seatsBefore, err := service.GetAssignedSeats(getevt)
	assert.NoError(t, err)

	err = service.ModifySeats(modifyLicRequestFromVars("system", "o1", []string{"noone_in_particular"}, []string{"noone_in_particular"}))
	assert.Error(t, err) //Should fail on contradicting updates: rpc error: code = InvalidArgument desc = found more than one update with relationship `license_seats:o1/smarts#assigned@user:noone_in_particular` in this request; a relationship can only be specified in an update once per overall WriteRelationships request

	licAfter, err := service.GetLicense(getevt)
	assert.NoError(t, err)
	seatsAfter, err := service.GetAssignedSeats(getevt)
	assert.NoError(t, err)

	assert.Equal(t, licBefore.InUse, licAfter.InUse)
	assert.ElementsMatch(t, seatsBefore, seatsAfter)
}

func TestCanSwapWhenLicenseFull(t *testing.T) {
	//given
	store := spicedbSeatLicenseRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	err := fillUpLicense(lic)
	assert.NoError(t, err)
	spicedbContainer.WaitForQuantizationInterval()
	//when
	req := modifyLicRequestFromVars("okay", "o1", []string{"u7"}, []string{"u1"})
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
	assert.Contains(t, seats, domain.SubjectID("u7"))
	assert.NotContains(t, seats, domain.SubjectID("u1"))
}

func TestCantUnassignSeatThatWasNotAssigned(t *testing.T) {
	//given
	store := spicedbSeatLicenseRepository()
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
	assert.Equal(t, 2, license.InUse)
}

func fillUpLicense(lic *SeatLicenseService) error {
	toAssign := make([]string, 8) //8 free slots in seed data
	for i := range toAssign {
		toAssign[i] = "u" + strconv.Itoa(i+12) // +12 bc. u3 and u4 are disabled
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

	store := spicedbSeatLicenseRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	err := lic.ModifySeats(req)

	assert.ErrorIs(t, err, domain.ErrNotAuthorized)
}

func TestLicensingAssignUnassignRoundTrip(t *testing.T) {
	addReq := modifyLicRequestFromVars("okay",
		"o1",
		[]string{"u5"},
		[]string{})

	store := spicedbSeatLicenseRepository()
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
		[]string{"u5"})

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
