// Package grpc implements the grpc server of the grpc gateway
package grpc

import (
	"authz/api"
	core "authz/api/gen/v1alpha"
	"authz/api/grpc/interceptor"
	"authz/application"
	"authz/domain"
	"context"
	"errors"
	"net"
	"os"
	"sync"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// Server represents a Server host service
type Server struct {
	srv                      *grpc.Server
	UnaryServiceInterceptors *[]grpc.UnaryServerInterceptor
	AccessAppService         *application.AccessAppService
	LicenseAppService        *application.LicenseAppService
	ServerConfig             *api.ServerConfig
}

// GetLicense ToDo - just a stub for now.
func (s *Server) GetLicense(ctx context.Context, grpcReq *core.GetLicenseRequest) (*core.GetLicenseResponse, error) {
	requestor, err := s.getRequestorIdentityFromGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	req := application.GetSeatAssignmentCountsRequest{
		Requestor: requestor,
		OrgID:     grpcReq.OrgId,
		ServiceID: grpcReq.ServiceId,
	}
	limit, available, err := s.LicenseAppService.GetSeatAssignmentCounts(req)
	if err != nil {
		return nil, err
	}

	return &core.GetLicenseResponse{
		SeatsTotal:     int32(limit),
		SeatsAvailable: int32(available),
	}, nil
}

// ModifySeats ToDo - just a stub for now.
func (s *Server) ModifySeats(ctx context.Context, grpcReq *core.ModifySeatsRequest) (*core.ModifySeatsResponse, error) {
	requestor, err := s.getRequestorIdentityFromGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	req := application.ModifySeatAssignmentRequest{
		Requestor: requestor,
		OrgID:     grpcReq.OrgId,
		ServiceID: grpcReq.ServiceId,
		Assign:    grpcReq.Assign,
		Unassign:  grpcReq.Unassign,
	}

	err = s.LicenseAppService.ModifySeats(req)

	if err != nil {
		return nil, convertDomainErrorToGrpc(err)
	}
	return &core.ModifySeatsResponse{}, nil
}

// GetSeats ToDo - just a stub for now.
func (s *Server) GetSeats(ctx context.Context, grpcReq *core.GetSeatsRequest) (*core.GetSeatsResponse, error) {
	requestor, err := s.getRequestorIdentityFromGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	includeUsers := true
	if grpcReq.IncludeUsers != nil {
		includeUsers = *grpcReq.IncludeUsers
	}

	assigned := true
	if grpcReq.Filter != nil {
		filter := *grpcReq.Filter
		switch filter {
		case core.SeatFilterType_assigned:
			assigned = true
		case core.SeatFilterType_assignable:
			assigned = false
		}
	}

	req := application.GetSeatAssignmentRequest{
		Requestor:    requestor,
		OrgID:        grpcReq.OrgId,
		ServiceID:    grpcReq.ServiceId,
		IncludeUsers: includeUsers,
		Assigned:     assigned,
	}

	principals, err := s.LicenseAppService.GetSeatAssignments(req)
	if err != nil {
		return nil, err
	}

	resp := &core.GetSeatsResponse{Users: make([]*core.GetSeatsUserRepresentation, len(principals))}
	for i, p := range principals {
		resp.Users[i] = &core.GetSeatsUserRepresentation{
			DisplayName: p.DisplayName,
			Id:          string(p.ID),
			Assigned:    assigned,
		}
	}

	return resp, nil
}

// NewServer creates a new Server object to use.
func NewServer(h application.AccessAppService, l application.LicenseAppService, c api.ServerConfig) *Server {
	return &Server{AccessAppService: &h, ServerConfig: &c, LicenseAppService: &l}
}

// Serve exposes a GRPC endpoint and blocks until processing ends, at which point the waitgroup is signalled. This should be run as a goroutine.
func (s *Server) Serve(wait *sync.WaitGroup) error {
	defer wait.Done()

	ls, err := net.Listen("tcp", ":"+s.ServerConfig.GrpcPort)

	if err != nil {
		glog.Errorf("Error opening TCP port: %s", err)
		return err
	}

	var creds credentials.TransportCredentials

	if _, err = os.Stat(s.ServerConfig.TLSConfig.CertPath); err == nil {
		if _, err := os.Stat(s.ServerConfig.TLSConfig.KeyPath); err == nil { //Cert and key exists start server in TLS mode
			glog.Info("TLS cert and Key found  - Starting gRPC server in secure TLS mode")

			creds, err = credentials.NewServerTLSFromFile(s.ServerConfig.TLSConfig.CertPath, s.ServerConfig.TLSConfig.KeyPath)
			if err != nil {
				glog.Errorf("Error loading certs: %s", err)
				return err
			}
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Infof("TLS cert or Key not found  - Starting gRPC server in insecure mode on port %s",
			s.ServerConfig.GrpcPort)
	}

	logMw := interceptor.NewAuthnInterceptor(s.ServerConfig.AuthConfig)
	s.srv = grpc.NewServer(grpc.Creds(creds), logMw.Unary())
	core.RegisterCheckPermissionServer(s.srv, s)
	core.RegisterLicenseServiceServer(s.srv, s)
	err = s.srv.Serve(ls)
	if err != nil {
		glog.Errorf("Error hosting gRPC service: %s", err)
		return err
	}
	return nil
}

// Stop gracefully stops the server.
func (s *Server) Stop() {
	s.srv.GracefulStop()
}

// GetName returns the impl name
func (s *Server) GetName() string {
	return "grpc"
}

// CheckPermission processes an authorization check and returns whether or not the operation would be allowed
func (s *Server) CheckPermission(ctx context.Context, rpcReq *core.CheckPermissionRequest) (*core.CheckPermissionResponse, error) {
	requestor, err := s.getRequestorIdentityFromGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	req := application.CheckRequest{
		Requestor:    requestor,
		Subject:      rpcReq.Subject,
		Operation:    rpcReq.Operation,
		ResourceType: rpcReq.Resourcetype,
		ResourceID:   rpcReq.Resourceid,
	}

	result, err := s.AccessAppService.Check(req)

	if err != nil {
		return nil, convertDomainErrorToGrpc(err)
	}

	return &core.CheckPermissionResponse{Result: bool(result)}, nil
}

func (s *Server) getRequestorIdentityFromGrpcContext(ctx context.Context) (string, error) {
	requestor := ctx.Value(interceptor.RequestorContextKey)
	reqStr := requestor.(string)
	if reqStr == "" {
		return "", convertDomainErrorToGrpc(domain.ErrNotAuthenticated)
	}

	return reqStr, nil
}

func convertDomainErrorToGrpc(err error) error {
	switch {
	case errors.Is(err, domain.ErrNotAuthenticated):
		return status.Error(codes.Unauthenticated, "Anonymous access is not allowed.")
	case errors.Is(err, domain.ErrNotAuthorized):
		return status.Error(codes.PermissionDenied, "Access denied.")
	case errors.Is(err, domain.ErrLicenseLimitExceeded):
		return status.Error(codes.FailedPrecondition, "License limits exceeded.")
	case errors.Is(err, domain.ErrConflict):
		return status.Error(codes.FailedPrecondition, "Conflict")
	default:
		return status.Error(codes.Unknown, "Internal server error.")
	}
}
