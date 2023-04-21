// Package http implements the http server. For GRPC Gateway, it references the actual grpc server.
package http

import (
	"authz/api"
	core "authz/api/gen/v1alpha"
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
)

// Server serves a HTTP api based on the generated grpc gateway code
type Server struct {
	ServerConfig       *api.ServerConfig
	GrpcCheckService   core.CheckPermissionServer
	GrpcLicenseService core.LicenseServiceServer
}

// Serve starts serving
func (s *Server) Serve(wait *sync.WaitGroup) error {
	defer wait.Done()

	mux, err := createMultiplexer(s.GrpcCheckService, s.GrpcLicenseService)
	if err != nil {
		glog.Errorf("Error creating multiplexer: %s", err)
		return err
	}

	if _, err = os.Stat(s.ServerConfig.TLSConfig.CertPath); err == nil {
		if _, err := os.Stat(s.ServerConfig.TLSConfig.KeyPath); err == nil { //Cert and key exists start server in HTTPS mode
			glog.Infof("TLS cert and Key found  - Starting server in secure HTTPS mode on port %s",
				s.ServerConfig.HTTPSPort)

			err = http.ListenAndServeTLS(
				":"+s.ServerConfig.HTTPSPort,
				s.ServerConfig.TLSConfig.CertPath, //TODO: Needs sanity checking.
				s.ServerConfig.TLSConfig.KeyPath, mux)
			if err != nil {
				glog.Errorf("Error hosting TLS service: %s", err)
				return err
			}
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Infof("TLS cert or Key not found  - Starting server in insecure plain HTTP mode on Port %s",
			s.ServerConfig.HTTPPort)
		err = http.ListenAndServe(":"+s.ServerConfig.HTTPPort, mux)

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
func (s *Server) SetSeatRef(ss core.LicenseServiceServer) {
	s.GrpcLicenseService = ss
}

// NewServer creates a new Server object to use.
func NewServer(c api.ServerConfig) *Server {
	return &Server{
		ServerConfig: &c,
	}
}

// GetName returns the Name of the impl
func (s *Server) GetName() string {
	return "grpcweb"
}

func createMultiplexer(h1 core.CheckPermissionServer, h2 core.LicenseServiceServer) (http.Handler, error) {
	mux := runtime.NewServeMux()

	if err := core.RegisterCheckPermissionHandlerServer(context.Background(), mux, h1); err != nil {
		return nil, err
	}

	if err := core.RegisterLicenseServiceHandlerServer(context.Background(), mux, h2); err != nil {
		return nil, err
	}

	chain := createChain(logMiddleware, corsMiddleware).then(mux)

	return chain, nil
}

func corsMiddleware(h http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedMethods:   []string{http.MethodHead, http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "ResponseType", "Content-Length", "Accept-Encoding", "Authorization", "Content-Type", "User-Agent"},
		AllowCredentials: true,
		MaxAge:           300,
		Debug:            true,
	}).Handler(h)
}

func logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		glog.V(0).Infof("Request incoming: %s %s", r.Method, r.RequestURI)
		glog.V(1).Infof("Request dump: %+v", *r)

		h.ServeHTTP(w, r)
	})
}
