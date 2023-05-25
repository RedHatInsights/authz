package application

import (
	"authz/domain"
	"authz/domain/contracts"
	"authz/domain/services"
	"context"
)

// LicenseAppService the handler for seat related endpoints.
type LicenseAppService struct {
	accessRepo    *contracts.AccessRepository
	seatRepo      *contracts.SeatLicenseRepository
	principalRepo contracts.PrincipalRepository
	ctx           context.Context
}

// GetSeatAssignmentRequest represents a request to get the users assigned seats on a license
type GetSeatAssignmentRequest struct {
	Requestor    string
	OrgID        string
	ServiceID    string
	IncludeUsers bool
	Assigned     bool
}

// ModifySeatAssignmentRequest represents a request to assign and/or unassign seat licenses
type ModifySeatAssignmentRequest struct {
	Requestor string
	OrgID     string
	ServiceID string
	Assign    []string
	Unassign  []string
}

// GetSeatAssignmentCountsRequest represents a request to get the seats limit and current allocation for a license
type GetSeatAssignmentCountsRequest struct {
	Requestor string
	OrgID     string
	ServiceID string
}

// NewLicenseAppService ctor.
func NewLicenseAppService(accessRepo *contracts.AccessRepository, seatRepo *contracts.SeatLicenseRepository, principalRepo contracts.PrincipalRepository) *LicenseAppService {
	return &LicenseAppService{
		accessRepo:    accessRepo,
		seatRepo:      seatRepo,
		principalRepo: principalRepo,
		ctx:           context.Background(),
	}
}

// GetSeatAssignmentCounts gets the seat limit and current allocation for a license
func (s *LicenseAppService) GetSeatAssignmentCounts(req GetSeatAssignmentCountsRequest) (limit int, available int, err error) {
	evt := domain.GetLicenseEvent{
		OrgID:     req.OrgID,
		ServiceID: req.ServiceID,
	}

	evt.Requestor = domain.SubjectID(req.Requestor)

	seatsService := services.NewSeatLicenseService(*s.seatRepo, *s.accessRepo)

	lic, err := seatsService.GetLicense(evt)
	if err != nil {
		return 0, 0, err
	}

	limit = lic.MaxSeats
	available = lic.GetAvailableSeats()
	err = nil
	return
}

// GetSeatAssignments gets the subjects assigned to seats in a license
func (s *LicenseAppService) GetSeatAssignments(req GetSeatAssignmentRequest) ([]domain.Principal, error) {
	evt := domain.GetLicenseEvent{
		OrgID:     req.OrgID,
		ServiceID: req.ServiceID,
	}

	evt.Requestor = domain.SubjectID(req.Requestor)

	seatService := services.NewSeatLicenseService(*s.seatRepo, *s.accessRepo)

	var resultIds []domain.SubjectID
	var err error
	if req.Assigned {
		resultIds, err = seatService.GetAssignedSeats(evt)
	} else {
		resultIds, err = seatService.GetAssignableSeats(evt)
	}

	if err != nil {
		return nil, err
	}

	if req.IncludeUsers {
		return s.principalRepo.GetByIDs(resultIds)
	}

	principals := make([]domain.Principal, len(resultIds))
	for i, id := range resultIds {
		principals[i] = domain.Principal{ID: id}
	}
	return principals, nil
}

// ModifySeats Assign and/or unassign a number of users for a given org and service
func (s *LicenseAppService) ModifySeats(req ModifySeatAssignmentRequest) error {
	evt := domain.ModifySeatAssignmentEvent{
		Org:     domain.Organization{ID: req.OrgID},
		Service: domain.Service{ID: req.ServiceID},
	}

	evt.Requestor = domain.SubjectID(req.Requestor)

	evt.Assign = make([]domain.SubjectID, len(req.Assign))
	for i, id := range req.Assign {
		evt.Assign[i] = domain.SubjectID(id)
	}

	evt.UnAssign = make([]domain.SubjectID, len(req.Unassign))
	for i, id := range req.Unassign {
		evt.UnAssign[i] = domain.SubjectID(id)
	}

	seatService := services.NewSeatLicenseService(*s.seatRepo, *s.accessRepo)

	return seatService.ModifySeats(evt)
}

//func subtract(first []domain.SubjectID, second []domain.SubjectID) []domain.SubjectID { //Move to a SubjectSet or something?
//	subtrahend := map[domain.SubjectID]interface{}{} //idiomatic set
//	for _, id := range second {
//		subtrahend[id] = struct{}{}
//	}
//
//	result := make([]domain.SubjectID, 0, len(first))
//
//	for _, id := range first {
//		if _, ok := subtrahend[id]; ok {
//			continue
//		}
//
//		result = append(result, id)
//	}
//
//	return result
//}
