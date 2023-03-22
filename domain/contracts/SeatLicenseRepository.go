package contracts

import "authz/domain/model"

type SeatLicenseRepository interface {
	AssignSeat(principal model.Principal, svc model.Service) error
	UnAssignSeat(principal model.Principal, svc model.Service) error
}

// TODO
// To show license information, we need:
// GetLicensedUsers(product/service) -> returns user representations for licensed seat
// GetLicenseInfo(service) -> returns curr & max
// GetUnlicensedUser(org org) -> returns all users of an org without a license
