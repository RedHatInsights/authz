package controllers

import (
	"authz/app"
	"authz/app/contracts"
	"authz/app/dependencies"
)

type Licensing struct {
	licenseStore dependencies.LicenseStore
	authzStore   dependencies.AuthzStore
}

func (l Licensing) AssignSeats(req contracts.ModifySeatAssignmentRequest) error {
	if err := l.ensureRequestorIsAuthorizedToManageLicenses(req.Requestor, req.Org); err != nil {
		return err
	}

	if !req.IsValid() {
		return app.ErrInvalidRequest
	}

	for _, principal := range req.Principals {
		l.licenseStore.AssignSeat(principal, req.Service)
	}

	return nil
}

func (l Licensing) UnAssignSeats(req contracts.ModifySeatAssignmentRequest) error {
	if err := l.ensureRequestorIsAuthorizedToManageLicenses(req.Requestor, req.Org); err != nil {
		return err
	}

	if !req.IsValid() {
		return app.ErrInvalidRequest
	}

	for _, principal := range req.Principals {
		l.licenseStore.UnAssignSeat(principal, req.Service)
	}

	return nil
}

func (l Licensing) GetLicensedSeats(req contracts.GetSeatsRequest) ([]app.Principal, error) {
	if err := l.ensureRequestorIsAuthorizedToReadLicenses(req.Requestor, req.Org); err != nil {
		return nil, err
	}

	return []app.Principal{}, nil // TODO
}

func (l Licensing) GetUnlicensedSeats(req contracts.GetSeatsRequest) ([]app.Principal, error) {
	if err := l.ensureRequestorIsAuthorizedToReadLicenses(req.Requestor, req.Org); err != nil {
		return nil, err
	}

	return []app.Principal{}, nil // TODO
}

func (l Licensing) GetLicenseInformation(req contracts.GetSeatsRequest) (app.LicenseInformation, error) {
	err := l.ensureRequestorIsAuthorizedToReadLicenses(req.Requestor, req.Org)
	if err != nil {
		return app.LicenseInformation{}, err
	}

	return app.LicenseInformation{}, nil
}

func NewLicensing(licenseStore dependencies.LicenseStore, authz dependencies.AuthzStore) Licensing {
	return Licensing{licenseStore: licenseStore, authzStore: authz}
}

func (l Licensing) ensureRequestorIsAuthorizedToManageLicenses(requestor app.Principal, org app.Organization) error {
	if requestor.IsAnonymous() {
		return app.ErrNotAuthenticated
	}

	authz, err := l.authzStore.CheckAccess(requestor, "manage_license", org.AsResource()) //Maybe on a per-service basis?
	if err != nil {
		return err
	}

	if !authz {
		return app.ErrNotAuthorized
	}

	return nil
}

func (l Licensing) ensureRequestorIsAuthorizedToReadLicenses(requestor app.Principal, org app.Organization) error {
	if requestor.IsAnonymous() {
		return app.ErrNotAuthenticated
	}

	authz, err := l.authzStore.CheckAccess(requestor, "view_license", org.AsResource()) //Maybe on a per-service basis?
	if err != nil {
		return err
	}

	if !authz {
		return app.ErrNotAuthorized
	}

	return nil
}
