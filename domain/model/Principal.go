package model

// A Principal is an identity that may have some authority
type Principal struct {
	//IDs are permanent and unique identifying values.
	ID string
	//OrgIDs represent the organization a principal is a member of.
	OrgID string
}

// IsAnonymous returns true if this Principal has no identity information and returns false if this Principal represents a specific identity
func (p Principal) IsAnonymous() bool {
	return p.ID == ""
}

// NewPrincipal constructs a new principal with the given identifier.
func NewPrincipal(id string, orgId string) Principal {
	return Principal{ID: id, OrgID: orgId}
}

// NewAnonymousPrincipal constructs a new principal without an identity.
func NewAnonymousPrincipal() Principal {
	return Principal{ID: ""}
}
