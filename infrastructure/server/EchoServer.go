// Package server contains different serving libraries, including gin, echo and grpc-gateway implementations.
package server

import (
	contracts2 "authz/app/contracts"
	"authz/domain/contracts"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// EchoServer underlying struct
type EchoServer struct {
	AccessRepo contracts.AccessRepository
}

// GetName returns the server name
func (e *EchoServer) GetName() string {
	return "echo"
}

// Serve starts a gin server with a wrapped http Handler from the domain layer.
func (e *EchoServer) Serve(wait *sync.WaitGroup, ports ...string) error {
	defer wait.Done()

	e2 := echo.New()
	e2.Use(middleware.Logger())
	e2.Use(middleware.Recover()) //TODO: eval real necessary middlewares, this is just added as per the docs

	// Routes
	e2.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "Hello from echo!")
	})
	e2.Logger.Fatal(e2.Start(":" + ports[0]))
	return nil //interesting nothing here throws errs... well, for later.
}

// NewServer object to call serve from, implementing contract.
func (e *EchoServer) NewServer() contracts2.Server {
	return &EchoServer{}
}

// SetAccessRepository sets the AccessRepo
func (e *EchoServer) SetAccessRepository(eng contracts.AccessRepository) {
	e.AccessRepo = eng
}
