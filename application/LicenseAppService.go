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

// GetSeatAssignmentRequest represents a request to get the users assigned seats on a license
type GetSeatAssignmentRequest struct {
	Requestor    string
	OrgID        string
	ServiceID    string
	IncludeUsers bool
}

// ModifySeatAssignmentRequest represents a request to assign and/or unassign seat licenses
type ModifySeatAssignmentRequest struct {
	Requestor string
	OrgID     string
	ServiceID string
	Assign    []string
	Unassign  []string
}

type GetSeatAssignmentCountsRequest struct {
	Requestor string
	OrgID     string
	ServiceID string
}

type GetSeatDetailsRequest struct {
	Requestor         string
	OrgID             string
	ServiceID         string
	includeUserDetail bool
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

func (s *LicenseAppService) GetSeatAssignmentCounts(req GetSeatAssignmentCountsRequest) (current int, max int, err error) {
	evt := model.GetLicenseEvent{
		OrgID:     req.OrgID,
		ServiceID: req.ServiceID,
	}

	evt.Requestor = vo.SubjectID(req.Requestor)

	seatsService := services.NewSeatLicenseService(s.seatRepo, s.accessRepo)

	lic, err := seatsService.GetLicense(evt)
	if err != nil {
		return 0, 0, err
	}

	current = lic.InUse
	max = lic.MaxSeats
	err = nil
	return
}

func (s *LicenseAppService) GetSeatAssignments(req GetSeatAssignmentRequest) ([]model.Principal, error) {
	evt := model.GetLicenseEvent{
		OrgID:     req.OrgID,
		ServiceID: req.ServiceID,
	}

	evt.Requestor = vo.SubjectID(req.Requestor)

	seatService := services.NewSeatLicenseService(s.seatRepo, s.accessRepo)

	assigned, err := seatService.GetAssignedSeats(evt)
	if err != nil {
		return nil, err
	}

	if req.IncludeUsers {
		return s.principalRepo.GetByIDs(assigned)
	} else {
		principals := make([]model.Principal, len(assigned))
		for i, id := range assigned {
			principals[i] = model.NewPrincipal(id)
		}
		return principals, nil
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
