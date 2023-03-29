package model

type License struct {
	OrgID     string
	ServiceID string
	MaxSeats  int
	InUse     int
}

func NewLicense(orgID string, serviceID string, maxSeats int, assigned int) *License {
	return &License{
		OrgID:     orgID,
		ServiceID: serviceID,
		MaxSeats:  maxSeats,
		InUse:     assigned,
	}
}
