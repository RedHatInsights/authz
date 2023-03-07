package contracts

import "authz/app"

type ModifySeatAssignmentRequest struct {
	Request
	Principals []app.Principal
	Org        app.Organization
	Service    app.Service
}

func (m ModifySeatAssignmentRequest) IsValid() bool {
	for _, principal := range m.Principals {
		if principal.OrgID != m.Org.Id {
			return false
		}
	}

	return true
}
