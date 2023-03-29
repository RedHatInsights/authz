package model

import vo "authz/domain/valueobjects"

// A Principal is an identity that may have some authority
type Principal struct {
	//IDs are permanent and unique identifying values.
	ID vo.SubjectID
}

// IsAnonymous returns true if this Principal has no identity information and returns false if this Principal represents a specific identity
func (p Principal) IsAnonymous() bool {
	return p.ID == ""
}

// NewPrincipal constructs a new principal with the given identifier.
func NewPrincipal(id vo.SubjectID) Principal {
	return Principal{ID: id}
}

// NewAnonymousPrincipal constructs a new principal without an identity.
func NewAnonymousPrincipal() Principal {
	return Principal{ID: ""}
}
