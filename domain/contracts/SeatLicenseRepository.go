package contracts

import "authz/domain/model"

// SeatLicenseRepository is a contract that describes the required operations for accessing and manipulating per-seat license data
type SeatLicenseRepository interface {
	// AssignSeat assigns the given principal a seat for the given service
	AssignSeat(principal model.Principal, orgId string, svc model.Service) error
	// UnAssignSeat removes the seat assignment for the given principal for the given service
	UnAssignSeat(principal model.Principal, orgId string, svc model.Service) error
}

// TODO
// To show license information, we need:
// GetLicensedUsers(product/service) -> returns user representations for licensed seat
// GetLicenseInfo(service) -> returns curr & max
// GetUnlicensedUser(org org) -> returns all users of an org without a license
