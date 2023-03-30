package model

// License represents a license purchased by an org for a service
type License struct {
	OrgID     string
	ServiceID string
	MaxSeats  int
	InUse     int
}

// NewLicense constructs a new License entity
func NewLicense(orgID string, serviceID string, maxSeats int, assigned int) *License {
	return &License{
		OrgID:     orgID,
		ServiceID: serviceID,
		MaxSeats:  maxSeats,
		InUse:     assigned,
	}
}

// GetAvailableSeats - Get available seats Max - InUSe
func (l *License) GetAvailableSeats() int {
	return l.MaxSeats - l.InUse
}
