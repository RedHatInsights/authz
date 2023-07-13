package domain

import "errors"

// ErrNotAuthorized is returned when the identity invoking the API does not have permission to invoke that operation.
var ErrNotAuthorized = errors.New("NotAuthorized")

// ErrNotAuthenticated is returned when anonymously invoking an endpoint that requires an identity
var ErrNotAuthenticated = errors.New("NotAuthenticated")

// ErrInvalidRequest is returned when some part of the request is incompatible with another part.
type ErrInvalidRequest struct {
	error
	Reason string
}

// NewErrInvalidRequest creates a new InvalidRequest error for the given reason
func NewErrInvalidRequest(reason string) ErrInvalidRequest {
	return ErrInvalidRequest{
		error:  errors.New("InvalidRequest"),
		Reason: reason,
	}
}

// ErrLicenseLimitExceeded is returned when an operation attempts to allocate more licenses than are available
var ErrLicenseLimitExceeded = errors.New("LicenseLimitExceeded")

// ErrConflict is returned when a request cannot be processed due to an apparent conflicting request (ex: concurrency)
var ErrConflict = errors.New("Conflict")

// ErrSubjectAlreadyExists is returned whenever we try to add a subject in OrganizationRepository that already exists
var ErrSubjectAlreadyExists = errors.New("ErrSubjectAlreadyExists")
