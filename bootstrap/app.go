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

// grpcServer is used as pointer to access the current server and re-initialize it, mainly for integration testing
var grpcServer *grpc.Server

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
		StoreConfig: api.StoreConfig{
			Store:     store,
			Endpoint:  endpoint,
			AuthToken: token,
			UseTLS:    useTLS,
		},
	}

	ar := initAccessRepository(&srvCfg)
	sr := initSeatRepository(&srvCfg, ar)
	pr := initPrincipalRepository(store)

	aas := application.NewAccessAppService(&ar, pr)
	sas := application.NewLicenseAppService(&ar, &sr, pr)

	srv := initGrpcServer(aas, sas, &srvCfg)

	webSrv := initHTTPServer(&srvCfg)
	webSrv.SetCheckRef(srv)
	webSrv.SetSeatRef(srv)
	grpcServer = srv
	return srv, webSrv
}

// initGrpcServer initializes a new grpc server struct
func initGrpcServer(aas *application.AccessAppService, sas *application.LicenseAppService, serverConfig *api.ServerConfig) *grpc.Server {
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

// initHttpServer initializes new http server struct
func initHTTPServer(serverConfig *api.ServerConfig) *http.Server {
	srv, err := NewServerBuilder().
		WithServerConfig(serverConfig).
		BuildHTTP()

	if err != nil {
		glog.Fatal("Could not initialize http server: ", err)
	}
	return srv
}

// getGrpcServer returns the pointer to the current running server struct. Mainly used for re-initializing services in tests.
func getGrpcServer() *grpc.Server {
	return grpcServer
}

func initSeatRepository(config *api.ServerConfig, potentialStub interface{}) contracts.SeatLicenseRepository {
	b := NewSeatLicenseRepositoryBuilder()
	if stub, ok := potentialStub.(contracts.SeatLicenseRepository); ok {
		b.WithStub(stub)
	}

	return b.WithConfig(config).Build()
}

func initAccessRepository(config *api.ServerConfig) contracts.AccessRepository {
	r, err := NewAccessRepositoryBuilder().
		WithConfig(config).Build()

	if err != nil {
		glog.Fatal("Could not initialize access repository: ", err)
	}
	return r
}

func initPrincipalRepository(store string) contracts.PrincipalRepository {
	return NewPrincipalRepositoryBuilder().WithStore(store).Build()
}
