package authzed

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
)

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
		Repository:   serviceconfig.SpicedbImage,
		Tag:          serviceconfig.SpicedbVersion, // Replace this with an actual version
		Cmd:          []string{"serve-testing", "--load-configs", "/mnt/spicedb_bootstrap.yaml"},
		Mounts:       []string{path.Join(basepath, "../../../schema/spicedb_bootstrap.yaml") + ":/mnt/spicedb_bootstrap.yaml"},
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
func spicedbTestClient() (*SpiceDbAccessRepository, error) {
	// Generate a random credential to isolate this client from any others.
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	randomKey := base64.StdEncoding.EncodeToString(buf)

	e := &SpiceDbAccessRepository{}
	e.NewConnection("localhost:"+port, randomKey, true, false)

	return e, nil
}

func TestCheckAccess(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	client, err := spicedbTestClient()
	assert.NoError(t, err)

	cases := []struct {
		sub       domain.SubjectID
		operation string
		resource  domain.Resource
		expected  domain.AccessDecision
	}{
		{sub: "u1", operation: "access", resource: domain.Resource{Type: "license", ID: "o1/smarts"}, expected: true},
		{sub: "u1", operation: "access", resource: domain.Resource{Type: "license", ID: "o1/doesnotexist"}, expected: false},
		{sub: "doesnotexist", operation: "access", resource: domain.Resource{Type: "license", ID: "o1/smarts"}, expected: false},
	}

	for _, testcase := range cases {
		actual, err := client.CheckAccess(testcase.sub, testcase.operation, testcase.resource)
		assert.NoError(t, err, fmt.Sprintf("Error in case (subject: %s, operation: %s, resource: [%s, %s])", testcase.sub, testcase.operation, testcase.resource.Type, testcase.resource.ID))
		assert.Equal(t, testcase.expected, actual, "Unexpected result for case (subject: %s, operation: %s, resource: [%s, %s])", testcase.sub, testcase.operation, testcase.resource.Type, testcase.resource.ID)
	}
}

func TestGetLicense(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()

	client, err := spicedbTestClient()
	assert.NoError(t, err)

	lic, err := client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	assert.Equal(t, "o1", lic.OrgID)
	assert.Equal(t, "smarts", lic.ServiceID)
	assert.Equal(t, 10, lic.MaxSeats)
	assert.Equal(t, 1, lic.InUse)
}

func TestGetAssigned(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()

	client, err := spicedbTestClient()
	assert.NoError(t, err)

	assigned, err := client.GetAssigned("o1", "smarts")
	assert.NoError(t, err)

	assert.ElementsMatch(t, []domain.SubjectID{"u1"}, assigned)
}

func TestRapidAssignments(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Parallel()

	client, err := spicedbTestClient()
	assert.NoError(t, err)

	for i := 2; i <= 10; i++ {
		err = client.AssignSeat(domain.SubjectID(fmt.Sprintf("u%d", i)), "o1", domain.Service{ID: "smarts"})
		assert.NoError(t, err)
	}

	lic, err := client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	assert.Equal(t, 10, lic.InUse)
}

func TestAssignUnassign(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Parallel()

	client, err := spicedbTestClient()
	assert.NoError(t, err)

	err = client.AssignSeat("u2", "o1", domain.Service{ID: "smarts"})
	assert.NoError(t, err)

	lic, err := client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	assert.Equal(t, 2, lic.InUse)

	err = client.UnAssignSeat("u2", "o1", domain.Service{ID: "smarts"})
	assert.NoError(t, err)

	lic, err = client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	assert.Equal(t, 1, lic.InUse)
}
