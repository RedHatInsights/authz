package application

import (
	"authz/domain"
	"authz/domain/contracts"
	"authz/domain/services"
	"context"
	"fmt"

	"github.com/golang/glog"
)

// LicenseAppService the handler for seat related endpoints.
type LicenseAppService struct {
	accessRepo    contracts.AccessRepository
	seatRepo      contracts.SeatLicenseRepository
	principalRepo contracts.PrincipalRepository
	subjectRepo   contracts.SubjectRepository
	orgRepo       contracts.OrganizationRepository
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

// OrgEntitledEvent represents an event where an organization has been entitled with a new license
type OrgEntitledEvent struct {
	OrgID     string
	ServiceID string
	MaxSeats  int
}

// NewLicenseAppService ctor.
func NewLicenseAppService(accessRepo contracts.AccessRepository, seatRepo contracts.SeatLicenseRepository, principalRepo contracts.PrincipalRepository, subjectRepo contracts.SubjectRepository, orgRepo contracts.OrganizationRepository) *LicenseAppService {
	return &LicenseAppService{
		accessRepo:    accessRepo,
		seatRepo:      seatRepo,
		principalRepo: principalRepo,
		subjectRepo:   subjectRepo,
		orgRepo:       orgRepo,
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

	seatsService := services.NewSeatLicenseService(s.seatRepo, s.accessRepo)

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

	seatService := services.NewSeatLicenseService(s.seatRepo, s.accessRepo)

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

	seatService := services.NewSeatLicenseService(s.seatRepo, s.accessRepo)

	return seatService.ModifySeats(evt)
}

// ProcessOrgEntitledEvent handles the OrgEntitledEvent by storing the license and importing users
func (s *LicenseAppService) ProcessOrgEntitledEvent(evt OrgEntitledEvent, strictMode bool) error {
	// first, check if there's already a license/org and/or existing org users.
	licenseImported, usersImported, err := s.seatRepo.IsImported(evt.OrgID, evt.ServiceID)
	if err != nil {
		return err
	}

	if licenseImported {
		licExistsErr := fmt.Errorf("License already exists for the given org in %v: ", evt)

		if strictMode {
			return licExistsErr
		}
		glog.Warning(licExistsErr)

	} else {
		// we only create/touch an existing license for a given org if it doesn't already exist
		err = s.seatRepo.ApplyLicense(&domain.License{
			OrgID:     evt.OrgID,
			ServiceID: evt.ServiceID,
			MaxSeats:  evt.MaxSeats,
			Version:   "",
			InUse:     0,
		})

		if err != nil {
			return err
		}

		glog.Infof("License for service %s with %v seats added.", evt.ServiceID, evt.MaxSeats)
	}

	// whether an existing license exists or not, we can try to figure out if users have already been imported, and if not add them
	// this is safe to re-run
	if usersImported {
		glog.Infof("Skipping user import. Org already exists.", evt.ServiceID, evt.MaxSeats)
		return nil
	}

	subjects, errors := s.subjectRepo.GetByOrgID(evt.OrgID)

loop:
	for {
		select {
		case subject, ok := <-subjects:
			if ok {
				err = s.orgRepo.AddSubject(evt.OrgID, subject)
				if err != nil {
					glog.Errorf("Failed to import user %s to org %s", subject.SubjectID, evt.OrgID)

					if errorShouldBeRetried(err) { // TODO: add test to test 'true' path
						return err // TODO: add retry mechanism (but for now it's fine to bomb out and retry the whole processing of the event since all ops are idempotent)
					}
				}
			} else {
				break loop
			}
		case err, ok := <-errors:
			if !ok {
				break loop
			}

			glog.Errorf(err.Error()) // TODO: think more about the contract. Is it possible to reason about individual errors in a channel and what they refer to and whether we should stop or continue?

			return err
		}
	}

	return nil
}

func errorShouldBeRetried(err error) bool {
	return err != domain.ErrSubjectAlreadyExists
}
