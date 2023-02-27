package host

import (
	"authz/app/dependencies"
)

//A Services object acts as a stub IoC container. The host provides implementations of infrastructure objects here which can then be passed around.
type Services struct {
	//The Store is the current implementation of AuthzStore
	Store dependencies.AuthzStore
}
