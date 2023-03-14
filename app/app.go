// Package app sticks the application parts together and runs it.
package app

import (
	apicontracts "authz/api/contracts"
	"authz/api/handler"
	"authz/api/server"
	appcfg "authz/app/config"
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
			configPath,
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
	srvCfg := parseServerConfig()
	ar := getAccessRepository()
	ar.NewConnection(
		Cfg.GetString("app.accessRepository.endpoint"),
		Cfg.GetString("app.accessRepository.token"))
	ph := initPermissionHandler(&ar)
	sh := initSeatHandler(&ar)
	srv := getServer(ph, sh, &srvCfg)

	wait := sync.WaitGroup{}

	delta := 1
	srvKind := Cfg.GetString("app.server.kind")

	//2 chans for grpc gateway for http and grpc
	if srvKind == "grpc" {
		delta = 2
	}

	wait.Add(delta)

	go func() {
		err := srv.Serve(&wait)
		if err != nil {
			glog.Fatal("Could not start serving: ", err)
		}
	}()

	if srvKind == "grpc" {
		webSrv, err := NewServerBuilder().
			WithFramework("grpcweb").
			WithPermissionHandler(ph).
			WithServerConfig(&srvCfg).
			Build()
		webSrv.(*server.GrpcWebServer).SetCheckRef(srv.(*server.GrpcGatewayServer)) //ugly typeassertion hack.
		if err != nil {
			glog.Fatal("Could not start serving grpc & web using grpc gateway: ", err)

		}

		go func() {
			err := webSrv.
				Serve(&wait)
			if err != nil {
				glog.Fatal("Could not start serving grpc webserver: ", err)

			}
		}()
	}

	wait.Wait()
}

func parseServerConfig() appcfg.ServerConfig {
	kind := Cfg.GetString("app.server.kind")
	return appcfg.ServerConfig{
		Kind:             kind,
		MainPort:         Cfg.GetString("app.server.port"),
		GrpcWebHttpPort:  Cfg.GetString("app.server.grpc-web-httpPort"),
		GrpcWebHttpsPort: Cfg.GetString("app.server.grpc-web-httpsPort"),
	}
}

func initPermissionHandler(ar *contracts.AccessRepository) *handler.PermissionHandler {
	permissionHandler := handler.PermissionHandler{}
	return permissionHandler.NewPermissionHandler(ar)
}

func initSeatHandler(ar *contracts.AccessRepository) *handler.SeatHandler {
	seatHandler := handler.SeatHandler{}
	return seatHandler.NewSeatHandler(ar)
}

func getServer(ph *handler.PermissionHandler, sh *handler.SeatHandler, serverConfig *appcfg.ServerConfig) apicontracts.Server {
	srv, err := NewServerBuilder().
		WithFramework(Cfg.GetString("app.server.kind")).
		WithPermissionHandler(ph).
		WithServerConfig(serverConfig).
		Build()

	if err != nil {
		glog.Fatal("Could not initialize server: ", err)
	}
	return srv
}

func getAccessRepository() contracts.AccessRepository {
	r, err := NewAccessRepositoryBuilder().
		WithImplementation(Cfg.GetString("app.accessRepository.kind")).
		Build()

	if err != nil {
		glog.Fatal("Could not initialize access repository: ", err)
	}
	return r
}
