package domain

// GetLicenseEvent represents a request for a license
type GetLicenseEvent struct {
	Requestor SubjectID
	OrgID     string
	ServiceID string
}
