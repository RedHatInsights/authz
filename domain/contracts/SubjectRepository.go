package contracts

import "authz/domain"

// SubjectRepository represents functionality required to get access-relevant data about subjects
type SubjectRepository interface {
	// GetByOrgID retrieves all members of the given organization
	GetByOrgID(orgID string) (chan domain.Subject, chan error)
}
