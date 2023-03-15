// Package bootstrap sticks the application parts together and runs it.
package bootstrap

import (
	"authz/api/grpc"
	"authz/api/http"
	"authz/application"
	appcfg "authz/bootstrap/config"
	appcontracts "authz/bootstrap/contracts"
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

// Run configures and runs the actual bootstrap.
func Run(configPath string) {
	Cfg = getConfig(configPath)
	srvCfg := parseServerConfig()
	ar := getAccessRepository()
	ar.NewConnection(
		Cfg.GetString("bootstrap.accessRepository.endpoint"),
		Cfg.GetString("bootstrap.accessRepository.token"))
	aas := initAccessAppService(&ar)
	sas := initSeatAppService(&ar)

	wait := sync.WaitGroup{}

	delta := 2

	wait.Add(delta)

	srv := getGrpcServer(aas, sas, &srvCfg)

	go func() {
		err := srv.Serve(&wait)
		if err != nil {
			glog.Fatal("Could not start grpc serving: ", err)
		}
	}()

	webSrv := getHttpServer(&srvCfg)
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

func parseServerConfig() appcfg.ServerConfig {
	kind := Cfg.GetString("bootstrap.server.kind")
	return appcfg.ServerConfig{
		Kind:             kind,
		MainPort:         Cfg.GetString("bootstrap.server.port"),
		GrpcWebHttpPort:  Cfg.GetString("bootstrap.server.grpc-web-httpPort"),
		GrpcWebHttpsPort: Cfg.GetString("bootstrap.server.grpc-web-httpsPort"),
	}
}

func initAccessAppService(ar *contracts.AccessRepository) *application.AccessAppService {
	permissionHandler := application.AccessAppService{}
	return permissionHandler.NewPermissionHandler(ar)
}

func initSeatAppService(ar *contracts.AccessRepository) *application.SeatAppService {
	seatHandler := application.SeatAppService{}
	return seatHandler.NewSeatAppService(ar)
}

func getGrpcServer(aas *application.AccessAppService, sas *application.SeatAppService, serverConfig *appcfg.ServerConfig) *grpc.Server {
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

func getHttpServer(serverConfig *appcfg.ServerConfig) *http.Server {
	srv, err := NewServerBuilder().
		WithServerConfig(serverConfig).
		BuildHttp()

	if err != nil {
		glog.Fatal("Could not initialize http server: ", err)
	}
	return srv
}

func getAccessRepository() contracts.AccessRepository {
	r, err := NewAccessRepositoryBuilder().
		WithImplementation(Cfg.GetString("bootstrap.accessRepository.kind")).Build()

	if err != nil {
		glog.Fatal("Could not initialize access repository: ", err)
	}
	return r
}
