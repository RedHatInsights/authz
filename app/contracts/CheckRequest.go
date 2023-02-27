package contracts

import "authz/app"

//A CheckRequest contains the parameters to request whether a subject can perform an operation on a resource
type CheckRequest struct {
	//The common request parameters
	Request
	//The operation that would be performed
	Operation string
	//The candidate subject who would perform the operation
	Subject app.Principal
	//The resource on which the operation would be performed
	Resource app.Resource
}
