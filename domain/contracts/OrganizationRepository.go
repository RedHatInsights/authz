package contracts

import "authz/domain"

// OrganizationRepository is a contract that describes the required operations for accessing and manipulating organization and membership data
type OrganizationRepository interface {
	AddSubject(orgID string, subject domain.Subject) error
	UpsertSubject(orgID string, subject domain.Subject) error
}
