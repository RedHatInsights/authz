// Package bootstrap sticks the application parts together and runs it.
package bootstrap

import (
	"authz/api/grpc"
	"authz/api/http"
	"authz/application"
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"strconv"
	"sync"

	"github.com/go-playground/validator/v10"

	"github.com/golang/glog"
)

// grpcServer is used as pointer to access the current server and re-initialize it, mainly for integration testing
var grpcServer *grpc.Server
var httpServer *http.Server
var waitForCompletion *sync.WaitGroup

// Cfg holds the parsed and validated service config.
var Cfg *serviceconfig.ServiceConfig

// getConfig loads the config based on the technical implementation "viper".
func getConfig(configPath string) serviceconfig.Config {
	cfg, err := NewConfigurationBuilder().
		ConfigName(configPath).
		ConfigType("yaml").
		ConfigPaths(
			".",
			"/",
		).
		Defaults(map[string]interface{}{}).
		Options().
		Build()

	if err != nil {
		glog.Fatalf("Could not initialize config: %v", err)
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

	grpcServer, httpServer = initialize(srvCfg)

	wait := &sync.WaitGroup{}
	wait.Add(2)

	go func() {
		err := grpcServer.Serve(wait)
		if err != nil {
			glog.Fatal("Could not start grpc serving: ", err)
		}
	}()

	go func() {
		err := httpServer.Serve(wait)
		if err != nil {
			glog.Fatal("Could not start http serving: ", err)
		}
	}()

	waitForCompletion = wait
	wait.Wait()
}

// Stop shuts down the server endpoints and performs teardown functions and blocks until completed
func Stop() {
	glog.Info("Attempting graceful shutdown...")
	err := httpServer.Stop() //Stop accepting HTTP requests and shut it down
	if err != nil {
		glog.Errorf("Error stopping HTTP server: %s", err)
	}
	grpcServer.Stop() //Stop accepting gRPC/adapted HTTP requests after shutting down HTTP

	waitForCompletion.Wait()

	grpcServer = nil
	httpServer = nil
	waitForCompletion = nil
}

func initialize(srvCfg serviceconfig.ServiceConfig) (*grpc.Server, *http.Server) {

	ar := initAccessRepository(&srvCfg)
	sr := initSeatRepository(&srvCfg, ar)
	pr := initPrincipalRepository(srvCfg.StoreConfig.Kind)

	aas := application.NewAccessAppService(&ar, pr)
	sas := application.NewLicenseAppService(&ar, &sr, pr)

	srv := initGrpcServer(aas, sas, &srvCfg)

	webSrv := initHTTPServer(&srvCfg)
	webSrv.SetCheckRef(srv)
	webSrv.SetSeatRef(srv)
	grpcServer = srv
	return srv, webSrv
}

// initGrpcServer initializes a new grpc server struct
func initGrpcServer(aas *application.AccessAppService, sas *application.LicenseAppService, serviceConfig *serviceconfig.ServiceConfig) *grpc.Server {
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

// initHttpServer initializes new http server struct
func initHTTPServer(serviceConfig *serviceconfig.ServiceConfig) *http.Server {
	srv, err := NewServerBuilder().
		WithServiceConfig(serviceConfig).
		BuildHTTP()

	if err != nil {
		glog.Fatal("Could not initialize http server: ", err)
	}
	return srv
}

func initSeatRepository(config *serviceconfig.ServiceConfig, potentialStub interface{}) contracts.SeatLicenseRepository {
	b := NewSeatLicenseRepositoryBuilder()
	if stub, ok := potentialStub.(contracts.SeatLicenseRepository); ok {
		b.WithStub(stub)
	}

	return b.WithConfig(config).Build()
}

func initAccessRepository(config *serviceconfig.ServiceConfig) contracts.AccessRepository {
	r, err := NewAccessRepositoryBuilder().
		WithConfig(config).Build()

	if err != nil {
		glog.Fatal("Could not initialize access repository: ", err)
	}
	return r
}

func initPrincipalRepository(store string) contracts.PrincipalRepository {
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
		CorsConfig: serviceconfig.CorsConfig{
			AllowedMethods:   cfg.GetStringSlice("app.cors.allowedMethods"),
			AllowedHeaders:   cfg.GetStringSlice("app.cors.allowedHeaders"),
			AllowCredentials: cfg.GetBool("app.cors.allowCredentials"),
			MaxAge:           cfg.GetInt("app.cors.maxAge"),
			Debug:            cfg.GetBool("app.cors.debug"),
		},
		AuthConfig: serviceconfig.AuthConfig{
			DiscoveryEndpoint: cfg.GetString("app.auth.discoveryEndpoint"),
			Audience:          cfg.GetString("app.auth.audience"),
			RequiredScope:     cfg.GetString("app.auth.requiredScope"),
		},
	}
}
