package host

import (
	core "authz/api/gen/v1alpha"
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

func (r *GrpcServer) CreateSeats(ctx context.Context, rpcReq *core.ModifySeatsRequest) (*core.ModifySeatsResponse, error) {
	requestor, err := r.getRequestorIdentityFromContext(ctx)
	if err != nil {
		return nil, err
	}

	principals, err := r.convertSubjectIdsToPrincipals(rpcReq.Subjects)
	if err != nil {
		return nil, err
	}

	req := contracts.ModifySeatAssignmentRequest{
		Request: contracts.Request{
			Requestor: requestor,
		},
		Principals: principals,
		Org:        app.Organization{Id: rpcReq.TenantId},
		Service:    app.Service{Id: rpcReq.ServiceId},
	}

	lic := controllers.NewLicensing(r.services.Licensing, r.services.Authz)

	if err := lic.AssignSeats(req); err != nil {
		return nil, convertDomainErrorToGrpc(err)
	}

	return &core.ModifySeatsResponse{}, nil
}

func (r *GrpcServer) DeleteSeats(ctx context.Context, rpcReq *core.ModifySeatsRequest) (*core.ModifySeatsResponse, error) {
	requestor, err := r.getRequestorIdentityFromContext(ctx)
	if err != nil {
		return nil, err
	}

	principals, err := r.convertSubjectIdsToPrincipals(rpcReq.Subjects)
	if err != nil {
		return nil, err
	}

	req := contracts.ModifySeatAssignmentRequest{
		Request: contracts.Request{
			Requestor: requestor,
		},
		Principals: principals,
		Org:        app.Organization{Id: rpcReq.TenantId},
		Service:    app.Service{Id: rpcReq.ServiceId},
	}

	lic := controllers.NewLicensing(r.services.Licensing, r.services.Authz)

	if err := lic.UnAssignSeats(req); err != nil {
		return nil, convertDomainErrorToGrpc(err)
	}

	return &core.ModifySeatsResponse{}, nil
}

func (r *GrpcServer) GetSeats(ctx context.Context, rpcReq *core.GetSeatsRequest) (*core.GetSeatsResponse, error) {
	return nil, nil
}

// CheckPermission processes an authorization check and returns whether or not the operation would be allowed
func (r *GrpcServer) CheckPermission(ctx context.Context, rpcReq *core.CheckPermissionRequest) (*core.CheckPermissionResponse, error) {
	requestor, err := r.getRequestorIdentityFromContext(ctx)
	if err != nil {
		return nil, err
	}
	subject, err := r.services.Principals.GetByID(rpcReq.Subject)
	if err != nil {
		return nil, err
	}

	req := contracts.CheckRequest{
		Request: contracts.Request{
			Requestor: requestor,
		},
		Subject:   subject,
		Operation: rpcReq.Operation,
		Resource:  app.Resource{Type: rpcReq.Resourcetype, ID: rpcReq.Resourceid},
	}

	action := controllers.NewAccess(r.services.Authz)

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
	core.RegisterSeatsServiceServer(srv, r)
	err = srv.Serve(ls)
	if err != nil {
		glog.Errorf("Error hosting gRPC service: %s", err)
		return
	}
}

func (r *GrpcServer) convertSubjectIdsToPrincipals(subjectIds []string) ([]app.Principal, error) {
	principals := make([]app.Principal, len(subjectIds))
	for i, subId := range subjectIds {
		if principal, err := r.services.Principals.GetByID(subId); err != nil {
			return nil, err
		} else {
			principals[i] = principal
		}
	}

	return principals, nil
}

func convertDomainErrorToGrpc(err error) error {
	switch {
	case errors.Is(err, app.ErrNotAuthenticated):
		return status.Error(codes.Unauthenticated, "Anonymous access is not allowed.")
	case errors.Is(err, app.ErrNotAuthorized):
		return status.Error(codes.PermissionDenied, "Access denied.")
	case errors.Is(err, app.ErrInvalidRequest):
		return status.Error(codes.InvalidArgument, "Problem with request.")
	default:
		return status.Error(codes.Unknown, "Internal server error.")
	}
}

func (r *GrpcServer) getRequestorIdentityFromContext(ctx context.Context) (app.Principal, error) {
	for _, name := range []string{"grpcgateway-authorization", "bearer-token"} {
		if metadata, ok := metadata.FromIncomingContext(ctx); ok {
			headers := metadata.Get(name)
			if len(headers) > 0 {
				if sub, err := r.services.Principals.GetByToken(headers[0]); err != nil {
					return sub, err
				} else {
					return sub, nil
				}
			}
		}
	}

	return app.NewAnonymousPrincipal(), nil
}
