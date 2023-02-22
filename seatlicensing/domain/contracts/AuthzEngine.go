package contracts

import "authz/seatlicensing/domain/model"

type AuthzEngine interface {
	CheckAccess(principal model.Principal, operation string, resource model.Resource) (bool, error)
}
