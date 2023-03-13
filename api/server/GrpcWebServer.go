package server

import (
	apicontracts "authz/api/contracts"
	core "authz/api/gen/v1alpha"
	"authz/api/handler"
	"authz/app/config"
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// GrpcWebServer serves a HTTP api based on the generated grpc gateway code
type GrpcWebServer struct {
	ServerConfig *config.ServerConfig
	Handler      core.CheckPermissionServer
}

// Serve starts serving
func (w *GrpcWebServer) Serve(wait *sync.WaitGroup) error {
	defer wait.Done()

	mux, err := createMultiplexer(w.Handler)
	if err != nil {
		glog.Errorf("Error creating multiplexer: %s", err)
		return err
	}

	if _, err = os.Stat("/etc/tls/tls.crt"); err == nil {
		if _, err := os.Stat("/etc/tls/tls.key"); err == nil { //Cert and key exists start server in HTTPS mode
			glog.Infof("TLS cert and Key found  - Starting server in secure HTTPS mode on port %s",
				w.ServerConfig.GrpcWebHttpsPort)

			err = http.ListenAndServeTLS(
				":"+w.ServerConfig.GrpcWebHttpsPort,
				"/etc/tls/tls.crt", //TODO: Needs sanity checking and get from config.
				"/etc/tls/tls.key", mux)
			if err != nil {
				glog.Errorf("Error hosting TLS service: %s", err)
				return err
			}
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Infof("TLS cert or Key not found  - Starting server in insecure plain HTTP mode on Port %s",
			w.ServerConfig.GrpcWebHttpPort)
		err = http.ListenAndServe(":"+w.ServerConfig.GrpcWebHttpPort, mux)

		if err != nil {
			glog.Errorf("Error hosting insecure service: %s", err)
			return err
		}
	}
	return nil
}

// SetHandler sets the handler for reference to grpc
func (w *GrpcWebServer) SetHandler(h core.CheckPermissionServer) {
	w.Handler = h
}

// NewServer creates a new Server object to use.
func (w *GrpcWebServer) NewServer(_ handler.PermissionHandler, c config.ServerConfig) apicontracts.Server {
	return &GrpcWebServer{
		ServerConfig: &c,
	}
}

// GetName returns the Name of the impl
func (w *GrpcWebServer) GetName() string {
	return "grpcweb"
}

func createMultiplexer(handler core.CheckPermissionServer) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux()

	if err := core.RegisterCheckPermissionHandlerServer(context.Background(), mux, handler); err != nil {
		return nil, err
	}

	return mux, nil
}
