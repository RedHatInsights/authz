package domain

// A Principal is an identity that may have some authority
type Principal struct {
	//IDs are permanent and unique identifying values.
	ID        SubjectID
	FirstName string
	LastName  string
	Username  string
	OrgID     string
}

// DisplayName generates a display name from the principal's data
func (p Principal) DisplayName() string {
	return p.FirstName + " " + p.LastName
}

// IsAnonymous returns true if this Principal has no identity information and returns false if this Principal represents a specific identity
func (p Principal) IsAnonymous() bool {
	return p.ID == ""
}

// NewPrincipal constructs a new principal with the given identifier.
func NewPrincipal(id SubjectID, firstName string, lastName string, userName string, orgID string) Principal {
	return Principal{ID: id, FirstName: firstName, LastName: lastName, Username: userName, OrgID: orgID}
}

// NewAnonymousPrincipal constructs a new principal without an identity.
func NewAnonymousPrincipal() Principal {
	return Principal{ID: ""}
}
