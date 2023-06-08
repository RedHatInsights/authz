//go:build !release

package authzed

import (
	"authz/bootstrap/serviceconfig"
	"authz/infrastructure/grpcutil"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/authzed/authzed-go/v1"
	"log"
	"path"
	"path/filepath"
	"runtime"
	"time"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/golang/glog"

	"github.com/ory/dockertest"
)

// LocalSpiceDbContainerFactory is only used for test setup and not included in builds with the release tag
type LocalSpiceDbContainerFactory struct {
}

// LocalSpiceDbContainer struct that holds pointers to the container, dockertest pool and exposes the port
type LocalSpiceDbContainer struct {
	port          string
	container     *dockertest.Resource
	AuthzedClient *authzed.Client
	pool          *dockertest.Pool
}

// NewLocalSpiceDbContainerFactory constructor for the factory
func NewLocalSpiceDbContainerFactory() *LocalSpiceDbContainerFactory {
	return &LocalSpiceDbContainerFactory{}
}

// CreateContainer creates a new SpiceDbContainer using dockertest
func (l *LocalSpiceDbContainerFactory) CreateContainer() (*LocalSpiceDbContainer, error) {
	pool, err := dockertest.NewPool("") // Empty string uses default docker env
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	pool.MaxWait = 3 * time.Minute

	var (
		_, b, _, _ = runtime.Caller(0)
		basepath   = filepath.Dir(b)
	)

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: serviceconfig.SpicedbImage,
		Tag:        serviceconfig.SpicedbVersion, // Replace this with an actual version
		Cmd:        []string{"serve-testing", "--skip-release-check=true", "--load-configs", "/mnt/spicedb_bootstrap.yaml,/mnt/spicedb_bootstrap_relations.yaml"},
		Mounts: []string{
			path.Join(basepath, "../../../schema/spicedb_bootstrap.yaml") + ":/mnt/spicedb_bootstrap.yaml",
			path.Join(basepath, "../../../schema/spicedb_bootstrap_relations.yaml") + ":/mnt/spicedb_bootstrap_relations.yaml",
		},
		ExposedPorts: []string{"50051/tcp", "50052/tcp"},
	})

	if err != nil {
		return nil, fmt.Errorf("could not start spicedb resource: %w", err)
	}

	port := resource.GetPort("50051/tcp")

	// Give the service time to boot.
	cErr := pool.Retry(func() error {
		log.Print("Attempting to connect to spicedb...")

		conn, err := grpc.Dial(
			fmt.Sprintf("localhost:%s", port),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpcutil.WithInsecureBearerToken("test"),
		)
		if err != nil {
			return fmt.Errorf("error connecting to spiceDB: %v", err.Error())
		}

		client := v1.NewSchemaServiceClient(conn)

		//read scheme we add via mount
		_, err = client.ReadSchema(context.Background(), &v1.ReadSchemaRequest{})

		return err
	})

	if cErr != nil {
		return nil, cErr
	}

	return &LocalSpiceDbContainer{
		port:      port,
		container: resource,
		pool:      pool,
	}, nil
}

// Port returns the Port the container is listening
func (l *LocalSpiceDbContainer) Port() string {
	return l.port
}

// NewToken returns a new token used for the container so a new store is created in serve-testing
func (l *LocalSpiceDbContainer) NewToken() (string, error) {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}

// WaitForQuantizationInterval needed to avoid read-before-write when loading the schema
func (l *LocalSpiceDbContainer) WaitForQuantizationInterval() {
	time.Sleep(10 * time.Millisecond)
}

// CreateClient creates a new client that connects to the dockerized spicedb instance and the right store
func (l *LocalSpiceDbContainer) CreateClient() (*SpiceDbAccessRepository, *authzed.Client, error) {

	randomKey, err := l.NewToken()
	if err != nil {
		return nil, nil, err
	}

	e := &SpiceDbAccessRepository{}
	err = e.NewConnection("localhost:"+l.port, randomKey, true, false)
	if err != nil {
		return nil, nil, err
	}

	return e, e.client, nil
}

// Close purges the container
func (l *LocalSpiceDbContainer) Close() {
	err := l.pool.Purge(l.container)
	if err != nil {
		glog.Error("Could not purge SpiceDB Container from test. Please delete manually.")
	}
}
