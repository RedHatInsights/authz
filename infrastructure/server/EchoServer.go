// Package server contains different serving libraries, including gin, echo and grpc-gateway implementations.
package server

import (
	"authz/domain/contracts"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// EchoServer underlying struct
type EchoServer struct {
	Engine contracts.AuthzEngine
}

// GetName returns the server name
func (e *EchoServer) GetName() string {
	return "echo"
}

// Serve starts a gin server with a wrapped http Handler from the domain layer.
func (e *EchoServer) Serve(host string, wait *sync.WaitGroup) error {
	defer wait.Done()

	e2 := echo.New()
	e2.Use(middleware.Logger())
	e2.Use(middleware.Recover()) //TODO: eval real necessary middlewares, this is just added as per the docs

	// Routes
	e2.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "Hello from echo!")
	})
	e2.Logger.Fatal(e2.Start(":" + host))
	return nil //interesting nothing here throws errs... well, for later.
}

// NewServer object to call serve from, implementing contract.
func (e *EchoServer) NewServer() contracts.Server {
	return &EchoServer{}
}

// SetEngine sets the AuthzEngine
func (e *EchoServer) SetEngine(eng contracts.AuthzEngine) {
	e.Engine = eng
}
