// Package bootstrap sticks the application parts together and runs it.
package bootstrap

import (
	"authz/api"
	"authz/api/grpc"
	"authz/api/http"
	"authz/application"
	"authz/domain/contracts"
	"sync"

	"github.com/golang/glog"
)

// Run configures and runs the actual bootstrap.
func Run(endpoint string, token string, store string, useTLS bool) {
	srv, webSrv := initialize(endpoint, token, store, useTLS)

	wait := sync.WaitGroup{}

	go func() {
		err := srv.Serve(&wait)
		if err != nil {
			glog.Fatal("Could not start grpc serving: ", err)
		}
	}()

	go func() {
		err := webSrv.Serve(&wait)
		if err != nil {
			glog.Fatal("Could not start http serving: ", err)
		}
	}()

	wait.Add(2)
	wait.Wait()
}

func initialize(endpoint string, token string, store string, useTLS bool) (*grpc.Server, *http.Server) {
	ar := getAccessRepository(store, endpoint, token, useTLS)
	sr := getSeatRepository(store, endpoint, token, useTLS, ar)
	pr := getPrincipalRepository(store)

	srvCfg := api.ServerConfig{ //TODO: Discuss config.
		GrpcPort:  "50051",
		HTTPPort:  "8081",
		HTTPSPort: "8443",
		TLSConfig: api.TLSConfig{
			CertPath: "/etc/tls/tls.crt",
			CertName: "",
			KeyPath:  "/etc/tls/tls.key",
			KeyName:  "",
		},
	}
	aas := application.NewAccessAppService(&ar, pr)
	sas := application.NewLicenseAppService(&ar, &sr, pr)

	srv := getGrpcServer(aas, sas, &srvCfg)

	webSrv := getHTTPServer(&srvCfg)
	webSrv.SetCheckRef(srv)
	webSrv.SetSeatRef(srv)

	return srv, webSrv
}

func getGrpcServer(aas *application.AccessAppService, sas *application.LicenseAppService, serverConfig *api.ServerConfig) *grpc.Server {
	srv, err := NewServerBuilder().
		WithAccessAppService(aas).
		WithLicenseAppService(sas).
		WithServerConfig(serverConfig).
		BuildGrpc()

	if err != nil {
		glog.Fatal("Could not initialize grpc server: ", err)
	}
	return srv
}

func getHTTPServer(serverConfig *api.ServerConfig) *http.Server {
	srv, err := NewServerBuilder().
		WithServerConfig(serverConfig).
		BuildHTTP()

	if err != nil {
		glog.Fatal("Could not initialize http server: ", err)
	}
	return srv
}

func getSeatRepository(store string, endpoint string, token string, useTLS bool, potentialStub interface{}) contracts.SeatLicenseRepository {
	b := NewSeatLicenseRepositoryBuilder()
	if stub, ok := potentialStub.(contracts.SeatLicenseRepository); ok {
		b.WithStub(stub)
	}

	return b.WithStore(store).WithConnectionInfo(endpoint, token, useTLS).Build()
}

func getAccessRepository(store string, endpoint string, token string, useTLS bool) contracts.AccessRepository {
	r, err := NewAccessRepositoryBuilder().
		WithImplementation(store).
		WithConnectionInfo(endpoint, token, useTLS).Build()

	if err != nil {
		glog.Fatal("Could not initialize access repository: ", err)
	}
	return r
}

func getPrincipalRepository(store string) contracts.PrincipalRepository {
	return NewPrincipalRepositoryBuilder().WithStore(store).Build()
}
