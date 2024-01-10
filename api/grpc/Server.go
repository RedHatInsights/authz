// Package grpc implements the grpc server of the grpc gateway
package grpc

import (
	core "authz/api/gen/v1alpha"
	"authz/api/grpc/interceptor"
	"authz/application"
	"authz/bootstrap/serviceconfig"
	"authz/domain"
	"context"
	"net"
	"os"
	"sync"

	"github.com/golang/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Server represents a Server host service
type Server struct {
	srv              *grpc.Server
	AccessAppService *application.AccessAppService
	ServiceConfig    *serviceconfig.ServiceConfig
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
func NewServer(h application.AccessAppService, c serviceconfig.ServiceConfig) *Server {
	return &Server{AccessAppService: &h, ServiceConfig: &c}
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

	var authnhandler grpc.UnaryServerInterceptor
	// TODO: Evaluate better way to init. This impl is ugly, but `...ServerOptions` (2nd param in NewServer call) is an interface
	if anyEnabled(s.ServiceConfig.AuthConfigs) {
		authMiddleware, err := interceptor.NewAuthnInterceptor(s.ServiceConfig.AuthConfigs)
		if err != nil {
			glog.Fatalf("Error: Not able to reach discovery endpoint to initialize authentication middleware.")
		}
		authnhandler = authMiddleware.Unary()
	} else {
		// local dev: no authconfig given, so we enable a passthrough middleware to get the requestor from authorization header.
		authMiddleware := interceptor.NewPassthroughAuthnInterceptor()
		glog.Warning("Client authorization disabled. Do not use in production use cases!")
		authnhandler = authMiddleware.Unary()
	}
	errorHandler := interceptor.NewErrorConvertingInterceptor().Unary()
	s.srv = grpc.NewServer(grpc.Creds(creds), grpc.ChainUnaryInterceptor(authnhandler, errorHandler))

	core.RegisterHealthCheckServiceServer(s.srv, s)
	core.RegisterCheckPermissionServer(s.srv, s)

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
		return nil, domain.ErrNotAuthorized
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
		return nil, err
	}

	return &core.CheckPermissionResponse{Result: bool(result)}, nil
}

func (s *Server) getRequestorIdentityFromGrpcContext(ctx context.Context) (string, error) {
	requestor := ctx.Value(interceptor.RequestorContextKey)
	reqStr := requestor.(string)
	if reqStr == "" {
		return "", domain.ErrNotAuthenticated
	}

	return reqStr, nil
}

//func (s *Server) getIsOrgAdminFromGrpcContext(ctx context.Context) (isOrgAdmin bool) {
//	requestor := ctx.Value(interceptor.IsRequestorOrgAdminContextKey)
//	isOrgAdmin = requestor.(bool)
//	return
//}

//func (s *Server) getRequestorOrgIDFromGrpcContext(ctx context.Context) (result string) {
//	requestorOrgID := ctx.Value(interceptor.RequestorOrgContextKey)
//	result = requestorOrgID.(string)
//	return
//}

func isRequestorOrgAdmin(requestorOrgAdmin bool, requestorOrgID, orgIDInRequest string) bool {
	if requestorOrgAdmin && requestorOrgID == orgIDInRequest { // Requestor should be org Admin of the org received in the request path
		return true
	}
	return false
}
