package engine

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"testing"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/authzed-go/v1"
	"github.com/ory/dockertest/v3"
)

// runSpiceDBTestServer spins up a SpiceDB container running the integration
// test server.
func runSpiceDBTestServer(t *testing.T) (port string, err error) {
	pool, err := dockertest.NewPool("") // Empty string uses default docker env
	if err != nil {
		return
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "authzed/spicedb",
		Tag:          "v1.17.0", // Replace this with an actual version
		Cmd:          []string{"serve-testing"},
		ExposedPorts: []string{"50051/tcp", "50052/tcp"},
	})
	if err != nil {
		return
	}

	// When you're done, kill and remove the container
	t.Cleanup(func() {
		_ = pool.Purge(resource)
	})

	return resource.GetPort("50051/tcp"), nil
}

// spicedbTestClient creates a new SpiceDB client with random credentials.
//
// The test server gives each set of a credentials its own isolated datastore
// so that tests can be ran in parallel.
func spicedbTestClient(t *testing.T, port string) (*authzed.Client, error) {
	// Generate a random credential to isolate this client from any others.
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return nil, err
	}
	randomKey := base64.StdEncoding.EncodeToString(buf)

	e := &SpiceDbAuthzEngine{}
	e.NewConnection("localhost:"+port, randomKey)

	return accessConn.client, nil
}

func TestSpiceDB(t *testing.T) {
	port, err := runSpiceDBTestServer(t)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		schema string
	}{
		{
			"basic readback",
			`definition user {}`,
		},
		{
			"readback 2",
			`definition user {}`,
		},
		{
			"Nr 3",
			`definition user {}`,
		},
	}

	for _, tt := range tests {
		tt2 := tt
		t.Run(tt2.name, func(t *testing.T) {
			t.Parallel()

			client, err := spicedbTestClient(t, port)
			if err != nil {
				t.Fatal(err)
			}

			_, err = client.WriteSchema(context.TODO(), &v1.WriteSchemaRequest{Schema: tt2.schema})
			if err != nil {
				t.Fatal(err)
			}

			resp, err := client.ReadSchema(context.TODO(), &v1.ReadSchemaRequest{})
			if err != nil {
				t.Fatal(err)
			}

			if tt2.schema != resp.SchemaText {
				t.Fatal(err)
			}
		})
	}
}
