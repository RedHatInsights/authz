package host

import (
	"authz/app/dependencies"
)

type Services struct {
	Store dependencies.AuthzStore
}
