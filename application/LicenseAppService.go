package application

import (
	"authz/domain/contracts"
	"context"
)

// LicenseAppService the handler for seat related endpoints.
type LicenseAppService struct {
	accessRepo *contracts.AccessRepository
	ctx        context.Context
}

// NewLicenseAppService ctor.
func (s *LicenseAppService) NewLicenseAppService(accessRepo *contracts.AccessRepository) *LicenseAppService {
	return &LicenseAppService{
		accessRepo: accessRepo,
		ctx:        context.Background(),
	}
}

// ModifySeats TODO
func (s *LicenseAppService) ModifySeats(_ string) error {
	return nil
}
