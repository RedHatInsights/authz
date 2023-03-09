package contracts

import (
	"net/http"
	"sync"
)

// Server the interface for the runtime
type Server interface {
	Serve(host string, defaultHandler http.HandlerFunc, wait *sync.WaitGroup) error
	NewServer() Server
}
