package services

import (
	"authz/domain/contracts"
	"authz/domain/model"
	vo "authz/domain/valueobjects"
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

	assert.ErrorIs(t, err, model.ErrNotAuthenticated)
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

	assert.ErrorIs(t, err, model.ErrNotAuthorized)
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

func modifyLicRequestFromVars(requestorID string, subjectOrg string, assign []string, unassign []string) model.ModifySeatAssignmentEvent {
	evt := model.ModifySeatAssignmentEvent{
		Request: model.Request{
			Requestor: vo.SubjectID(requestorID),
		},
		Org:     model.Organization{ID: subjectOrg},
		Service: model.Service{ID: "smarts"},
	}

	evt.Assign = make([]vo.SubjectID, len(assign))
	for i, id := range assign {
		evt.Assign[i] = vo.SubjectID(id)
	}

	evt.UnAssign = make([]vo.SubjectID, len(unassign))
	for i, id := range unassign {
		evt.UnAssign[i] = vo.SubjectID(id)
	}

	return evt
}
