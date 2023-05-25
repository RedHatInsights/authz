package services

import (
	"authz/domain"
	"authz/domain/contracts"
)

// SeatLicenseService performs operations related to per-seat licensing
type SeatLicenseService struct {
	seats contracts.SeatLicenseRepository
	authz contracts.AccessRepository
}

// ModifySeats handles ModifySeatAssignmentEvents to assign and unassign seats
func (l *SeatLicenseService) ModifySeats(evt domain.ModifySeatAssignmentEvent) error {
	if err := l.ensureRequestorIsAuthorizedToManageLicenses(evt.Requestor); err != nil {
		return err
	}

	license, err := l.seats.GetLicense(evt.Org.ID, evt.Service.ID)
	if err != nil {
		return err
	}

	if license.GetAvailableSeats() < (len(evt.Assign) - len(evt.UnAssign)) {
		return domain.ErrLicenseLimitExceeded
	}

	return l.seats.ModifySeats(evt.Assign, evt.UnAssign, license, evt.Org.ID, evt.Service)
}

// GetLicense gets the License for the provided information
func (l *SeatLicenseService) GetLicense(evt domain.GetLicenseEvent) (*domain.License, error) {
	if err := l.ensureRequestorIsAuthorizedToManageLicenses(evt.Requestor); err != nil {
		return nil, err
	}

	return l.seats.GetLicense(evt.OrgID, evt.ServiceID)
}

func (l *SeatLicenseService) GetAssignableSeats(evt domain.GetLicenseEvent) ([]domain.SubjectID, error) {
	if err := l.ensureRequestorIsAuthorizedToManageLicenses(evt.Requestor); err != nil {
		return nil, err
	}

	// TODO: return new GetAssignable in SeatRepositoryService contract
	return nil, nil
}

// GetAssignedSeats gets the subjects assigned to the given license
func (l *SeatLicenseService) GetAssignedSeats(evt domain.GetLicenseEvent) ([]domain.SubjectID, error) {
	if err := l.ensureRequestorIsAuthorizedToManageLicenses(evt.Requestor); err != nil {
		return nil, err
	}

	return l.seats.GetAssigned(evt.OrgID, evt.ServiceID)
}

// NewSeatLicenseService constructs a new SeatLicenseService
func NewSeatLicenseService(seats contracts.SeatLicenseRepository, authz contracts.AccessRepository) *SeatLicenseService {
	return &SeatLicenseService{seats: seats, authz: authz}
}

func (l *SeatLicenseService) ensureRequestorIsAuthorizedToManageLicenses(requestor domain.SubjectID) error {
	if !requestor.HasIdentity() {
		return domain.ErrNotAuthenticated
	}

	//TODO: implement meta-authz
	authz, err := true, error(nil) //l.authz.CheckAccess(requestor, "manage_license", org.AsResource()) //Maybe on a per-service basis?
	if err != nil {
		return err
	}

	if !authz {
		return domain.ErrNotAuthorized
	}

	return nil
}
