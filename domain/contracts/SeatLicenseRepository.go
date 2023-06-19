package contracts

import (
	"authz/domain"
)

// SeatLicenseRepository is a contract that describes the required operations for accessing and manipulating per-seat license data
type SeatLicenseRepository interface {
	// ModifySeats atomically persists changes to seat assignments for a license
	ModifySeats(assignedSubjectIDs []domain.SubjectID, removedSubjectIDs []domain.SubjectID, license *domain.License, orgID string, svc domain.Service) error
	// GetLicense retrieves the stored license for the given organization and service, if any.
	GetLicense(orgID string, serviceID string) (*domain.License, error)
	// GetAssignable retrieves the IDs of the subjects who are assignable, but not already assigned, to seats in the current license
	GetAssignable(orgID string, serviceID string) ([]domain.SubjectID, error)
	// GetAssigned retrieves the IDs of the subjects assigned seats in the current license
	GetAssigned(orgID string, serviceID string) ([]domain.SubjectID, error)
	// ApplyLicense stores the given license associated with its service and organization
	ApplyLicense(license *domain.License) error
	// IsImported returns true if a given Org has at least one license applied or exists and has at least one member in the schema.
	IsImported(orgID string, serviceID string) (bool, bool, error)
}
