package server

import (
	apicontracts "authz/api/contracts"
	"authz/api/handler"
	"authz/app/config"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// see https://github.com/gin-gonic/gin/issues/57

// GinServer underlying struct
type GinServer struct {
	PermissionHandler *handler.PermissionHandler
	ServerConfig      *config.ServerConfig
}

// Serve starts a gin server with a wrapped http Handler from the domain layer.
func (g *GinServer) Serve(wait *sync.WaitGroup) error {
	defer wait.Done()
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello from Gin!",
		})
	})
	err := router.Run(":" + g.ServerConfig.MainPort)
	return err
}

// NewServer creates a new Server object to use.
func (g *GinServer) NewServer(h handler.PermissionHandler, c config.ServerConfig) apicontracts.Server {
	return &GinServer{
		PermissionHandler: &h,
		ServerConfig:      &c,
	}
}

// GetName returns the impl name
func (g *GinServer) GetName() string {
	return "gin"
}
