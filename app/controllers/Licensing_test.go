package controllers

import (
	"authz/app"
	"authz/app/contracts"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLicensingAssignSeatsErrorsWhenNotAuthorized(t *testing.T) {
	req := modifyLicRequestFromVars("bad",
		"aspian",
		"aspian")

	store := mockAuthzStore()
	lic := NewLicensing(store, store)

	err := lic.AssignSeats(req)

	assert.ErrorIs(t, err, app.ErrNotAuthorized)
}

func TestLicensingUnAssignSeatsErrorsWhenNotAuthenticated(t *testing.T) {
	req := modifyLicRequestFromVars("",
		"aspian",
		"aspian")

	store := mockAuthzStore()
	lic := NewLicensing(store, store)

	err := lic.UnAssignSeats(req)

	assert.ErrorIs(t, err, app.ErrNotAuthenticated)
}

func TestLicensingUnAssignSeatsErrorsWhenNotAuthorized(t *testing.T) {
	req := modifyLicRequestFromVars("bad",
		"aspian",
		"aspian")

	store := mockAuthzStore()
	lic := NewLicensing(store, store)

	err := lic.UnAssignSeats(req)

	assert.ErrorIs(t, err, app.ErrNotAuthorized)
}

func TestLicensingAssignSeatsErrorsWhenMismatchedOrgs(t *testing.T) {
	req := modifyLicRequestFromVars("okay",
		"aspian",
		"bspian")

	store := mockAuthzStore()
	lic := NewLicensing(store, store)

	err := lic.AssignSeats(req)

	assert.ErrorIs(t, err, app.ErrInvalidRequest)
}

func TestLicensingUnAssignSeatsErrorsWhenSubjectAndRequestOrgsMismatched(t *testing.T) {
	req := modifyLicRequestFromVars("okay",
		"aspian",
		"bspian")

	store := mockAuthzStore()
	lic := NewLicensing(store, store)

	err := lic.UnAssignSeats(req)

	assert.ErrorIs(t, err, app.ErrInvalidRequest)
}

func TestLicensingAssignUnassignRoundTrip(t *testing.T) {
	req := modifyLicRequestFromVars("okay",
		"aspian",
		"aspian")

	store := mockAuthzStore()
	lic := NewLicensing(store, store)

	authz, err := store.CheckAccess(req.Principals[0], "use", req.Service.AsResource())
	assert.NoError(t, err)
	assert.False(t, authz, "Should not have been authorized without license.")

	err = lic.AssignSeats(req)
	assert.NoError(t, err)

	authz, err = store.CheckAccess(req.Principals[0], "use", req.Service.AsResource())
	assert.NoError(t, err)
	assert.True(t, authz, "Should have been authorized with license.")

	err = lic.UnAssignSeats(req)
	assert.NoError(t, err)

	authz, err = store.CheckAccess(req.Principals[0], "use", req.Service.AsResource())
	assert.NoError(t, err)
	assert.False(t, authz, "Should not have been authorized without license.")
}

func modifyLicRequestFromVars(requestorID string, requestorOrg string, subjectOrg string) contracts.ModifySeatAssignmentRequest {
	return contracts.ModifySeatAssignmentRequest{
		Request: contracts.Request{
			Requestor: app.NewPrincipal(requestorID, "requestorName", requestorOrg, true), //TODO: ok, severe sign that we need another struct i guess. we're mixing concerns :)
		},
		Org:        app.Organization{Id: "aspian"},
		Principals: []app.Principal{app.NewPrincipal("okay", "u1", subjectOrg, true)}, //TODO: see comment above.
		Service:    app.Service{Id: "wisdom"},
	}
}
