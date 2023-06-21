//go:build !release

package testenv

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/authzed/authzed-go/v1"
	"github.com/golang/glog"

	"github.com/ory/dockertest"
)

// LocalActiveMqContainerFactory is only used for test setup and not included in builds with the release tag
type LocalActiveMqContainerFactory struct {
}

// LocalActiveMqContainer struct that holds pointers to the container, dockertest pool and exposes the port
type LocalActiveMqContainer struct {
	mgmtPort      string
	amqpPort      string
	container     *dockertest.Resource
	AuthzedClient *authzed.Client
	pool          *dockertest.Pool
}

// NewLocalActiveMqContainerFactory constructor for the factory
func NewLocalActiveMqContainerFactory() *LocalActiveMqContainerFactory {
	return &LocalActiveMqContainerFactory{}
}

// CreateContainer creates a new SpiceDbContainer using dockertest
func (l *LocalActiveMqContainerFactory) CreateContainer() (*LocalActiveMqContainer, error) {
	pool, err := dockertest.NewPool("") // Empty string uses default docker env
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	pool.MaxWait = 3 * time.Minute

	/*var (
		_, b, _, _ = runtime.Caller(0)
		basepath   = filepath.Dir(b)
	)*/

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "vromero/activemq-artemis",
		Tag:          "latest-alpine",
		ExposedPorts: []string{"61616/tcp", "5672/tcp", "8161/tcp"},
	})

	if err != nil {
		return nil, fmt.Errorf("could not start activeMQ resource: %w", err)
	}

	mgmtPort := resource.GetPort("8161/tcp")
	amqpPort := resource.GetPort("5672/tcp")

	// Give the service time to boot.
	cErr := pool.Retry(func() error {
		log.Print("Attempting to connect to activeMQ...")

		result, err := http.Get(fmt.Sprintf("http://localhost:%s", mgmtPort))
		_ = result
		if err != nil {
			return fmt.Errorf("error connecting to spiceDB: %v", err.Error())
		}

		return err
	})

	if cErr != nil {
		return nil, cErr
	}

	return &LocalActiveMqContainer{
		mgmtPort:  mgmtPort,
		amqpPort:  amqpPort,
		container: resource,
		pool:      pool,
	}, nil
}

// AmqpPort returns the Port the container is listening
func (l *LocalActiveMqContainer) AmqpPort() string {
	return l.amqpPort
}

// Close purges the container
func (l *LocalActiveMqContainer) Close() {
	err := l.pool.Purge(l.container)
	if err != nil {
		glog.Error("Could not purge SpiceDB Container from test. Please delete manually.")
	}
}
