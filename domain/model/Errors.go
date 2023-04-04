package model

import (
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/pkg/errors"
)

const (
	// ErrorCodePrefix Prefix error code
	ErrorCodePrefix = "CIAM-Authz"
	// ErrorForbidden Forbidden occurs when a user is not allowed to access the service
	ErrorForbidden ServiceErrorCode = 4
	// ErrorForbiddenReason Forbidden occurs when a user is not allowed to access the service
	ErrorForbiddenReason string = "Forbidden to perform this action"

	// ErrorGeneral general error
	ErrorGeneral ServiceErrorCode = 9
	// ErrorGeneralReason general error reason
	ErrorGeneralReason string = "Unspecified error"

	// ErrorUnauthenticated unauthenticated error code
	ErrorUnauthenticated ServiceErrorCode = 15
	// ErrorUnauthenticatedReason unauthenticated error reason
	ErrorUnauthenticatedReason string = "Account authentication could not be verified"

	// ErrorUnauthorized unauthorized error code
	ErrorUnauthorized ServiceErrorCode = 11

	// ErrorUnauthorizedReason unauthorized error reason
	ErrorUnauthorizedReason string = "Account is unauthorized to perform this action"
	// ErrorNotImplemented error not implemented
	ErrorNotImplemented ServiceErrorCode = 10
	// ErrorNotImplementedReason error not implemented
	ErrorNotImplementedReason string = "HTTP Method not implemented for this endpoint"
	// ErrorConflict An entity with the specified unique values already exists
	ErrorConflict ServiceErrorCode = 6
	// ErrorConflictReason An entity with the specified unique values already exists
	ErrorConflictReason string = "An entity with the specified unique values already exists"

	// ErrorNotFound Resource not found
	ErrorNotFound ServiceErrorCode = 7
	// ErrorNotFoundReason Resource not found
	ErrorNotFoundReason string = "Resource not found"
	// ErrorBadRequest bad request
	ErrorBadRequest ServiceErrorCode = 21
	// ErrorBadRequestReason bad request
	ErrorBadRequestReason string = "Bad request"
	// ErrorValidation validation failed
	ErrorValidation ServiceErrorCode = 8
	// ErrorValidationReason validation failed
	ErrorValidationReason string = "General validation failure"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// ServiceError model
type ServiceError struct {
	// Code is the numeric and distinct ID for the error
	Code ServiceErrorCode
	// Reason is the context-specific reason the error was generated
	Reason string
	// HTTPCode is the HTTPCode associated with the error when the error is returned as an API response
	HTTPCode int
	// The original error that is causing the ServiceError, can be used for inspection
	cause error
}

// Errors - service errors with code and reason
func Errors() ServiceErrors {
	return ServiceErrors{
		ServiceError{ErrorForbidden, ErrorForbiddenReason, http.StatusForbidden, nil},
		ServiceError{ErrorConflict, ErrorConflictReason, http.StatusConflict, nil},
		ServiceError{ErrorNotFound, ErrorNotFoundReason, http.StatusNotFound, nil},
		ServiceError{ErrorValidation, ErrorValidationReason, http.StatusBadRequest, nil},
		ServiceError{ErrorGeneral, ErrorGeneralReason, http.StatusInternalServerError, nil},
		ServiceError{ErrorNotImplemented, ErrorNotImplementedReason, http.StatusMethodNotAllowed, nil},
		ServiceError{ErrorUnauthorized, ErrorUnauthorizedReason, http.StatusForbidden, nil},
		ServiceError{ErrorUnauthenticated, ErrorUnauthenticatedReason, http.StatusUnauthorized, nil},
		ServiceError{ErrorBadRequest, ErrorBadRequestReason, http.StatusBadRequest, nil},
	}
}

// Find error with code
func Find(code ServiceErrorCode) (bool, *ServiceError) {
	for _, err := range Errors() {
		if err.Code == code {
			return true, &err
		}
	}
	return false, nil
}

// New initialize error with code and reason
func New(code ServiceErrorCode, reason string, values ...interface{}) *ServiceError {
	return NewWithCause(code, nil, reason, values...)
}

// NewWithCause initialize error with cause
func NewWithCause(code ServiceErrorCode, cause error, reason string, values ...interface{}) *ServiceError {
	// If the code isn't defined, use the general error code
	var err *ServiceError
	exists, err := Find(code)
	if !exists {
		glog.Errorf("Undefined error code used: %d", code)
		err = &ServiceError{ErrorGeneral, "unspecified error", http.StatusInternalServerError, nil}
	}

	// TODO - if cause is nil, should we use the reason as the cause?
	if cause != nil {
		_, ok := cause.(stackTracer)
		if !ok {
			cause = errors.WithStack(cause) // add stacktrace info
		}
	}
	err.cause = cause

	// If the reason is unspecified, use the default
	if reason != "" {
		err.Reason = fmt.Sprintf(reason, values...)
	}

	return err
}

// ErrorList type
type ErrorList []error

// AddErrors adds the provided list of errors to the ErrorList.
// If the provided list of errors contain error elements that are of type
// ErrorList those are recursively "unrolled" so the result does not contain
// appended ErrorList elements.
// The method modifies the underlying slice.
func (e *ErrorList) AddErrors(errs ...error) {
	for _, err := range errs {
		var errList ErrorList
		if errors.As(err, &errList) {
			e.AddErrors(errList...)
		} else {
			*e = append(*e, err)
		}
	}
}

// StackTrace returns error stack
func (e *ServiceError) StackTrace() errors.StackTrace {
	if e.cause == nil {
		return nil
	}

	err, ok := e.cause.(stackTracer)
	if !ok {
		return nil
	}

	return err.StackTrace()
}

// ErrNotAuthorized is returned when the identity invoking the API does not have permission to invoke that operation.
var ErrNotAuthorized = errors.New("NotAuthorized")

// ErrNotAuthenticated is returned when anonymously invoking an endpoint that requires an identity
var ErrNotAuthenticated = errors.New("NotAuthenticated")

// ErrInvalidRequest is returned when some part of the request is incompatible with another part.
var ErrInvalidRequest = errors.New("InvalidRequest")

// NewErrorFromHTTPStatusCode returns error from http code
func NewErrorFromHTTPStatusCode(httpCode int, reason string, values ...interface{}) *ServiceError {
	if httpCode >= http.StatusBadRequest && httpCode < http.StatusInternalServerError {
		switch httpCode {
		case http.StatusUnauthorized:
			return Unauthorized(reason, values...)
		case http.StatusForbidden:
			return Forbidden(reason, values...)
		case http.StatusNotFound:
			return NotFound(reason, values...)
		case http.StatusMethodNotAllowed:
			return NotImplemented(reason, values...)
		case http.StatusConflict:
			return Conflict(reason, values...)
		//StatusBadRequest and all other errors will result in BadRequest error being created
		default:
			return BadRequest(reason, values...)
		}
	}

	if httpCode >= http.StatusInternalServerError {
		switch httpCode {
		//StatusInternalServerError and all other errors will result in GeneralError() error being created
		default:
			return GeneralError(reason, values...)
		}
	}

	return GeneralError(reason, values...)
}

// Unauthorized error
func Unauthorized(reason string, values ...interface{}) *ServiceError {
	return New(ErrorUnauthorized, reason, values...)
}

// GeneralError error
func GeneralError(reason string, values ...interface{}) *ServiceError {
	return New(ErrorGeneral, reason, values...)
}

// NotFound error
func NotFound(reason string, values ...interface{}) *ServiceError {
	return New(ErrorNotFound, reason, values...)
}

// Unauthenticated error
func Unauthenticated(reason string, values ...interface{}) *ServiceError {
	return New(ErrorUnauthenticated, reason, values...)
}

// Forbidden error
func Forbidden(reason string, values ...interface{}) *ServiceError {
	return New(ErrorForbidden, reason, values...)
}

// NotImplemented error
func NotImplemented(reason string, values ...interface{}) *ServiceError {
	return New(ErrorNotImplemented, reason, values...)
}

// Conflict error
func Conflict(reason string, values ...interface{}) *ServiceError {
	return New(ErrorConflict, reason, values...)
}

// Validation error
func Validation(reason string, values ...interface{}) *ServiceError {
	return New(ErrorValidation, reason, values...)
}

// BadRequest error
func BadRequest(reason string, values ...interface{}) *ServiceError {
	return New(ErrorBadRequest, reason, values...)
}

// ServiceErrorCode error codes
type ServiceErrorCode int

// ServiceErrors service errors
type ServiceErrors []ServiceError

// CodeStr Add error code prefix
func CodeStr(code ServiceErrorCode) string {
	return fmt.Sprintf("%s-%d", ErrorCodePrefix, code)
}

// AsError request
func (e *ServiceError) AsError() error {
	return fmt.Errorf(e.Error())
}

func (e *ServiceError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s\n caused by: %s", CodeStr(e.Code), e.Reason, e.cause.Error())
	}

	return fmt.Sprintf("%s: %s", CodeStr(e.Code), e.Reason)
}

func (e *ServiceError) Unwrap() error {
	return e.cause
}

// Is404 request
func (e *ServiceError) Is404() bool {
	return e.Code == NotFound("").Code
}

// IsConflict Request
func (e *ServiceError) IsConflict() bool {
	return e.Code == Conflict("").Code
}

// IsForbidden Request
func (e *ServiceError) IsForbidden() bool {
	return e.Code == Forbidden("").Code
}
