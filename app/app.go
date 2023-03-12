// Package app sticks the application parts together and runs it.
package app

import (
	apicontracts "authz/api/contracts"
	"authz/api/handler"
	"authz/api/server"
	appcontracts "authz/app/contracts"
	"authz/domain/contracts"
	"authz/infrastructure/config"
	"sync"

	"github.com/golang/glog"
)

// Cfg holds the config from yaml. package private for now.
var Cfg appcontracts.Config

// getConfig uses the interface to load the config based on the technical implementation "viper".
func getConfig(configPath string) appcontracts.Config {
	cfg, err := config.NewBuilder().
		ConfigName("config").
		ConfigType("yaml").
		ConfigPaths(
			configPath, //TODO: configurable via flag. this only works when binary is in rootdir and code is there.
		).
		Defaults(map[string]interface{}{}).
		Options().
		Build()

	if err != nil {
		glog.Fatal("Could not initialize config: ", err)
	}
	return cfg
}

// Run configures and runs the actual app.
func Run(configPath string) {
	Cfg = getConfig(configPath)
	ar := getAccessRepository()
	ar.NewConnection(Cfg.GetString("app.accessRepository.endpoint"), Cfg.GetString("app.accessRepository.token"))
	ph := initPermissionHandler(&ar) //TODO: discuss and think about it.

	srv := getServer(ph)

	wait := sync.WaitGroup{}

	delta := 1
	srvKind := Cfg.GetString("app.server.kind")

	//2 chans for grpc gateway for http and grpc
	if srvKind == "grpc" {
		delta = 2
	}

	wait.Add(delta)

	go func() {
		err := srv.Serve(&wait, Cfg.GetString("app.server.port"))
		if err != nil {
			glog.Fatal("Could not start serving: ", err)
		}
	}()

	if srvKind == "grpc" {
		webSrv, err := NewServerBuilder().WithFramework("grpcweb").WithPermissionHandler(ph).Build()
		webSrv.(*server.GrpcWebServer).SetHandler(srv.(*server.GrpcGatewayServer)) //ugly typeassertion hack.
		if err != nil {
			glog.Fatal("Could not start serving grpc & web using grpc gateway: ", err)

		}

		go func() {
			err := webSrv.Serve(&wait, Cfg.GetString("app.server.grpc-web-httpPort"), Cfg.GetString("app.server.grpc-web-httpsPort"))
			if err != nil {
				glog.Fatal("Could not start serving grpc webserver: ", err)

			}
		}()
	}

	wait.Wait()
}

// init permissionhandler with repo.
func initPermissionHandler(ar *contracts.AccessRepository) *handler.PermissionHandler {
	permissionHandler := handler.PermissionHandler{}
	return permissionHandler.NewPermissionHandler(ar)
}

func getServer(h *handler.PermissionHandler) apicontracts.Server {
	srv, err := NewServerBuilder().WithFramework(Cfg.GetString("app.server.kind")).WithPermissionHandler(h).Build()

	if err != nil {
		glog.Fatal("Could not initialize server: ", err)
	}
	return srv
}

func getAccessRepository() contracts.AccessRepository {
	r, err := NewAccessRepositoryBuilder().WithImplementation(Cfg.GetString("app.accessRepository.kind")).Build()
	if err != nil {
		glog.Fatal("Could not initialize access repository: ", err)
	}
	return r
}
