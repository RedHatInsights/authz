package app

import (
	"authz/domain/contracts"
	"authz/infrastructure/config"
	"authz/infrastructure/server"
	"sync"
)

// cfg holds the config from yaml. package private for now.
var cfg contracts.Config

// getConfig uses the interface to load the config based on the technical implementation "viper".
func getConfig() contracts.Config {
	cfg, err := config.NewBuilder().
		ConfigName("config").
		ConfigType("yaml").
		ConfigPaths(
			"app/config/", //TODO: configurable via flag. this only works when binary is in rootdir and code is there.
		).
		Defaults(map[string]interface{}{}).
		Options().
		Build()

	if err != nil {
		panic(err)
	}
	return cfg
}

// Run configures and runs the actual app. DEMO! switch the server from "echo" to "gin". see what happens.
func Run() {
	cfg = getConfig()

	srv := getServer()
	e := getAuthzEngine()
	e.NewConnection(cfg.GetString("app.engine.endpoint"), cfg.GetString("app.engine.token"))
	srv.SetEngine(e)
	wait := sync.WaitGroup{}

	delta := 1
	srvKind := cfg.GetString("app.server.kind")

	//2 chans for grpc gateway for http and grpc
	if srvKind == "grpc" {
		delta = 2
	}

	wait.Add(delta)

	go func() {
		err := srv.Serve(&wait, cfg.GetString("app.server.port"))
		if err != nil {
			panic(err)
		}
	}()

	if srvKind == "grpc" {
		webSrv, err := NewServerBuilder().WithFramework("grpcweb").Build()
		webSrv.SetEngine(e)
		webSrv.(*server.GrpcWebServer).SetHandler(srv.(*server.GrpcGatewayServer)) //ugly typeassertion hack.
		if err != nil {
			panic(err)
		}

		go func() {
			err := webSrv.Serve(&wait, cfg.GetString("app.server.grpc-web-httpPort"), cfg.GetString("app.server.grpc-web-httpsPort"))
			if err != nil {
				panic(err)
			}
		}()
	}

	wait.Wait()
}

func getServer() contracts.Server {
	srv, err := NewServerBuilder().WithFramework(cfg.GetString("app.server.kind")).Build()
	if err != nil {
		panic(err)
	}
	return srv
}

func getAuthzEngine() contracts.AuthzEngine {
	eng, err := NewAuthzEngineBuilder().WithEngine(cfg.GetString("app.engine.kind")).Build()
	if err != nil {
		panic(err)
	}
	return eng
}
