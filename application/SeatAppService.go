package application

import (
	"authz/domain/contracts"
	"context"
)

// SeatAppService the handler for seat related endpoints.
type SeatAppService struct {
	accessRepo *contracts.AccessRepository
	ctx        context.Context
}

// NewSeatAppService ctor.
func (s *SeatAppService) NewSeatAppService(accessRepo *contracts.AccessRepository) *SeatAppService {
	return &SeatAppService{
		accessRepo: accessRepo,
		ctx:        context.Background(),
	}
}

// AddSeats TODO
func (s *SeatAppService) AddSeats(addSeatsRequest string) error {
	return nil
}
