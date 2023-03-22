package model

type ModifySeatAssignmentEvent struct {
	Request
	Assign   []Principal
	UnAssign []Principal
	Org      Organization
	Service  Service
}

func (m ModifySeatAssignmentEvent) IsValid() bool {
	for _, principal := range m.Assign {
		if principal.OrgID != m.Org.Id {
			return false
		}
	}

	for _, principal := range m.UnAssign {
		if principal.OrgID != m.Org.Id {
			return false
		}
	}

	return true
}
