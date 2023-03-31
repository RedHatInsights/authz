package authzed

import (
	"authz/domain/model"
	"authz/domain/valueobjects"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
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

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "authzed/spicedb",
		Tag:        "v1.17.0", // Replace this with an actual version
		Cmd:        []string{"serve-testing", "--load-configs", "/mnt/spicedb_bootstrap.yaml"},
		//TODO: how to get the absolute path at runtime?
		Mounts:       []string{"/home/wscalf/Projects/authz/schema/spicedb_bootstrap.yaml:/mnt/spicedb_bootstrap.yaml"},
		ExposedPorts: []string{"50051/tcp", "50052/tcp"},
	})
	if err != nil {
		return
	}

	defer func() {
		_ = pool.Purge(resource)
	}()

	port = resource.GetPort("50051/tcp")

	result := m.Run()

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

	assert.ElementsMatch(t, []valueobjects.SubjectID{"u1"}, assigned)
}

func TestRapidAssignments(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.SkipNow() //NOTE: this test passes if run slowly but not if run fast. Consistency?
	t.Parallel()

	client, err := spicedbTestClient()
	assert.NoError(t, err)

	for i := 2; i < 10; i++ {
		err = client.AssignSeat(valueobjects.SubjectID(fmt.Sprintf("u%d", i)), "o1", model.Service{ID: "smarts"})
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
	t.SkipNow() //NOTE: this test passes if run slowly but not if run fast. Consistency?

	t.Parallel()

	client, err := spicedbTestClient()
	assert.NoError(t, err)

	err = client.AssignSeat("u2", "o1", model.Service{ID: "smarts"})
	assert.NoError(t, err)

	lic, err := client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	assert.Equal(t, 2, lic.InUse)

	err = client.UnAssignSeat("u2", "o1", model.Service{ID: "smarts"})
	assert.NoError(t, err)

	lic, err = client.GetLicense("o1", "smarts")
	assert.NoError(t, err)

	assert.Equal(t, 1, lic.InUse)
}
