package services

import (
	"authz/api"
	"authz/domain"
	"authz/domain/contracts"
	"authz/infrastructure/repository/authzed"
	"crypto/rand"
	"encoding/base64"
	"github.com/ory/dockertest/v3"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCheckErrorsWhenCallerNotAuthorized(t *testing.T) {
	t.SkipNow() //Skip until meta-authz is in place
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
	client, err := spicedbTestClient()
	if err != nil {
		panic(err)
	}
	return client
}

var port string

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("") // Empty string uses default docker env
	if err != nil {
		return
	}

	var (
		_, b, _, _ = runtime.Caller(0)
		basepath   = filepath.Dir(b)
	)

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   api.SpicedbImage,
		Tag:          api.SpicedbVersion, // Replace this with an actual version
		Cmd:          []string{"serve-testing", "--load-configs", "/mnt/spicedb_bootstrap.yaml"},
		Mounts:       []string{path.Join(basepath, "../../schema/spicedb_bootstrap.yaml") + ":/mnt/spicedb_bootstrap.yaml"},
		ExposedPorts: []string{"50051/tcp", "50052/tcp"},
	})
	if err != nil {
		return
	}

	port = resource.GetPort("50051/tcp")

	result := m.Run()
	_ = pool.Purge(resource)

	os.Exit(result)
}

// spicedbTestClient creates a new SpiceDB client with random credentials.
//
// The test server gives each set of a credentials its own isolated datastore
// so that tests can be ran in parallel.
func spicedbTestClient() (*authzed.SpiceDbAccessRepository, error) {
	// Generate a random credential to isolate this client from any others.
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	randomKey := base64.StdEncoding.EncodeToString(buf)

	e := &authzed.SpiceDbAccessRepository{}
	e.NewConnection("localhost:"+port, randomKey, true, false)

	return e, nil
}
