package contracts

import (
	"sync"
)

// Server the interface for the runtime
type Server interface {
	Serve(host string, wait *sync.WaitGroup) error
	NewServer() Server
	SetEngine(eng AuthzEngine)
	GetName() string
}
