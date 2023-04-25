package domain

import "errors"

// ErrNotAuthorized is returned when the identity invoking the API does not have permission to invoke that operation.
var ErrNotAuthorized = errors.New("NotAuthorized")

// ErrNotAuthenticated is returned when anonymously invoking an endpoint that requires an identity
var ErrNotAuthenticated = errors.New("NotAuthenticated")

// ErrInvalidRequest is returned when some part of the request is incompatible with another part.
var ErrInvalidRequest = errors.New("InvalidRequest")

// ErrLicenseLimitExceeded is returned when an operation attempts to allocate more licenses than are available
var ErrLicenseLimitExceeded = errors.New("LicenseLimitExceeded")
