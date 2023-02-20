package contracts

import "authz/app"

type CheckRequest struct {
	Request
	Operation string
	Subject   app.Principal
	Resource  app.Resource
}
