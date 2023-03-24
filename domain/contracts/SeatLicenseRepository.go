package contracts

import "authz/domain/model"

// SeatLicenseRepository is a contract that describes the required operations for accessing and manipulating per-seat license data
type SeatLicenseRepository interface {
	// GetLicense retrieves the stored license for the given organization and service, if any.
	GetLicense(orgID string, serviceID string) (*model.License, error)
	// UpdateLicense saves updated license state
	UpdateLicense(lic *model.License) error
}

// TODO
// To show license information, we need:
// GetLicensedUsers(product/service) -> returns user representations for licensed seat
// GetLicenseInfo(service) -> returns curr & max
// GetUnlicensedUser(org org) -> returns all users of an org without a license
