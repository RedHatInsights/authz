// Package grpc implements the grpc server of the grpc gateway
package grpc

import (
	"authz/api"
	core "authz/api/gen/v1alpha"
	"authz/application"
	"authz/domain/model"
	"context"
	"errors"
	"net"
	"os"
	"sync"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Server represents a Server host service
type Server struct {
	AccessAppService  *application.AccessAppService
	LicenseAppService *application.LicenseAppService
	ServerConfig      *api.ServerConfig
}

// GetLicense ToDo - just a stub for now.
func (s *Server) GetLicense(_ context.Context, _ *core.GetLicenseRequest) (*core.GetLicenseResponse, error) {
	//TODO implement me
	panic("implement me")
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
func (s *Server) GetSeats(_ context.Context, _ *core.GetSeatsRequest) (*core.GetSeatsResponse, error) {
	//TODO implement me
	panic("implement me")
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

	srv := grpc.NewServer(grpc.Creds(creds))
	core.RegisterCheckPermissionServer(srv, s)
	core.RegisterLicenseServiceServer(srv, s)
	err = srv.Serve(ls)
	if err != nil {
		glog.Errorf("Error hosting gRPC service: %s", err)
		return err
	}
	return nil
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
	for _, name := range []string{"grpcgateway-authorization", "bearer-token"} {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			headers := md.Get(name)
			if len(headers) > 0 {
				return convertTokenToPrincipalID(headers[0])
			}
		}
	}

	return "", nil
}

func convertTokenToPrincipalID(token string) (string, error) {
	return token, nil //Placeholder for token introspection
}

func convertDomainErrorToGrpc(err error) error {
	switch {
	case errors.Is(err, model.ErrNotAuthenticated):
		return status.Error(codes.Unauthenticated, "Anonymous access is not allowed.")
	case errors.Is(err, model.ErrNotAuthorized):
		return status.Error(codes.PermissionDenied, "Access denied.")
	default:
		return status.Error(codes.Unknown, "Internal server error.")
	}
}
