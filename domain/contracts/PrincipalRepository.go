package contracts

import "authz/domain/model"

type PrincipalRepository interface {
	GetByID(id string) (model.Principal, error)
	GetByIDs(ids []string) ([]model.Principal, error)
	GetByToken(token string) (model.Principal, error)
}
