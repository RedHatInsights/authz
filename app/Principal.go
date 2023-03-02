package app

// A Principal is an identity that may have some authority
type Principal struct {
	//IDs are permanent and unique identifying values.
	ID string
}

// HasIdentity returns true if this principal represents an identity or false if this principal is anonymous.
func (p *Principal) HasIdentity() bool {
	return p.ID != ""
}
