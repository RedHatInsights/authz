package services

import (
	"authz/domain/contracts"
	"authz/domain/model"
)

type SeatLicenseService struct {
	seats contracts.SeatLicenseRepository
	authz contracts.AccessRepository
}

func (l *SeatLicenseService) ModifySeats(evt model.ModifySeatAssignmentEvent) error {
	if err := l.ensureRequestorIsAuthorizedToManageLicenses(evt.Requestor, evt.Org); err != nil {
		return err
	}

	if !evt.IsValid() {
		return model.ErrInvalidRequest
	}

	for _, principal := range evt.UnAssign {
		l.seats.UnAssignSeat(principal, evt.Service)
	}

	for _, principal := range evt.Assign {
		l.seats.AssignSeat(principal, evt.Service)
	}

	return nil
}

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
