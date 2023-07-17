package interceptor

import (
	"authz/domain"
	"context"
	"errors"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorConvertingInterceptor -
type ErrorConvertingInterceptor struct{}

// NewErrorConvertingInterceptor -
func NewErrorConvertingInterceptor() *ErrorConvertingInterceptor {
	return &ErrorConvertingInterceptor{}
}

// Unary -
func (i *ErrorConvertingInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = handler(ctx, req)

		if err != nil {
			err = convertDomainErrorToGrpc(err)
		}
		return
	}
}

func convertDomainErrorToGrpc(err error) error {
	var validationErr domain.ErrInvalidRequest

	_, ok := status.FromError(err) //if already gRPC error, don't convert
	if ok {
		return err
	}

	switch {

	case errors.Is(err, domain.ErrNotAuthenticated):
		return status.Error(codes.Unauthenticated, "Anonymous access is not allowed.")
	case errors.Is(err, domain.ErrNotAuthorized):
		return status.Error(codes.PermissionDenied, "Access denied.")
	case errors.Is(err, domain.ErrLicenseLimitExceeded):
		return status.Error(codes.FailedPrecondition, "License limits exceeded.")
	case errors.Is(err, domain.ErrConflict):
		return status.Error(codes.FailedPrecondition, "Conflict")
	case errors.As(err, &validationErr):
		glog.Errorf("Validation error: %s", validationErr.Reason)
		return status.Error(codes.InvalidArgument, validationErr.Reason)
	default:
		glog.Errorf("Unhandled error: %+v", err)
		return status.Error(codes.Unknown, "Internal server error.")
	}
}
