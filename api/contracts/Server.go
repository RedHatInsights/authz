// Package contracts in the api package defines API / runtime specific contracts
// for things that are communication-related and
// need to be abstracted away (TBD)
package contracts

import (
	"authz/api/handler"
	"sync"
)

// Server the interface for the runtime
type Server interface {
	Serve(wait *sync.WaitGroup, ports ...string) error
	NewServer(handler handler.PermissionHandler) Server
}
