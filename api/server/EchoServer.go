// Package server contains different serving libraries, including gin, echo and grpc-gateway implementations.
package server

import (
	apicontracts "authz/api/contracts"
	"authz/api/handler"
	"authz/app/config"
	"authz/domain/model"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// EchoServer underlying struct
type EchoServer struct {
	PermissionHandler *handler.PermissionHandler
	SeatHandler       *handler.SeatHandler
	ServerConfig      *config.ServerConfig
}

// GetName returns the server name
func (e *EchoServer) GetName() string {
	return "echo"
}

// Serve starts a gin server with a wrapped http GrpcCheckService from the domain layer.
func (e *EchoServer) Serve(wait *sync.WaitGroup) error {
	defer wait.Done()

	e2 := echo.New()
	e2.Use(middleware.Logger())
	e2.Use(middleware.Recover()) //TODO: eval real necessary middlewares, this is just added as per the docs

	// Routes
	e2.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "Hello from echo!")
	})

	e2.POST("/permissions/check", func(c echo.Context) error {
		checkRes, _ := e.PermissionHandler.Check(handler.CheckRequest{
			Requestor:    model.Principal{ID: "token"}, // these fields should come from the actual http req body
			Subject:      "",
			ResourceType: "",
			ResourceID:   "",
			Operation:    "",
		})
		return c.JSON(http.StatusOK, checkRes)
	})
	e2.Logger.Fatal(e2.Start(":" + e.ServerConfig.MainPort))
	return nil //interesting nothing here throws errs... well, for later.
}

// NewServer object to call serve from, implementing contract.
func (e *EchoServer) NewServer(h handler.PermissionHandler, c config.ServerConfig) apicontracts.Server {
	return &EchoServer{
		PermissionHandler: &h,
		ServerConfig:      &c,
	}
}
