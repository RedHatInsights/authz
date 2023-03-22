package application

import (
	"authz/domain/contracts"
	"authz/domain/model"
	"authz/domain/services"
	"context"
)

// LicenseAppService the handler for seat related endpoints.
type LicenseAppService struct {
	accessRepo    contracts.AccessRepository
	seatRepo      contracts.SeatLicenseRepository
	principalRepo contracts.PrincipalRepository
	ctx           context.Context
}

type ModifySeatAssignmentRequest struct {
	Requestor model.Principal
	OrgId     string
	ServiceId string
	Assign    []string
	Unassign  []string
}

// NewLicenseAppService ctor.
func NewLicenseAppService(accessRepo contracts.AccessRepository, seatRepo contracts.SeatLicenseRepository, principalRepo contracts.PrincipalRepository) *LicenseAppService {
	return &LicenseAppService{
		accessRepo:    accessRepo,
		seatRepo:      seatRepo,
		principalRepo: principalRepo,
		ctx:           context.Background(),
	}
}

// ModifySeats TODO
func (s *LicenseAppService) ModifySeats(req ModifySeatAssignmentRequest) error {
	evt := model.ModifySeatAssignmentEvent{
		Request: model.Request{
			Requestor: req.Requestor,
		},
		Org:     model.Organization{Id: req.OrgId},
		Service: model.Service{Id: req.ServiceId},
	}

	var err error
	evt.Assign, err = s.principalRepo.GetByIDs(req.Assign)
	if err != nil {
		return err
	}

	evt.UnAssign, err = s.principalRepo.GetByIDs(req.Unassign)
	if err != nil {
		return err
	}

	seatService := services.NewSeatLicenseService(s.seatRepo, s.accessRepo)

	return seatService.ModifySeats(evt)
}
