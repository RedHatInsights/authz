package handler

import (
	"authz/domain/contracts"
	"context"
)

// SeatHandler the handler for seat related endpoints.
type SeatHandler struct {
	accessRepo *contracts.AccessRepository
	ctx        context.Context
}

// NewSeatHandler ctor.
func (s *SeatHandler) NewSeatHandler(accessRepo *contracts.AccessRepository) *SeatHandler {
	return &SeatHandler{
		accessRepo: accessRepo,
		ctx:        context.Background(),
	}
}

// TODO
func (s *SeatHandler) AddSeats(addSeatsRequest string) error {
	return nil
}
