// Package mock implements a stub accessRepository
package mock

import (
	"authz/domain/model"
	vo "authz/domain/valueobjects"
)

// StubAccessRepository represents an in-memory authorization system with a fixed state
type StubAccessRepository struct {
	//The internal authorization state. The keys are subject IDs, and the values are the results. The results are the same per subject regardless of operation and resource.
	Data          map[string]bool
	LicensedSeats map[string]map[string]model.License
}

// NewConnection Stub impl
func (s *StubAccessRepository) NewConnection(_ string, _ string, _ bool) {
	// NOT USED IN STUB
}

// CheckAccess returns true if the subject has been specified to have access, otherwise false.
func (s *StubAccessRepository) CheckAccess(principal model.Principal, operation string, resource model.Resource) (vo.AccessDecision, error) {
	if authz, ok := s.Data[principal.ID]; ok {
		if authz && operation == "use" {
			if lic, ok := s.LicensedSeats[principal.OrgID][resource.ID]; ok {
				return vo.AccessDecision((&lic).IsAssigned(principal.ID)), nil //Authorized and license present, return license status
			}
			return vo.AccessDecision(false), nil //Authorized but no license, so return false
		}
		return vo.AccessDecision(authz), nil //No licensing required, passthrough authz status
	}

	return vo.AccessDecision(false), nil //Unknown principal, implicitly not authorized
}

// GetLicense retrieves the stored license for the given organization and service, if any.
func (s *StubAccessRepository) GetLicense(orgID string, serviceID string) (*model.License, error) {
	if lic, ok := s.LicensedSeats[orgID][serviceID]; ok {
		return &lic, nil
	}
	return nil, nil
}

// UpdateLicense saves updated license state
func (s *StubAccessRepository) UpdateLicense(lic *model.License) error {
	if lics, ok := s.LicensedSeats[lic.OrgID]; ok {
		lics[lic.ServiceID] = *lic
	} else {
		s.LicensedSeats[lic.OrgID] = map[string]model.License{lic.ServiceID: *lic}
	}

	return nil
}
