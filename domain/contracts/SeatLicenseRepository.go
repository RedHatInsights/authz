package contracts

import (
	"authz/domain"
)

// SeatLicenseRepository is a contract that describes the required operations for accessing and manipulating per-seat license data
type SeatLicenseRepository interface {
	// AssignSeats assigns a given range of principals for a license. When there are unknown values in the range, it fails.
	AssignSeats(subjectIDs []domain.SubjectID, license *domain.License, orgID string, svc domain.Service) error
	// UnAssignSeats atomically removes a given range of principals for a license. When there are unknown values in the range, it fails.
	UnAssignSeats(subjectIDs []domain.SubjectID, license *domain.License, orgID string, svc domain.Service) error
	// GetLicense retrieves the stored license for the given organization and service, if any.
	GetLicense(orgID string, serviceID string) (*domain.License, error)
	// GetAssigned retrieves the IDs of the subjects assigned seats in the current license
	GetAssigned(orgID string, serviceID string) ([]domain.SubjectID, error)
}

// TODO
// To show license information, we need:
// GetLicensedUsers(product/service) -> returns user representations for licensed seat
// GetLicenseInfo(service) -> returns curr & max
// GetUnlicensedUser(org org) -> returns all users of an org without a license
