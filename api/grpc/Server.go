// Package grpc implements the grpc server of the grpc gateway
package grpc

import (
	core "authz/api/gen/v1alpha"
	"authz/api/grpc/interceptor"
	"authz/application"
	"authz/bootstrap/serviceconfig"
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
	srv               *grpc.Server
	AccessAppService  *application.AccessAppService
	LicenseAppService *application.LicenseAppService
	ServiceConfig     *serviceconfig.ServiceConfig
}

// GetLicense returns licenses for a given org and service
func (s *Server) GetLicense(ctx context.Context, grpcReq *core.GetLicenseRequest) (*core.GetLicenseResponse, error) {
	requestor, err := s.getRequestorIdentityFromGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	// Validate if the requestor is the orgAdmin of the org in the req
	requestorOrgAdmin := s.getIsOrgAdminFromGrpcContext(ctx)
	requestorOrgID := s.getRequestorOrgIDFromGrpcContext(ctx)

	if !isRequestorOrgAdmin(requestorOrgAdmin, requestorOrgID, grpcReq.OrgId) {
		return nil, convertDomainErrorToGrpc(domain.ErrNotAuthorized)
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
		SeatsTotal:     int64(limit),
		SeatsAvailable: int64(available),
	}, nil
}

// ModifySeats assigns and/or unassigns users to/from seats for a given org and service
func (s *Server) ModifySeats(ctx context.Context, grpcReq *core.ModifySeatsRequest) (*core.ModifySeatsResponse, error) {
	requestor, err := s.getRequestorIdentityFromGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	// Validate if the requestor is the orgAdmin of the org in the req
	requestorOrgAdmin := s.getIsOrgAdminFromGrpcContext(ctx)
	requestorOrgID := s.getRequestorOrgIDFromGrpcContext(ctx)

	if !isRequestorOrgAdmin(requestorOrgAdmin, requestorOrgID, grpcReq.OrgId) {
		return nil, convertDomainErrorToGrpc(domain.ErrNotAuthorized)
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

// GetSeats returns seats for a given org and service
func (s *Server) GetSeats(ctx context.Context, grpcReq *core.GetSeatsRequest) (*core.GetSeatsResponse, error) {
	requestor, err := s.getRequestorIdentityFromGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	// Validate if the requestor is the orgAdmin of the org in the req
	requestorOrgAdmin := s.getIsOrgAdminFromGrpcContext(ctx)
	requestorOrgID := s.getRequestorOrgIDFromGrpcContext(ctx)

	if !isRequestorOrgAdmin(requestorOrgAdmin, requestorOrgID, grpcReq.OrgId) {
		return nil, convertDomainErrorToGrpc(domain.ErrNotAuthorized)
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

// EntitleOrg entitles an Org for a license to an existing service.
// TODO: This is a temporary domain endpoint / handler until we get this from another service. So e.g. no serviceId-exists check will be added. We assume that the service is already there instead for now. Also updates not included for now.
func (s *Server) EntitleOrg(ctx context.Context, entitleOrgReq *core.EntitleOrgRequest) (*core.EntitleOrgResponse, error) {
	requestor, err := s.getRequestorIdentityFromGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	if !sliceContains(s.ServiceConfig.AuthzConfig.LicenseImportAllowlist, requestor) {
		glog.Infof("Received request to entitle Org: %s from Requestor: %s. Requestor. Requestor not authorized. ", entitleOrgReq.OrgId, requestor)
		return nil, convertDomainErrorToGrpc(domain.ErrNotAuthorized)
	}

	if entitleOrgReq.MaxSeats < 1 {
		return nil, errors.New("maxSeats value not valid")
	}
	glog.Infof("Received request to entitle Org: %s for a license to service %s with %v seats from Requestor: %s", entitleOrgReq.OrgId, entitleOrgReq.ServiceId, entitleOrgReq.MaxSeats, requestor)
	evt := application.OrgEntitledEvent{
		OrgID:     entitleOrgReq.OrgId,
		ServiceID: entitleOrgReq.ServiceId,
		MaxSeats:  int(entitleOrgReq.MaxSeats),
	}

	err = s.LicenseAppService.HandleOrgEntitledEvent(evt)
	if err != nil {
		return nil, err
	}

	resp := &core.EntitleOrgResponse{}

	return resp, nil
}

// ImportOrg imports users for a given orgID
func (s *Server) ImportOrg(ctx context.Context, importReq *core.ImportOrgRequest) (*core.ImportOrgResponse, error) {
	requestor, err := s.getRequestorIdentityFromGrpcContext(ctx)
	if err != nil {
		return nil, err
	}

	if !sliceContains(s.ServiceConfig.AuthzConfig.LicenseImportAllowlist, requestor) {
		glog.Infof("Received request to import Org: %s from Requestor: %s. Requestor. Requestor not authorized. ", importReq.OrgId, requestor)
		return nil, convertDomainErrorToGrpc(domain.ErrNotAuthorized)
	}

	glog.Infof("Received request to import users for Org: %s from Requestor: %s", importReq.OrgId, requestor)
	evt := application.ImportOrgEvent{
		OrgID: importReq.OrgId,
	}
	result, e2 := s.LicenseAppService.ImportUsersForOrg(evt)

	if e2 != nil {
		return nil, e2
	}

	return &core.ImportOrgResponse{
		ImportedUsersCount:    result.ImportedUsersCount,
		NotImportedUsersCount: result.NotImportedUsersCount,
	}, nil
}

// HealthCheck - heathcheck implementation returns 200 OK
func (s *Server) HealthCheck(_ context.Context, _ *core.Empty) (*core.Empty, error) {
	return &core.Empty{}, nil
}

func sliceContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

// NewServer creates a new Server object to use.
func NewServer(h application.AccessAppService, l application.LicenseAppService, c serviceconfig.ServiceConfig) *Server {
	return &Server{AccessAppService: &h, ServiceConfig: &c, LicenseAppService: &l}
}

// Serve exposes a GRPC endpoint and blocks until processing ends, at which point the waitgroup is signalled. This should be run as a goroutine.
func (s *Server) Serve(wait *sync.WaitGroup) error {
	defer wait.Done()

	ls, err := net.Listen("tcp", ":"+s.ServiceConfig.GrpcPortStr)

	if err != nil {
		glog.Errorf("Error opening TCP port: %s", err)
		return err
	}

	var creds credentials.TransportCredentials

	if _, err = os.Stat(s.ServiceConfig.TLSConfig.CertFile); err == nil {
		if _, err := os.Stat(s.ServiceConfig.TLSConfig.KeyFile); err == nil { // Cert and key exists start server in TLS mode
			glog.Info("TLS cert and Key found  - Starting gRPC server in secure TLS mode")

			creds, err = credentials.NewServerTLSFromFile(s.ServiceConfig.TLSConfig.CertFile, s.ServiceConfig.TLSConfig.KeyFile)
			if err != nil {
				glog.Errorf("Error loading certs: %s", err)
				return err
			}
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Infof("TLS cert or Key not found  - Starting gRPC server in insecure mode on port %s",
			s.ServiceConfig.GrpcPortStr)
	}

	// TODO: Evaluate better way to init. This impl is ugly, but `...ServerOptions` (2nd param in NewServer call) is an interface
	if anyEnabled(s.ServiceConfig.AuthConfigs) {
		authMiddleware, err := interceptor.NewAuthnInterceptor(s.ServiceConfig.AuthConfigs)
		if err != nil {
			glog.Fatalf("Error: Not able to reach discovery endpoint to initialize authentication middleware.")
		}
		s.srv = grpc.NewServer(grpc.Creds(creds), grpc.ChainUnaryInterceptor(authMiddleware.Unary()))
	} else {
		// local dev: no authconfig given, so we enable a passthrough middleware to get the requestor from authorization header.
		authMiddleware := interceptor.NewPassthroughAuthnInterceptor()
		glog.Warning("Client authorization disabled. Do not use in production use cases!")
		s.srv = grpc.NewServer(grpc.Creds(creds), grpc.ChainUnaryInterceptor(authMiddleware.Unary()))
	}

	core.RegisterHealthCheckServiceServer(s.srv, s)
	core.RegisterCheckPermissionServer(s.srv, s)
	core.RegisterLicenseServiceServer(s.srv, s)
	core.RegisterImportServiceServer(s.srv, s)

	err = s.srv.Serve(ls)
	if err != nil {
		glog.Errorf("Error hosting gRPC service: %s", err)
		return err
	}
	return nil
}

func anyEnabled(authConfigs []serviceconfig.AuthConfig) bool {
	for _, config := range authConfigs {
		if config.Enabled {
			return true
		}
	}

	return false
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

	if !sliceContains(s.ServiceConfig.AuthzConfig.CheckAllowList, requestor) {
		glog.Infof("Received CheckPermission from Requestor: %s. Requestor not authorized. Request: %v", requestor, rpcReq)
		return nil, convertDomainErrorToGrpc(domain.ErrNotAuthorized)
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
	var validationErr domain.ErrInvalidRequest

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
		return status.Error(codes.InvalidArgument, validationErr.Reason)
	default:
		return status.Error(codes.Unknown, "Internal server error.")
	}
}

func (s *Server) getIsOrgAdminFromGrpcContext(ctx context.Context) (isOrgAdmin bool) {
	requestor := ctx.Value(interceptor.IsRequestorOrgAdminContextKey)
	isOrgAdmin = requestor.(bool)
	return
}

func (s *Server) getRequestorOrgIDFromGrpcContext(ctx context.Context) (result string) {
	requestorOrgID := ctx.Value(interceptor.RequestorOrgContextKey)
	result = requestorOrgID.(string)
	return
}

func isRequestorOrgAdmin(requestorOrgAdmin bool, requestorOrgID, orgIDInRequest string) bool {
	if requestorOrgAdmin && requestorOrgID == orgIDInRequest { // Requestor should be org Admin of the org received in the request path
		return true
	}
	return false
}
