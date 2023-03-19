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
func Run(endpoint string, token string, store string) {
	ar := getAccessRepository(store)
	ar.NewConnection(
		endpoint,
		token,
		false)

	srvCfg := api.ServerConfig{ //TODO: Discuss config.
		GrpcPort:  "50051",
		HTTPPort:  "8080",
		HTTPSPort: "8443",
		TLSConfig: api.TLSConfig{},
	}
	aas := initAccessAppService(&ar)
	sas := initSeatAppService(&ar)

	wait := sync.WaitGroup{}

	wait.Add(2)

	srv := getGrpcServer(aas, sas, &srvCfg)

	go func() {
		err := srv.Serve(&wait)
		if err != nil {
			glog.Fatal("Could not start grpc serving: ", err)
		}
	}()

	webSrv := getHTTPServer(&srvCfg)
	webSrv.SetCheckRef(srv)

	go func() {
		err := webSrv.
			Serve(&wait)
		if err != nil {
			glog.Fatal("Could not start http serving: ", err)

		}
	}()

	wait.Wait()
}

func initAccessAppService(ar *contracts.AccessRepository) *application.AccessAppService {
	permissionHandler := application.AccessAppService{}
	return permissionHandler.NewPermissionHandler(ar)
}

func initSeatAppService(ar *contracts.AccessRepository) *application.SeatAppService {
	seatHandler := application.SeatAppService{}
	return seatHandler.NewSeatAppService(ar)
}

func getGrpcServer(aas *application.AccessAppService, sas *application.SeatAppService, serverConfig *api.ServerConfig) *grpc.Server {
	srv, err := NewServerBuilder().
		WithAccessAppService(aas).
		WithSeatAppService(sas).
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

func getAccessRepository(store string) contracts.AccessRepository {
	r, err := NewAccessRepositoryBuilder().
		WithImplementation(store).Build()

	if err != nil {
		glog.Fatal("Could not initialize access repository: ", err)
	}
	return r
}
