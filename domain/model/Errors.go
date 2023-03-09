package model

import "errors"

// ErrNotAuthorized is returned when the identity invoking the API does not have permission to invoke that operation.
var ErrNotAuthorized = errors.New("NotAuthorized")

// ErrNotAuthenticated is returned when anonymously invoking an endpoint that requires an identity
var ErrNotAuthenticated = errors.New("NotAuthenticated")
