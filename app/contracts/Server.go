package contracts

import (
	"authz/domain/contracts"
	"sync"
)

// Server the interface for the runtime
type Server interface {
	Serve(wait *sync.WaitGroup, ports ...string) error
	NewServer() Server
	SetEngine(eng contracts.AuthzEngine)
}
