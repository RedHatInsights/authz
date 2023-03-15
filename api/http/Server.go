package http

import (
	core "authz/api/gen/v1alpha"
	"authz/bootstrap/config"
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// Server serves a HTTP api based on the generated grpc gateway code
type Server struct {
	ServerConfig     *config.ServerConfig
	GrpcCheckService core.CheckPermissionServer
	GrpcSeatsService core.SeatsServiceServer
}

// Serve starts serving
func (s *Server) Serve(wait *sync.WaitGroup) error {
	defer wait.Done()

	mux, err := createMultiplexer(s.GrpcCheckService, s.GrpcSeatsService)
	if err != nil {
		glog.Errorf("Error creating multiplexer: %s", err)
		return err
	}

	if _, err = os.Stat("/etc/tls/tls.crt"); err == nil {
		if _, err := os.Stat("/etc/tls/tls.key"); err == nil { //Cert and key exists start server in HTTPS mode
			glog.Infof("TLS cert and Key found  - Starting server in secure HTTPS mode on port %s",
				s.ServerConfig.GrpcWebHttpsPort)

			err = http.ListenAndServeTLS(
				":"+s.ServerConfig.GrpcWebHttpsPort,
				"/etc/tls/tls.crt", //TODO: Needs sanity checking and get from config.
				"/etc/tls/tls.key", mux)
			if err != nil {
				glog.Errorf("Error hosting TLS service: %s", err)
				return err
			}
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Infof("TLS cert or Key not found  - Starting server in insecure plain HTTP mode on Port %s",
			s.ServerConfig.GrpcWebHttpPort)
		err = http.ListenAndServe(":"+s.ServerConfig.GrpcWebHttpPort, mux)

		if err != nil {
			glog.Errorf("Error hosting insecure service: %s", err)
			return err
		}
	}
	return nil
}

// SetCheckRef sets the reference to the grpc CheckPermissionService
func (s *Server) SetCheckRef(h core.CheckPermissionServer) {
	s.GrpcCheckService = h
}

// SetSeatRef sets the reference to the grp SeatsServerService
func (s *Server) SetSeatRef(ss core.SeatsServiceServer) {
	s.GrpcSeatsService = ss
}

// NewServer creates a new Server object to use.
func NewServer(c config.ServerConfig) *Server {
	return &Server{
		ServerConfig: &c,
	}
}

// GetName returns the Name of the impl
func (s *Server) GetName() string {
	return "grpcweb"
}

func createMultiplexer(h1 core.CheckPermissionServer, h2 core.SeatsServiceServer) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux()

	if err := core.RegisterCheckPermissionHandlerServer(context.Background(), mux, h1); err != nil {
		return nil, err
	}

	if err := core.RegisterSeatsServiceHandlerServer(context.Background(), mux, h2); err != nil {
		return nil, err
	}

	return mux, nil
}
