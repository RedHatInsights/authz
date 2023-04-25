package domain

import (
	"errors"
	"fmt"
)

// ErrNotAuthorized is returned when the identity invoking the API does not have permission to invoke that operation.
var ErrNotAuthorized = errors.New("NotAuthorized")

// ErrNotAuthenticated is returned when anonymously invoking an endpoint that requires an identity
var ErrNotAuthenticated = errors.New("NotAuthenticated")

// ErrInvalidRequest is returned when some part of the request is incompatible with another part.
var ErrInvalidRequest = errors.New("InvalidRequest")

// ErrLicenseLimitExceeded is returned when an operation attempts to allocate more licenses than are available
type ErrLicenseLimitExceeded struct {
	// MaxSeats is the total number of seats permitted by the license
	MaxSeats int
	// AvailableSeats represents the number of seats currently free
	AvailableSeats int
}

// Error formats a human-readable message from the data contained in the ErrLicenseLimitExceeded struct
func (e ErrLicenseLimitExceeded) Error() string {
	return fmt.Sprintf("License limit would have been exceeded. %d of %d available.", e.AvailableSeats, e.MaxSeats)
}

// NewErrLicenseLimitExceeded constructs a new ErrLicenseLimitExceeded error
func NewErrLicenseLimitExceeded(max int, available int) ErrLicenseLimitExceeded {
	return ErrLicenseLimitExceeded{MaxSeats: max, AvailableSeats: available}
}
