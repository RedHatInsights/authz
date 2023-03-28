package services

import (
	"authz/domain/contracts"
	"authz/domain/model"
	vo "authz/domain/valueobjects"
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

func objFromRequest(requestorID string, subjectID string, operation string, resourceType string, resourceID string) model.CheckEvent {
	return model.CheckEvent{
		Request: model.Request{
			Requestor: vo.SubjectID(requestorID),
		},
		SubjectID: vo.SubjectID(subjectID),
		Operation: operation,
		Resource:  model.Resource{Type: resourceType, ID: resourceID},
	}
}

func mockAuthzRepository() contracts.AccessRepository {
	return &mock.StubAccessRepository{Data: map[vo.SubjectID]bool{
		"system": true,
		"okay":   true,
		"bad":    false,
	}, LicensedSeats: map[vo.SubjectID]map[string]bool{}}
}
