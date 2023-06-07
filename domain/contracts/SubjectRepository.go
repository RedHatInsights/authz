package contracts

import "authz/domain"

type SubjectRepository interface {
	// GetByOrgID retrieves all members of the given organization
	GetByOrgID(orgID string) (chan domain.Subject, chan error)
}
