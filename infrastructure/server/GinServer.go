package server

import (
	contracts2 "authz/app/contracts"
	"authz/domain/contracts"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

// see https://github.com/gin-gonic/gin/issues/57

// GinServer underlying struct
type GinServer struct {
	Engine contracts.AuthzEngine
}

// Serve starts a gin server with a wrapped http Handler from the domain layer.
func (g *GinServer) Serve(wait *sync.WaitGroup, ports ...string) error {
	defer wait.Done()
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello from Gin!",
		})
	})
	err := router.Run(":" + ports[0])
	return err
}

// NewServer creates a new Server object to use.
func (g *GinServer) NewServer() contracts2.Server {
	return &GinServer{}
}

// SetEngine Sets the AuthzEngine
func (g *GinServer) SetEngine(eng contracts.AuthzEngine) {
	g.Engine = eng
}

// GetName returns the impl name
func (g *GinServer) GetName() string {
	return "gin"
}