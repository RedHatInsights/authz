package app

import (
	"authz/domain/contracts"
	"authz/infrastructure/config"
	"authz/infrastructure/server"
	"fmt"
	"sync"
)

// getConfig uses the interface to load the config based on the technical implementation "viper".
func getConfig() contracts.Config {
	cfg, err := config.NewBuilder().
		ConfigName("viperexampleconfig").
		ConfigType("yaml").
		ConfigPaths(
			"app/exampleconfig/", //TODO: configurable via flag. this only works when binary is in rootdir and code is there.
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
	cfg := getConfig()
	fmt.Println(cfg.GetAll())
	fmt.Println(cfg.GetBool("example.boolVal"))
	fmt.Println(cfg.GetString("example.stringVal"))
	fmt.Println(cfg.GetStringSlice("example.list"))

	srv := getServer()
	e := getAuthzEngine()
	srv.SetEngine(e)
	wait := sync.WaitGroup{}

	delta := 1
	if srv.GetName() == "grpc" { //2 chans for grpc
		delta = 2
	}
	wait.Add(delta)

	go func() {
		err := srv.Serve("50051", &wait)
		if err != nil {
			panic(err)
		}
	}()

	if srv.GetName() == "grpc" {
		webSrv, err := NewServerBuilder().WithFramework("grpcweb").Build()
		webSrv.SetEngine(e)
		webSrv.(*server.GrpcWebServer).SetHandler(srv.(*server.GrpcGatewayServer)) //ugly typeassertion hack. :)
		if err != nil {
			panic(err)
		}

		go func() {
			err := webSrv.Serve("8080", &wait)
			if err != nil {
				panic(err)
			}
		}()
	}

	wait.Wait()
}

func getServer() contracts.Server {
	srv, err := NewServerBuilder().WithFramework("grpc").Build()
	if err != nil {
		panic(err)
	}
	return srv
}

func getAuthzEngine() contracts.AuthzEngine {
	eng, err := NewAuthzEngineBuilder().WithEngine("stub").Build()
	if err != nil {
		panic(err)
	}
	return eng
}
