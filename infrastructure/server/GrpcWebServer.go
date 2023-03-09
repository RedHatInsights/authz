package server

import (
	core "authz/api/gen/v1alpha"
	"authz/domain/contracts"
	"context"
	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"net/http"
	"os"
	"sync"
)

// GrpcWebServer serves a HTTP api based on the generated grpc gateway code
type GrpcWebServer struct {
	Engine  contracts.AuthzEngine
	Handler core.CheckPermissionServer
}

func (w *GrpcWebServer) Serve(host string, wait *sync.WaitGroup) error {
	defer wait.Done()

	mux, err := createMultiplexer(w.Handler)
	if err != nil {
		glog.Errorf("Error creating multiplexer: %s", err)
		return err
	}

	if _, err = os.Stat("/etc/tls/tls.crt"); err == nil {
		if _, err := os.Stat("/etc/tls/tls.key"); err == nil { //Cert and key exists start server in HTTPS mode
			glog.Info("TLS cert and Key found  - Starting server in secure HTTPs mode")

			err = http.ListenAndServeTLS(":8443", "/etc/tls/tls.crt", "/etc/tls/tls.key", mux)
			if err != nil {
				glog.Errorf("Error hosting TLS service: %s", err)
				return err
			}
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Info("TLS cert or Key not found  - Starting server in unsercure plain HTTP mode")
		err = http.ListenAndServe(":"+host, mux)

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
func (w *GrpcWebServer) NewServer() contracts.Server {
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
