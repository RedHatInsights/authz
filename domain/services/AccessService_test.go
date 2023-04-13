package services

import (
	"authz/domain"
	"authz/domain/contracts"
	"authz/infrastructure/repository/mock"
	"testing"
)

func TestCheckErrorsWhenCallerNotAuthorized(t *testing.T) {
	t.SkipNow()
	access := NewAccessService(mockAuthzRepository())
	_, err := access.Check(objFromRequest(
		"other system",
		"okay",
		"check",
		"license",
		"seat"))

	if err == nil {
		t.Error("Expected caller authorization error, got success")
	}
}

func TestCheckReturnsTrueWhenStoreReturnsTrue(t *testing.T) {
	access := NewAccessService(mockAuthzRepository())
	result, err := access.Check(objFromRequest(
		"system",
		"okay",
		"check",
		"license",
		"seat"))

	if err != nil {
		t.Errorf("Expected a result, got error: %s", err)
	}

	if result != true {
		t.Errorf("Expected success, got fail.")
	}
}

func TestCheckReturnsFalseWhenStoreReturnsFalse(t *testing.T) {
	access := NewAccessService(mockAuthzRepository())
	result, err := access.Check(objFromRequest(
		"system",
		"bad",
		"check",
		"license",
		"seat"))

	if err != nil {
		t.Errorf("Expected a result, got error: %s", err)
	}

	if result != false {
		t.Errorf("Expected fail, got success.")
	}
}

func objFromRequest(requestorID string, subjectID string, operation string, resourceType string, resourceID string) domain.CheckEvent {
	return domain.CheckEvent{
		Request: domain.Request{
			Requestor: domain.SubjectID(requestorID),
		},
		SubjectID: domain.SubjectID(subjectID),
		Operation: operation,
		Resource:  domain.Resource{Type: resourceType, ID: resourceID},
	}
}

func mockAuthzRepository() contracts.AccessRepository {
	return &mock.StubAccessRepository{Data: map[domain.SubjectID]bool{
		"system": true,
		"okay":   true,
		"bad":    false,
	}, LicensedSeats: map[string]map[domain.SubjectID]bool{}}
}
