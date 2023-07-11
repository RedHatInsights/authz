// Package http implements the http server. For GRPC Gateway, it references the actual grpc server.
package http

import (
	core "authz/api/gen/v1alpha"
	"authz/bootstrap/serviceconfig"
	"authz/infrastructure/grpcutil"
	"context"
	"errors"
	"net/http"
	"os"
	"sync"

	"google.golang.org/grpc/credentials/insecure"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"google.golang.org/grpc"
)

// Server serves an HTTP api based on the generated grpc gateway code
type Server struct {
	srv                *http.Server
	ServiceConfig      *serviceconfig.ServiceConfig
	GrpcCheckService   core.CheckPermissionServer
	GrpcLicenseService core.LicenseServiceServer
}

// Serve starts serving
func (s *Server) Serve(wait *sync.WaitGroup) error {
	defer wait.Done()

	mux, err := createMultiplexer(s.ServiceConfig)
	if err != nil {
		glog.Errorf("Error creating multiplexer: %s", err)
		return err
	}

	if _, err = os.Stat(s.ServiceConfig.TLSConfig.CertFile); err == nil {
		if _, err := os.Stat(s.ServiceConfig.TLSConfig.KeyFile); err == nil { //Cert and key exists start server in HTTPS mode
			glog.Infof("TLS cert and Key found  - Starting server in secure HTTPS mode on port %s",
				s.ServiceConfig.HTTPSPortStr)

			s.srv = &http.Server{Addr: ":" + s.ServiceConfig.HTTPSPortStr, Handler: mux}
			err := s.srv.ListenAndServeTLS(s.ServiceConfig.TLSConfig.CertFile, s.ServiceConfig.TLSConfig.KeyFile)
			if err != nil && !errors.Is(err, http.ErrServerClosed) { //ErrServerClosed is returned when the server stops serving
				glog.Errorf("Error hosting TLS service: %s", err)
				return err
			}
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Infof("TLS cert or Key not found  - Starting server in insecure plain HTTP mode on Port %s",
			s.ServiceConfig.HTTPPortStr)
		s.srv = &http.Server{Addr: ":" + s.ServiceConfig.HTTPPortStr, Handler: mux}
		err = s.srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) { //ErrServerClosed is returned when the server stops serving
			glog.Errorf("Error hosting insecure service: %s", err)
			return err
		}
	}
	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop() error {
	return s.srv.Shutdown(context.Background())
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
func NewServer(c serviceconfig.ServiceConfig) *Server {
	return &Server{
		ServiceConfig: &c,
	}
}

// GetName returns the Name of the impl
func (s *Server) GetName() string {
	return "grpcweb"
}

func createMultiplexer(cnf *serviceconfig.ServiceConfig) (http.Handler, error) {
	mux := runtime.NewServeMux()

	var opts []grpc.DialOption

	if _, err := os.Stat(cnf.TLSConfig.CertFile); err == nil {
		if _, err := os.Stat(cnf.TLSConfig.KeyFile); err == nil { //Cert and key exists start server in TLS mode
			glog.Info("Creating multiplexer for HTTP: TLS cert and Key found - connecting to gRPC server in secure TLS mode")

			if err != nil {
				glog.Errorf("Error loading certs: %s", err)
				return nil, err
			}

			//Skipping cert verification because the cert Subject doesn't cover loopback addresses
			sysCertOption, err := grpcutil.WithSystemCerts(grpcutil.SkipVerifyCA)
			if err != nil {
				return nil, err
			}
			opts = append(opts, sysCertOption)
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Infof("Creating multiplexer for HTTP: TLS cert or Key not found  - connecting to  gRPC server in insecure mode on port %s",
			cnf.GrpcPort)
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	if err := core.RegisterCheckPermissionHandlerFromEndpoint(context.Background(), mux, "localhost:"+cnf.GrpcPortStr, opts); err != nil {
		return nil, err
	}

	if err := core.RegisterLicenseServiceHandlerFromEndpoint(context.Background(), mux, "localhost:"+cnf.GrpcPortStr, opts); err != nil {
		return nil, err
	}

	if err := core.RegisterImportServiceHandlerFromEndpoint(context.Background(), mux, "localhost:"+cnf.GrpcPortStr, opts); err != nil {
		return nil, err
	}

	if err := core.RegisterHealthCheckServiceHandlerFromEndpoint(context.Background(), mux, "localhost:"+cnf.GrpcPortStr, opts); err != nil {
		return nil, err
	}

	chain := createChain(logMiddleware(*cnf), corsMiddleware(*cnf)).then(mux)

	return chain, nil
}

func corsMiddleware(c serviceconfig.ServiceConfig) middleware {
	return func(h http.Handler) http.Handler {
		return cors.New(cors.Options{
			AllowedOrigins:   c.CorsConfig.AllowedOrigins,
			AllowedMethods:   c.CorsConfig.AllowedMethods,
			AllowedHeaders:   c.CorsConfig.AllowedHeaders,
			AllowCredentials: c.CorsConfig.AllowCredentials,
			MaxAge:           c.CorsConfig.MaxAge,
			Debug:            c.CorsConfig.Debug,
		}).Handler(h)
	}
}

func logMiddleware(c serviceconfig.ServiceConfig) middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if c.LogRequests {
				glog.V(0).Infof("Request incoming: %s %s", r.Method, r.RequestURI)
				glog.V(1).Infof("Request dump: %+v", *r)
			}

			h.ServeHTTP(w, r)
		})
	}
}
