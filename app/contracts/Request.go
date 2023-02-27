package contracts

import (
	"authz/app"
)

//A Request represents the parameters common to all requests
type Request struct {
	//The principal sending the request
	Requestor app.Principal
}
