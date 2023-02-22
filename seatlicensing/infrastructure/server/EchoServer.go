package server

import (
	"authz/seatlicensing/domain/contracts"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
)

// see https://github.com/labstack/echo/issues/397

type EchoServer struct{}

// Serve starts a gin server with a wrapped http Handler from the domain layer.
func (e EchoServer) Serve(host string, handler http.HandlerFunc) error {
	e2 := echo.New()
	e2.Use(middleware.Logger())
	e2.Use(middleware.Recover()) //TODO: eval real necessary middlewares, this is just added as per the docs

	// Routes
	e2.GET("/", echo.WrapHandler(handler))
	e2.Logger.Fatal(e2.Start(":" + host))
	return nil //interesting nothing here throws errs... well, for later.
}

func (e EchoServer) NewServer() contracts.Server {
	return EchoServer{}
}