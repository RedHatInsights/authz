package contracts

import (
	"authz/app"
)

type Request struct {
	Requestor app.Principal
}
