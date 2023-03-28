package services

import (
	"authz/domain/contracts"
	"authz/domain/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicensingModifySeatsErrorsWhenNotAuthenticated(t *testing.T) {
	req := modifyLicRequestFromVars("",
		"aspian",
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
		"aspian",
		[]string{"okay"},
		[]string{})

	store := mockAuthzRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	err := lic.ModifySeats(req)

	assert.ErrorIs(t, err, model.ErrNotAuthorized)
}

func TestLicensingUnAssignSeatsErrorsWhenSubjectAndRequestOrgsMismatched(t *testing.T) {
	req := modifyLicRequestFromVars("okay",
		"aspian",
		"bspian",
		[]string{"okay"},
		[]string{})

	store := mockAuthzRepository()
	lic := NewSeatLicenseService(store.(contracts.SeatLicenseRepository), store)

	err := lic.ModifySeats(req)

	assert.ErrorIs(t, err, model.ErrInvalidRequest)
}

func TestLicensingAssignUnassignRoundTrip(t *testing.T) {
	addReq := modifyLicRequestFromVars("okay",
		"aspian",
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
		"aspian",
		[]string{},
		[]string{"okay"})

	err = lic.ModifySeats(remReq)
	assert.NoError(t, err)

	authz, err = store.CheckAccess(remReq.UnAssign[0], "use", remReq.Service.AsResource())
	assert.NoError(t, err)
	assert.False(t, bool(authz), "Should not have been authorized without license.")
}

func modifyLicRequestFromVars(requestorID string, requestorOrg string, subjectOrg string, assign []string, unassign []string) model.ModifySeatAssignmentEvent {
	evt := model.ModifySeatAssignmentEvent{
		Request: model.Request{
			Requestor: model.NewPrincipal(requestorID, requestorOrg),
		},
		Org:     model.Organization{ID: subjectOrg},
		Service: model.Service{ID: "smarts"},
	}

	evt.Assign = make([]model.Principal, len(assign))
	for i, id := range assign {
		evt.Assign[i] = model.NewPrincipal(id, requestorOrg)
	}

	evt.UnAssign = make([]model.Principal, len(unassign))
	for i, id := range unassign {
		evt.UnAssign[i] = model.NewPrincipal(id, requestorOrg)
	}

	return evt
}
