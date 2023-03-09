package contracts

import "net/http"

// Server the interface for the runtime
type Server interface {
	Serve(host string, defaultHandler http.HandlerFunc) error
	NewServer() Server
}
