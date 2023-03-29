package model

import "authz/domain/valueobjects"

type GetLicenseEvent struct {
	Requestor valueobjects.SubjectID
	OrgID     string
	ServiceID string
}
