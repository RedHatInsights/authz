// Package bootstrap sticks the application parts together and runs it.
package bootstrap

import (
	"authz/api/grpc"
	"authz/api/http"
	"authz/application"
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"authz/infrastructure/config"
	"sync"

	"github.com/golang/glog"
)

// Cfg holds the config from yaml.
var Cfg serviceconfig.Config

// getConfig loads the config based on the technical implementation "viper".
func getConfig(configPath string) serviceconfig.Config {

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
	srvCfg := parseServiceConfig()
	srv, webSrv := initialize(srvCfg)

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

func initialize(srvCfg serviceconfig.ServiceConfig) (*grpc.Server, *http.Server) {

	ar := getAccessRepository(&srvCfg)
	sr := getSeatRepository(&srvCfg, ar)
	pr := getPrincipalRepository(srvCfg.StoreConfig.Kind)

	aas := application.NewAccessAppService(&ar, pr)
	sas := application.NewLicenseAppService(&ar, &sr, pr)

	srv := getGrpcServer(aas, sas, &srvCfg)

	webSrv := getHTTPServer(&srvCfg)
	webSrv.SetCheckRef(srv)
	webSrv.SetSeatRef(srv)

	return srv, webSrv
}

func getGrpcServer(aas *application.AccessAppService, sas *application.LicenseAppService, serverConfig *serviceconfig.ServiceConfig) *grpc.Server {
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

func getHTTPServer(serverConfig *serviceconfig.ServiceConfig) *http.Server {
	srv, err := NewServerBuilder().
		WithServerConfig(serverConfig).
		BuildHTTP()

	if err != nil {
		glog.Fatal("Could not initialize http server: ", err)
	}
	return srv
}

func getSeatRepository(config *serviceconfig.ServiceConfig, potentialStub interface{}) contracts.SeatLicenseRepository {
	b := NewSeatLicenseRepositoryBuilder()
	if stub, ok := potentialStub.(contracts.SeatLicenseRepository); ok {
		b.WithStub(stub)
	}

	return b.WithConfig(config).Build()
}

func getAccessRepository(config *serviceconfig.ServiceConfig) contracts.AccessRepository {
	r, err := NewAccessRepositoryBuilder().
		WithConfig(config).Build()

	if err != nil {
		glog.Fatal("Could not initialize access repository: ", err)
	}
	return r
}

func getPrincipalRepository(store string) contracts.PrincipalRepository {
	return NewPrincipalRepositoryBuilder().WithStore(store).Build()
}

func parseServiceConfig() serviceconfig.ServiceConfig {
	return serviceconfig.ServiceConfig{
		GrpcPort:  Cfg.GetString("app.server.grpcPort"),
		HTTPPort:  Cfg.GetString("app.server.httpPort"),
		HTTPSPort: Cfg.GetString("app.server.httpsPort"),
		TLSConfig: serviceconfig.TLSConfig{
			CertFile: Cfg.GetString("app.tls.certFile"),
			KeyFile:  Cfg.GetString("app.tls.keyFile"),
		},
		StoreConfig: serviceconfig.StoreConfig{
			Kind:      Cfg.GetString("app.store.kind"),
			Endpoint:  Cfg.GetString("app.store.endpoint"),
			AuthToken: Cfg.GetString("app.store.token"),
			UseTLS:    Cfg.GetBool("app.store.useTLS"),
		},
		CorsConfig: serviceconfig.CorsConfig{
			AllowedMethods:   Cfg.GetStringSlice("app.cors.allowedMethods"),
			AllowedHeaders:   Cfg.GetStringSlice("app.cors.allowedHeaders"),
			AllowCredentials: Cfg.GetBool("app.cors.allowCredentials"),
			MaxAge:           Cfg.GetInt("app.cors.maxAge"),
			Debug:            Cfg.GetBool("app.cors.debug"),
		},
	}
}
