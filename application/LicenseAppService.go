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

// ModifySeatAssignmentRequest represents a request to assign and/or unassign seat licenses
type ModifySeatAssignmentRequest struct {
	Requestor string
	OrgID     string
	ServiceID string
	Assign    []string
	Unassign  []string
}

type GetSeatAssignmentRequest struct {
	Requestor    string
	OrgID        string
	ServiceID    string
	IncludeUsers bool
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

func (s *LicenseAppService) GetSeatAssignmentCounts(orgID string, serviceID string) (current uint, max uint, err error) {
	lic, err := s.seatRepo.GetLicense(orgID, serviceID)
	if err != nil {
		return 0, 0, err
	}

	current = lic.InUse()
	max = lic.MaxSeats
	err = nil
	return
}

func (s *LicenseAppService) GetAssignedSeats(req GetSeatAssignmentRequest) ([]model.Principal, error) {
	lic, err := s.seatRepo.GetLicense(req.OrgID, req.ServiceID)
	if err != nil {
		return nil, err
	}

	ids := lic.GetAssigned()
	principals, err := s.principalRepo.GetByIDs(ids)
	if err != nil {
		return nil, err
	}

	return principals, nil
}

func (s *LicenseAppService) ModifySeats(req ModifySeatAssignmentRequest) error {
	evt := model.ModifySeatAssignmentEvent{
		Org:     model.Organization{ID: req.OrgID},
		Service: model.Service{ID: req.ServiceID},
	}

	var err error
	evt.Requestor, err = s.principalRepo.GetByID(req.Requestor)

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
