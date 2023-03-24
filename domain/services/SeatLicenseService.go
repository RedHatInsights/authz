package services

import (
	"authz/domain/contracts"
	"authz/domain/model"
)

// SeatLicenseService performs operations related to per-seat licensing
type SeatLicenseService struct {
	seats contracts.SeatLicenseRepository
	authz contracts.AccessRepository
}

// ModifySeats handles ModifySeatAssignmentEvents to assign and unassign seats
func (l *SeatLicenseService) ModifySeats(evt model.ModifySeatAssignmentEvent) error {
	if err := l.ensureRequestorIsAuthorizedToManageLicenses(evt.Requestor, evt.Org); err != nil {
		return err
	}

	if !evt.IsValid() {
		return model.ErrInvalidRequest
	}

	lic, err := l.seats.GetLicense(evt.Org.ID, evt.Service.ID)
	if err != nil {
		return err
	}

	for _, principal := range evt.UnAssign {
		lic.UnAssign(principal.ID)
	}

	for _, principal := range evt.Assign {
		lic.Assign(principal.ID)
	}

	err = l.seats.UpdateLicense(lic)
	if err != nil {
		return err
	}

	return nil
}

// NewSeatLicenseService constructs a new SeatLicenseService
func NewSeatLicenseService(seats contracts.SeatLicenseRepository, authz contracts.AccessRepository) *SeatLicenseService {
	return &SeatLicenseService{seats: seats, authz: authz}
}

func (l *SeatLicenseService) ensureRequestorIsAuthorizedToManageLicenses(requestor model.Principal, org model.Organization) error {
	if requestor.IsAnonymous() {
		return model.ErrNotAuthenticated
	}

	authz, err := l.authz.CheckAccess(requestor, "manage_license", org.AsResource()) //Maybe on a per-service basis?
	if err != nil {
		return err
	}

	if !authz {
		return model.ErrNotAuthorized
	}

	return nil
}
