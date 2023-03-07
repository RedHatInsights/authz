package host

import (
	"authz/app/dependencies"
)

// A Services object acts as a stub IoC container. The host provides implementations of infrastructure objects here which can then be passed around.
type Services struct {
	//The Authz is the current implementation of AuthzStore
	Authz      dependencies.AuthzStore
	Principals dependencies.PrincipalStore
	Licensing  dependencies.LicenseStore
}
