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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// GrpcServer represents a GrpcServer host service
type GrpcServer struct {
	services Services
}

// CheckPermission processes an authorization check and returns whether or not the operation would be allowed
func (r *GrpcServer) CheckPermission(ctx context.Context, rpcReq *core.CheckPermissionRequest) (*core.CheckPermissionResponse, error) {

	if token, ok := getBearerTokenFromContext(ctx, []string{"grpcgateway-authorization", "bearer-token"}); ok { //'bearer-token' is a guess at the metadata key for a token in a gRPC request

		req := contracts.CheckRequest{
			Request: contracts.Request{
				Requestor: app.Principal{ID: token},
			},
			Subject:   app.Principal{ID: rpcReq.Subject},
			Operation: rpcReq.Operation,
			Resource:  app.Resource{Type: rpcReq.Resourcetype, ID: rpcReq.Resourceid},
		}

		action := controllers.NewAccess(r.services.Store)

		result, err := action.Check(req)

		if err != nil {
			return nil, err
		}

		return &core.CheckPermissionResponse{Result: result}, nil
	}

	return nil, errors.New("Missing identity") //401?
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

func getBearerTokenFromContext(ctx context.Context, names []string) (string, bool) {
	for _, name := range names {
		if metadata, ok := metadata.FromIncomingContext(ctx); ok {
			headers := metadata.Get(name)
			if len(headers) > 0 {
				return headers[0], true
			}
		}
	}

	return "", false
}
