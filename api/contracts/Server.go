// Package contracts in the api package defines API / runtime specific contracts
// for things that are communication-related and
// need to be abstracted away (TBD)
package contracts

import (
	"authz/api/handler"
	"authz/app/config"
	"sync"
)

// Server the interface for the runtime
type Server interface {
	Serve(wait *sync.WaitGroup) error
	NewServer(handler handler.PermissionHandler, cfg config.ServerConfig) Server
}
