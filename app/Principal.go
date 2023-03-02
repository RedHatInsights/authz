package app

// A Principal is an identity that may have some authority
type Principal struct {
	//IDs are permanent and unique identifying values.
	ID string
}

func (p *Principal) HasIdentity() bool {
	return p.ID != ""
}
