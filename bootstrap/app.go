// Package bootstrap sticks the application parts together and runs it.
package bootstrap

import (
	"authz/api/grpc"
	"authz/api/http"
	"authz/application"
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"authz/infrastructure/config"
	"strconv"
	"sync"

	"github.com/go-playground/validator/v10"

	"github.com/golang/glog"
)

// Cfg holds the parsed and validated service config.
var Cfg *serviceconfig.ServiceConfig

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
	configProvider := getConfig(configPath)
	srvCfg := parseServiceConfig(configProvider)
	vl := validator.New()
	err := vl.Struct(srvCfg)

	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			glog.Errorf("Error in configuration: %v", e)
		}
		glog.Fatal("Can not start service with wrong configuration.")
	}
	//set global Config now that it is parsed and validated.
	Cfg = &srvCfg

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

func getGrpcServer(aas *application.AccessAppService, sas *application.LicenseAppService, serviceConfig *serviceconfig.ServiceConfig) *grpc.Server {
	srv, err := NewServerBuilder().
		WithAccessAppService(aas).
		WithLicenseAppService(sas).
		WithServiceConfig(serviceConfig).
		BuildGrpc()

	if err != nil {
		glog.Fatal("Could not initialize grpc server: ", err)
	}
	return srv
}

func getHTTPServer(serviceConfig *serviceconfig.ServiceConfig) *http.Server {
	srv, err := NewServerBuilder().
		WithServiceConfig(serviceConfig).
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

func parseServiceConfig(cfg serviceconfig.Config) serviceconfig.ServiceConfig {
	return serviceconfig.ServiceConfig{
		GrpcPort:     cfg.GetInt("app.server.grpcPort"),
		GrpcPortStr:  strconv.Itoa(cfg.GetInt("app.server.grpcPort")),
		HTTPPort:     cfg.GetInt("app.server.httpPort"),
		HTTPPortStr:  strconv.Itoa(cfg.GetInt("app.server.httpPort")),
		HTTPSPort:    cfg.GetInt("app.server.httpsPort"),
		HTTPSPortStr: strconv.Itoa(cfg.GetInt("app.server.httpsPort")),
		LogRequests:  cfg.GetBool("app.server.logRequests"),
		TLSConfig: serviceconfig.TLSConfig{
			CertFile: cfg.GetString("app.tls.certFile"),
			KeyFile:  cfg.GetString("app.tls.keyFile"),
		},
		StoreConfig: serviceconfig.StoreConfig{
			Kind:      cfg.GetString("app.store.kind"),
			Endpoint:  cfg.GetString("app.store.endpoint"),
			AuthToken: cfg.GetString("app.store.token"),
			UseTLS:    cfg.GetBool("app.store.useTLS"),
		},
		CorsConfig: serviceconfig.CorsConfig{ //TODO: see how to integrate in middlewares.
			AllowedMethods:   cfg.GetStringSlice("app.cors.allowedMethods"),
			AllowedHeaders:   cfg.GetStringSlice("app.cors.allowedHeaders"),
			AllowCredentials: cfg.GetBool("app.cors.allowCredentials"),
			MaxAge:           cfg.GetInt("app.cors.maxAge"),
			Debug:            cfg.GetBool("app.cors.debug"),
		},
	}
}
