// Package app sticks the application parts together and runs it.
package app

import (
	apicontracts "authz/api/contracts"
	"authz/api/server"
	appcontracts "authz/app/contracts"
	"authz/domain/contracts"
	"authz/infrastructure/config"
	"sync"
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
		panic(err)
	}
	return cfg
}

// Run configures and runs the actual app.
func Run(configPath string) {
	Cfg = getConfig(configPath)

	srv := getServer()
	e := getAccessRepository()
	e.NewConnection(Cfg.GetString("app.accessRepository.endpoint"), Cfg.GetString("app.accessRepository.token"))
	srv.SetAccessRepository(e)
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
			panic(err)
		}
	}()

	if srvKind == "grpc" {
		webSrv, err := NewServerBuilder().WithFramework("grpcweb").Build()
		webSrv.SetAccessRepository(e)
		webSrv.(*server.GrpcWebServer).SetHandler(srv.(*server.GrpcGatewayServer)) //ugly typeassertion hack.
		if err != nil {
			panic(err)
		}

		go func() {
			err := webSrv.Serve(&wait, Cfg.GetString("app.server.grpc-web-httpPort"), Cfg.GetString("app.server.grpc-web-httpsPort"))
			if err != nil {
				panic(err)
			}
		}()
	}

	wait.Wait()
}

func getServer() apicontracts.Server {
	srv, err := NewServerBuilder().WithFramework(Cfg.GetString("app.server.kind")).Build()
	if err != nil {
		panic(err)
	}
	return srv
}

func getAccessRepository() contracts.AccessRepository {
	r, err := NewAccessRepositoryBuilder().WithImplementation(Cfg.GetString("app.accessRepository.kind")).Build()
	if err != nil {
		panic(err)
	}
	return r
}
