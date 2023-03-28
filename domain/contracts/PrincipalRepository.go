package contracts

import (
	"authz/domain/model"
	"authz/domain/valueobjects"
)

// PrincipalRepository is a contract that describes the required operations for accessing principal data
type PrincipalRepository interface {
	// GetByID retrieves a principal for the given ID. If no ID is provided (ex: empty string), it returns an anonymous principal. If any error occurs, it's returned.
	GetByID(id valueobjects.SubjectID) (model.Principal, error)
	// GetByIDs is a bulk version of GetByID to allow the underlying implementation to optimize access to sets of principals and should otherwise have the same behavior.
	GetByIDs(ids []valueobjects.SubjectID) ([]model.Principal, error)
}
