package model

// ModifySeatAssignmentEvent represents a request to change per-seat license assignments for a given organization and service
type ModifySeatAssignmentEvent struct {
	Request
	Assign   []Principal
	UnAssign []Principal
	Org      Organization
	Service  Service
}

// IsValid performs some validation on the event to ensure internal consistency
func (m ModifySeatAssignmentEvent) IsValid() bool {
	for _, principal := range m.Assign {
		if principal.OrgID != m.Org.ID {
			return false
		}
	}

	for _, principal := range m.UnAssign {
		if principal.OrgID != m.Org.ID {
			return false
		}
	}

	return true
}
