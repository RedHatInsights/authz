package host

import (
	core "authz/api/gen/v1"
	"authz/app"
	"authz/app/contracts"
	"authz/app/controllers"
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

// GrpcServer represents a GrpcServer host service
type GrpcServer struct {
	services Services
}

// CheckPermission processes an authorization check and returns whether or not the operation would be allowed
func (r *GrpcServer) CheckPermission(ctx context.Context, rpcReq *core.CheckPermissionRequest) (*core.CheckPermissionResponse, error) {

	req := contracts.CheckRequest{
		Request: contracts.Request{
			Requestor: getRequestorIdentityFromContext(ctx),
		},
		Subject:   app.Principal{ID: rpcReq.Subject},
		Operation: rpcReq.Operation,
		Resource:  app.Resource{Type: rpcReq.Resourcetype, ID: rpcReq.Resourceid},
	}

	action := controllers.NewAccess(r.services.Store)

	result, err := action.Check(req)

	if err != nil {
		return nil, convertDomainErrorToGrpc(err)
	}

	return &core.CheckPermissionResponse{Result: result}, nil
}

// NewGrpcServer instantiates a new GRpc host service
func NewGrpcServer(services Services) *GrpcServer {
	return &GrpcServer{services: services}
}

// Host exposes a GRPC endpoint and blocks until processing ends, at which point the waitgroup is signalled. This should be run as a goroutine.
func (r *GrpcServer) Host(wait *sync.WaitGroup) {
	defer wait.Done()

	ls, err := net.Listen("tcp", ":8081")

	if err != nil {
		glog.Errorf("Error opening TCP port: %s", err)
		return
	}

	var creds credentials.TransportCredentials = nil
	if _, err = os.Stat("/etc/tls/tls.crt"); err == nil {
		if _, err := os.Stat("/etc/tls/tls.key"); err == nil { //Cert and key exists start server in TLS mode
			glog.Info("TLS cert and Key found  - Starting gRPC server in secure TLS mode")

			creds, err = credentials.NewServerTLSFromFile("/etc/tls/tls.crt", "/etc/tls/tls.key")
			if err != nil {
				glog.Errorf("Error loading certs: %s", err)
				return
			}
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Info("TLS cert or Key not found  - Starting gRPC server in insecure mode")
	}

	srv := grpc.NewServer(grpc.Creds(creds))
	core.RegisterCheckPermissionServer(srv, r)
	err = srv.Serve(ls)
	if err != nil {
		glog.Errorf("Error hosting gRPC service: %s", err)
		return
	}
}

func convertDomainErrorToGrpc(err error) error {
	switch {
	case errors.Is(err, app.ErrNotAuthenticated):
		return status.Error(codes.Unauthenticated, "Anonymous access is not allowed.")
	case errors.Is(err, app.ErrNotAuthorized):
		return status.Error(codes.PermissionDenied, "Access denied.")
	default:
		return status.Error(codes.Unknown, "Internal server error.")
	}
}

func getRequestorIdentityFromContext(ctx context.Context) app.Principal {
	for _, name := range []string{"grpcgateway-authorization", "bearer-token"} {
		if metadata, ok := metadata.FromIncomingContext(ctx); ok {
			headers := metadata.Get(name)
			if len(headers) > 0 {
				return app.NewPrincipal(headers[0])
			}
		}
	}

	return app.NewAnonymousPrincipal()
}
