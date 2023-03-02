package host

import (
	"context"
	"net/http"
	"os"
	"sync"

	core "authz/api/gen/v1"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// Web is delivery adapter for HTTP
type Web struct {
	services Services
}

// Host exposes an HTTP endpoint and blocks until processing ends, at which point the waitgroup is signalled. This should be run as a goroutine.
func (web Web) Host(wait *sync.WaitGroup, handler core.CheckPermissionServer) {
	defer wait.Done()

	mux, err := createMultiplexer(handler)
	if err != nil {
		glog.Errorf("Error creating multiplexer: %s", err)
		return
	}

	if _, err = os.Stat("/etc/tls/tls.crt"); err == nil {
		if _, err := os.Stat("/etc/tls/tls.key"); err == nil { //Cert and key exists start server in HTTPS mode
			glog.Info("TLS cert and Key found  - Starting server in secure HTTPs mode")

			err = http.ListenAndServeTLS(":8443", "/etc/tls/tls.crt", "/etc/tls/tls.key", mux)
			if err != nil {
				glog.Errorf("Error hosting TLS service: %s", err)
				return
			}
		}
	} else { // For all cases of error - we start a plain HTTP server
		glog.Info("TLS cert or Key not found  - Starting server in unsercure plain HTTP mode")
		err = http.ListenAndServe(":8080", mux)

		if err != nil {
			glog.Errorf("Error hosting insecure service: %s", err)
			return
		}
	}
}

// NewWeb Constructs a new instance of the Web delivery adapter
func NewWeb(services Services) Web {
	return Web{services: services}
}

func createMultiplexer(handler core.CheckPermissionServer) (*runtime.ServeMux, error) {
	mux := runtime.NewServeMux()

	if err := core.RegisterCheckPermissionHandlerServer(context.Background(), mux, handler); err != nil {
		return nil, err
	}

	return mux, nil
}
