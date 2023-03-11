package server

import (
	core "authz/api/gen/v1alpha"
	contracts2 "authz/app/contracts"
	"authz/domain/contracts"
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// GrpcWebServer serves a HTTP api based on the generated grpc gateway code
type GrpcWebServer struct {
	Engine  contracts.AuthzEngine
	Handler core.CheckPermissionServer
}

// Serve starts serving
func (w *GrpcWebServer) Serve(wait *sync.WaitGroup, ports ...string) error {
	defer wait.Done()

	mux, err := createMultiplexer(w.Handler)
	if err != nil {
		glog.Errorf("Error creating multiplexer: %s", err)
		return err
	}

	if _, err = os.Stat("/etc/tls/tls.crt"); err == nil {
		if _, err := os.Stat("/etc/tls/tls.key"); err == nil { //Cert and key exists start server in HTTPS mode
			glog.Infof("TLS cert and Key found  - Starting server in secure HTTPS mode on port %s", ports[1])

			err = http.ListenAndServeTLS(":"+ports[1], "/etc/tls/tls.crt", "/etc/tls/tls.key", mux)
			if err != nil {
				glog.Errorf("Error hosting TLS service: %s", err)
				return err
			}
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Infof("TLS cert or Key not found  - Starting server in insecure plain HTTP mode on Port %s", ports[0])
		err = http.ListenAndServe(":"+ports[0], mux)

		if err != nil {
			glog.Errorf("Error hosting insecure service: %s", err)
			return err
		}
	}
	return nil
}

// SetEngine sets the authzengine
func (w *GrpcWebServer) SetEngine(eng contracts.AuthzEngine) {
	w.Engine = eng
}

// SetHandler sets the handler for reference to grpc
func (w *GrpcWebServer) SetHandler(h core.CheckPermissionServer) {
	w.Handler = h
}

// NewServer creates a new Server object to use.
func (w *GrpcWebServer) NewServer() contracts2.Server {
	return &GrpcWebServer{}
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
