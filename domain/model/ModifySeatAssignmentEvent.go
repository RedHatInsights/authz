package model

// ModifySeatAssignmentEvent represents a request to change per-seat license assignments for a given organization and service
type ModifySeatAssignmentEvent struct {
	Request
	Assign   []Principal
	UnAssign []Principal
	Org      Organization
	Service  Service
}
