package server

import (
	apicontracts "authz/api/contracts"
	core "authz/api/gen/v1alpha"
	"authz/api/handler"
	"authz/domain/model"
	vo "authz/domain/valueobjects"
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

// GrpcGatewayServer represents a GrpcServer host service
type GrpcGatewayServer struct {
	PermissionHandler *handler.PermissionHandler
}

// NewServer creates a new Server object to use.
func (r *GrpcGatewayServer) NewServer(h handler.PermissionHandler) apicontracts.Server {
	return &GrpcGatewayServer{PermissionHandler: &h}
}

// Serve exposes a GRPC endpoint and blocks until processing ends, at which point the waitgroup is signalled. This should be run as a goroutine.
func (r *GrpcGatewayServer) Serve(wait *sync.WaitGroup, ports ...string) error {
	defer wait.Done()

	ls, err := net.Listen("tcp", ":"+ports[0])

	if err != nil {
		glog.Errorf("Error opening TCP port: %s", err)
		return err
	}

	var creds credentials.TransportCredentials
	if _, err = os.Stat("/etc/tls/tls.crt"); err == nil {
		if _, err := os.Stat("/etc/tls/tls.key"); err == nil { //Cert and key exists start server in TLS mode
			glog.Info("TLS cert and Key found  - Starting gRPC server in secure TLS mode")

			creds, err = credentials.NewServerTLSFromFile("/etc/tls/tls.crt", "/etc/tls/tls.key")
			if err != nil {
				glog.Errorf("Error loading certs: %s", err)
				return err
			}
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Infof("TLS cert or Key not found  - Starting gRPC server in insecure mode on port %s", ports[0])
	}

	srv := grpc.NewServer(grpc.Creds(creds))
	core.RegisterCheckPermissionServer(srv, r)
	err = srv.Serve(ls)
	if err != nil {
		glog.Errorf("Error hosting gRPC service: %s", err)
		return err
	}
	return nil
}

// GetName returns the impl name
func (r *GrpcGatewayServer) GetName() string {
	return "grpc"
}

// CheckPermission processes an authorization check and returns whether or not the operation would be allowed
func (r *GrpcGatewayServer) CheckPermission(ctx context.Context, rpcReq *core.CheckPermissionRequest) (*core.CheckPermissionResponse, error) {

	result, err := r.check(ctx, rpcReq)

	if err != nil {
		return nil, convertDomainErrorToGrpc(err)
	}

	return &core.CheckPermissionResponse{Result: bool(result)}, nil
}

func (r *GrpcGatewayServer) check(ctx context.Context, rpcReq *core.CheckPermissionRequest) (vo.AccessDecision, error) {
	req := handler.CheckRequest{
		Requestor:    getRequestorIdentityFromGrpcContext(ctx),
		Subject:      rpcReq.Subject,
		Operation:    rpcReq.Operation,
		ResourceType: rpcReq.Resourcetype,
		ResourceID:   rpcReq.Resourceid,
	}

	result, err := r.PermissionHandler.Check(req)
	return result, err
}

func getRequestorIdentityFromGrpcContext(ctx context.Context) model.Principal {
	for _, name := range []string{"grpcgateway-authorization", "bearer-token"} {
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			headers := md.Get(name)
			if len(headers) > 0 {
				return model.NewPrincipal(headers[0])
			}
		}
	}

	return model.NewAnonymousPrincipal()
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
