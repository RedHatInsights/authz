package contracts

import "net/http"

type Server interface {
	Serve(host string, defaultHandler http.HandlerFunc) error
	NewServer() Server
}
