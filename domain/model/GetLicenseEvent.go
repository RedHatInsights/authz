package model

import "authz/domain/valueobjects"

// GetLicenseEvent represents a request for a license
type GetLicenseEvent struct {
	Requestor valueobjects.SubjectID
	OrgID     string
	ServiceID string
}
