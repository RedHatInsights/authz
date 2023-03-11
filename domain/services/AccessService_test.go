package services

import (
	"authz/domain/contracts"
	"authz/domain/model"
	"authz/infrastructure/engine/mock"
	"testing"
)

func TestCheckErrorsWhenCallerNotAuthorized(t *testing.T) {
	access := NewAccessService(mockAuthzEngine())
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
	access := NewAccessService(mockAuthzEngine())
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
	access := NewAccessService(mockAuthzEngine())
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

func objFromRequest(requestorID string, subjectID string, operation string, resourceType string, resourceID string) model.CheckRequest {
	return model.CheckRequest{
		Request: model.Request{
			Requestor: model.Principal{ID: requestorID},
		},
		Subject:   model.Principal{ID: subjectID},
		Operation: operation,
		Resource:  model.Resource{Type: resourceType, ID: resourceID},
	}
}

func mockAuthzEngine() contracts.AuthzEngine {
	return &mock.StubAuthzEngine{Data: map[string]bool{
		"system": true,
		"okay":   true,
		"bad":    false,
	}}
}
