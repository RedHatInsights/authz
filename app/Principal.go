package app

// A Principal is an identity that may have some authority
type Principal struct {
	//IDs are permanent and unique identifying values.
	ID string
	//OrgIDs represent the organization a principal is a member of.
	OrgID string
	//Names are human-readable names
	Name string
	//Status of a license for the principal  in the context of a specific service
	IsLicenseActive bool //TODO: Not sure if we won't need another struct instead. may be mixing concerns.
}

// IsAnonymous returns true if this Principal has no identity information and returns false if this Principal represents a specific identity
func (p Principal) IsAnonymous() bool {
	return p.ID == ""
}

// NewPrincipal constructs a new principal with the given identifier.
func NewPrincipal(id string, name string, orgId string, isLicenseActive bool) Principal {
	return Principal{ID: id, Name: name, OrgID: orgId, IsLicenseActive: isLicenseActive}
}

// NewAnonymousPrincipal constructs a new principal without an identity.
func NewAnonymousPrincipal() Principal {
	return Principal{ID: ""}
}
