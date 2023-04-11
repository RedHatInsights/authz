package domain

// ModifySeatAssignmentEvent represents a request to change per-seat license assignments for a given organization and service
type ModifySeatAssignmentEvent struct {
	Request
	Assign   []SubjectID
	UnAssign []SubjectID
	Org      Organization
	Service  Service
}
