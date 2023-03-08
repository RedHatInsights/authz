package dependencies

import "authz/app"

type PrincipalStore interface {
	GetByID(id string) (app.Principal, error)
	GetByIDs(ids []string) ([]app.Principal, error)
	GetByToken(token string) (app.Principal, error)
}
