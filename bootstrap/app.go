// Package bootstrap sticks the application parts together and runs it.
package bootstrap

import (
	"authz/api/grpc"
	"authz/api/http"
	"authz/application"
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"authz/infrastructure/repository/messaging"
	"sync"

	"github.com/go-playground/validator/v10"

	"github.com/golang/glog"
)

// grpcServer is used as pointer to access the current server and re-initialize it, mainly for integration testing
var grpcServer *grpc.Server
var httpServer *http.Server
var waitForCompletion *sync.WaitGroup

// getConfig loads the config based on the technical implementation "viper".
func getConfig(configPath string) (serviceconfig.ServiceConfig, error) {
	cfg, err := NewConfigurationBuilder().
		ConfigFilePath(configPath).
		Defaults(serviceconfig.ServiceConfig{
			GrpcPort:     50051,
			GrpcPortStr:  "50051",
			HTTPPort:     8080,
			HTTPPortStr:  "8080",
			HTTPSPort:    8443,
			HTTPSPortStr: "8443",
			CorsConfig: serviceconfig.CorsConfig{
				AllowedMethods: []string{"HEAD", "GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders: []string{
					"Accept",
					"ResponseType",
					"Content-Length",
					"Accept-Encoding",
					"Accept-Language",
					"Authorization",
					"Content-Type",
					"User-Agent"},
				AllowCredentials: false,
				MaxAge:           300,
				Debug:            false,
			},
			StoreConfig: serviceconfig.StoreConfig{
				Kind:   "spicedb",
				UseTLS: true,
			},
			TLSConfig: serviceconfig.TLSConfig{
				CertFile: "/etc/tls/tls.crt",
				KeyFile:  "/etc/tls/tls.key ",
			},
			LogRequests: false,
		}).
		Build()

	if err != nil {
		glog.Fatalf("Could not initialize config: %v", err)
	}

	return cfg.Load()
}

var umbRepo contracts.MessageBusRepository

// Run configures and runs the actual bootstrap.
func Run(configPath string) {
	srvCfg, err := getConfig(configPath)
	if err != nil {
		glog.Errorf("Unable to load configuration: %v", err)
		return
	}
	vl := validator.New()
	err = vl.Struct(srvCfg)

	if err != nil {
		for _, e := range err.(validator.ValidationErrors) {
			glog.Errorf("Error in configuration: %v", e)
		}
		glog.Error("Can not start service with invalid configuration.")
		return
	}

	grpcServer, httpServer, err = initialize(srvCfg)
	if err != nil {
		glog.Error("Error in service initialization: ", err)
		return
	}

	wait := &sync.WaitGroup{}
	wait.Add(2)

	umbCfg := srvCfg.UMBConfig
	if umbCfg.Enabled {
		umb := messaging.NewUMBMessageBusRepository(umbCfg)
		evts, err := umb.Connect()
		if err != nil {
			glog.Errorf("Failed to connect to umb: %v", err)
		} else {
			umbRepo = umb
			glog.Info("Connected to UMB.")
			go func(evts contracts.UserEvents) {
				ok := true
				var evt contracts.SubjectAddOrUpdateEvent

				for ok {
					select {
					case evt, ok = <-evts.SubjectChanges:
						glog.Infof("Subject event from UMB connection: %+v", evt)
					case err, ok = <-evts.Errors:
						glog.Errorf("Error from UMB connection: %v", err)
					}
				}
			}(evts)
		}
	} else {
		glog.Info("Skipping UMB connectivity - not enabled.")
	}

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

	if umbRepo != nil {
		umbRepo.Disconnect()
	}

	grpcServer = nil
	httpServer = nil
	waitForCompletion = nil
}

func initialize(srvCfg serviceconfig.ServiceConfig) (*grpc.Server, *http.Server, error) {

	ar, err := initAccessRepository(&srvCfg)
	if err != nil {
		return nil, nil, err
	}

	// TODO: The init and builder functions need to be tidied up for SubjectRepository and Principal repository
	// TODO: The casting acrobatics between the two also needs fixing with an intersection type? or dependency injection
	stubPr := initStubPrincipalRepository(srvCfg.StoreConfig.Kind)
	pr := stubPr

	subr, err := initCombinedUserServiceSubjectPrincipalRepository(&srvCfg)
	// TODO: remove fallback when all is tested
	if err != nil {
		glog.Errorf("failed to initialise UserService SubjectRepository (falling back to stub PrincipalRepository for PrincipalRepository and SubjectRepository): %v", err)
		subr = stubPr.(contracts.SubjectRepository)
	} else {
		pr = subr.(contracts.PrincipalRepository)
	}

	sr, err := initSeatRepository(&srvCfg)
	if err != nil {
		return nil, nil, err
	}
	or := sr.(contracts.OrganizationRepository)

	aas := application.NewAccessAppService(&ar, pr)
	sas := application.NewLicenseAppService(ar, sr, pr, subr, or)

	srv := initGrpcServer(aas, sas, &srvCfg)

	webSrv := initHTTPServer(&srvCfg)
	webSrv.SetCheckRef(srv)
	webSrv.SetSeatRef(srv)
	grpcServer = srv
	return srv, webSrv, nil
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

func initSeatRepository(config *serviceconfig.ServiceConfig) (contracts.SeatLicenseRepository, error) {
	b := NewSeatLicenseRepositoryBuilder()

	return b.WithConfig(config).Build()
}

func initAccessRepository(config *serviceconfig.ServiceConfig) (contracts.AccessRepository, error) {
	return NewAccessRepositoryBuilder().
		WithConfig(config).Build()
}

func initCombinedUserServiceSubjectPrincipalRepository(config *serviceconfig.ServiceConfig) (contracts.SubjectRepository, error) {
	return NewSubjectRepositoryBuilder().
		WithConfig(config).Build()
}

func initStubPrincipalRepository(store string) contracts.PrincipalRepository {
	return NewPrincipalRepositoryBuilder().WithStore(store).Build()
}
