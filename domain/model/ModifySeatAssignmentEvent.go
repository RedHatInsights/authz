package model

import vo "authz/domain/valueobjects"

// ModifySeatAssignmentEvent represents a request to change per-seat license assignments for a given organization and service
type ModifySeatAssignmentEvent struct {
	Request
	Assign   []vo.SubjectID
	UnAssign []vo.SubjectID
	Org      Organization
	Service  Service
}
