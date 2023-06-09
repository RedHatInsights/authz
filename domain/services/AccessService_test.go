package services

import (
	"authz/domain"
	"authz/domain/contracts"
	"testing"
)

func TestCheckErrorsWhenCallerNotAuthorized(t *testing.T) {
	t.SkipNow() //Skip until meta-authz is in place
	access := NewAccessService(spicedbSeatLicenseRepository())
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
	access := NewAccessService(spicedbSeatLicenseRepository())
	result, err := access.Check(objFromRequest(
		"system",
		"u1",
		"access",
		"license",
		"o1/smarts"))

	if err != nil {
		t.Errorf("Expected a result, got error: %s", err)
	}

	if result != true {
		t.Errorf("Expected success, got fail.")
	}
}

func TestCheckReturnsFalseWhenStoreReturnsFalse(t *testing.T) {
	access := NewAccessService(spicedbSeatLicenseRepository())
	result, err := access.Check(objFromRequest(
		"system",
		"bad",
		"access",
		"license",
		"o1/smarts"))

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

func spicedbSeatLicenseRepository() contracts.AccessRepository {
	client, _, err := spicedbContainer.CreateClient()
	if err != nil {
		panic(err)
	}
	return client
}
