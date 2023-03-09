package app

import (
	"authz/domain/contracts"
	"authz/domain/handler"
	"authz/infrastructure/config"
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

func getServer() contracts.Server {
	srv, err := NewAppBuilder().WithFramework("echo").WithEngine("stub").Build()
	if err != nil {
		panic(err)
	}
	return srv
}

// Run configures and runs the actual app. DEMO! switch the server from "echo" to "gin". see what happens.
func Run() {
	cfg := getConfig()
	fmt.Println(cfg.GetAll())
	fmt.Println(cfg.GetBool("example.boolVal"))
	fmt.Println(cfg.GetString("example.stringVal"))
	fmt.Println(cfg.GetStringSlice("example.list"))

	srv := getServer()
	wait := sync.WaitGroup{}
	wait.Add(1)

	err := srv.Serve("8080", handler.GetHello, &wait) // port could e.g. be derived from config ;)
	if err != nil {
		panic(err)
	}
	wait.Wait()
}
