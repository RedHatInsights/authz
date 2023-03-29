package application

import (
	"authz/domain/contracts"
	"authz/domain/model"
	"authz/domain/services"
	vo "authz/domain/valueobjects"
	"context"
)

// LicenseAppService the handler for seat related endpoints.
type LicenseAppService struct {
	accessRepo    contracts.AccessRepository
	seatRepo      contracts.SeatLicenseRepository
	principalRepo contracts.PrincipalRepository
	ctx           context.Context
}

// ModifySeatAssignmentRequest represents a request to assign and/or unassign seat licenses
type ModifySeatAssignmentRequest struct {
	Requestor string
	OrgID     string
	ServiceID string
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
		Org:     model.Organization{ID: req.OrgID},
		Service: model.Service{ID: req.ServiceID},
	}

	evt.Requestor = vo.SubjectID(req.Requestor)

	evt.Assign = make([]vo.SubjectID, len(req.Assign))
	for i, id := range req.Assign {
		evt.Assign[i] = vo.SubjectID(id)
	}

	evt.UnAssign = make([]vo.SubjectID, len(req.Unassign))
	for i, id := range req.Unassign {
		evt.UnAssign[i] = vo.SubjectID(id)
	}

	seatService := services.NewSeatLicenseService(s.seatRepo, s.accessRepo)

	return seatService.ModifySeats(evt)
}
