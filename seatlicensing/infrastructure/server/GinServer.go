package server

import (
	"authz/seatlicensing/domain/contracts"
	"github.com/gin-gonic/gin"
	"net/http"
)

// see https://github.com/gin-gonic/gin/issues/57
type GinServer struct{}

// Serve starts a gin server with a wrapped http Handler from the domain layer.
func (g GinServer) Serve(host string, handler http.HandlerFunc) error {
	router := gin.Default()
	router.GET("/", gin.WrapF(handler))
	err := router.Run(":" + host)
	return err
}

func (g GinServer) NewServer() contracts.Server {
	return GinServer{}
}
