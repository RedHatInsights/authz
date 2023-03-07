package dependencies

import "authz/app"

type PrincipalStore interface {
	GetByID(id string) app.Principal
	GetByIDs(ids []string) []app.Principal
	GetByToken(token string) app.Principal
}
