package controllers

import (
	"authz/app"
	"authz/app/contracts"
	"authz/app/dependencies"
	"authz/host/impl"
	"testing"
)

func TestCheckErrorsWhenCallerNotAuthorized(t *testing.T) {
	access := NewAccess(mockAuthzStore())
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
	access := NewAccess(mockAuthzStore())
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
	access := NewAccess(mockAuthzStore())
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

func objFromRequest(requestorID string, subjectID string, operation string, resourceType string, resourceID string) contracts.CheckRequest {
	return contracts.CheckRequest{
		Request: contracts.Request{
			Requestor: app.Principal{ID: requestorID},
		},
		Subject:   app.Principal{ID: subjectID},
		Operation: operation,
		Resource:  app.Resource{Type: resourceType, ID: resourceID},
	}
}

func mockAuthzStore() dependencies.AuthzStore {
	return impl.StubAuthzStore{Data: map[string]bool{
		"system": true,
		"okay":   true,
		"bad":    false,
	}}
}
