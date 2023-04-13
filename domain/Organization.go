package domain

// Organization represents an organization/tenant within the system
type Organization struct {
	// ID is the unique id of the organization
	ID string
}

// AsResource converts the Organization into a Resource that can be used for access checks
func (o Organization) AsResource() Resource {
	return Resource{Type: "organization", ID: o.ID}
}
