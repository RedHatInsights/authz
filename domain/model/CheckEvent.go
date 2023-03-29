// Package model contains the Domain model classes.
package model

import "authz/domain/valueobjects"

// A CheckEvent contains the parameters to request whether a subject can perform an operation on a resource
type CheckEvent struct {
	//The common request parameters
	Request
	//The operation that would be performed
	Operation string
	//The candidate subject who would perform the operation
	SubjectID valueobjects.SubjectID
	//The resource on which the operation would be performed
	Resource Resource
}
