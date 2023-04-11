package domain

// Service represents a service/application in the larger environment
type Service struct {
	// ID is the unique name/id of the service
	ID string
}

// AsResource converts the Service into a Resource that can be used for access checks
func (s Service) AsResource() Resource {
	return Resource{Type: "service", ID: s.ID}
}
