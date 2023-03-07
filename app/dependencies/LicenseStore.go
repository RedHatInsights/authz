package dependencies

import "authz/app"

type LicenseStore interface {
	AssignSeat(principal app.Principal, svc app.Service) error
	UnAssignSeat(principal app.Principal, svc app.Service) error
}
