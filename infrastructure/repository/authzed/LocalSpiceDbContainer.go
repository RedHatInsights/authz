//go:build !release

package authzed

import (
	"authz/api"
	"crypto/rand"
	"encoding/base64"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang/glog"

	"github.com/ory/dockertest"
)

// LocalSpiceDbContainerFactory is only used for test setup and not included in builds with the release tag
type LocalSpiceDbContainerFactory struct {
}

// NewLocalSpiceDbContainerFactory constructor for the factory
func NewLocalSpiceDbContainerFactory() *LocalSpiceDbContainerFactory {
	return &LocalSpiceDbContainerFactory{}
}

// CreateContainer creates a new SpiceDbContainer using dockertest
func (l *LocalSpiceDbContainerFactory) CreateContainer() (*LocalSpiceDbContainer, error) {
	pool, err := dockertest.NewPool("") // Empty string uses default docker env
	if err != nil {
		return nil, err
	}

	var (
		_, b, _, _ = runtime.Caller(0)
		basepath   = filepath.Dir(b)
	)

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: api.SpicedbImage,
		Tag:        api.SpicedbVersion, // Replace this with an actual version
		Cmd:        []string{"serve-testing", "--load-configs", "/mnt/spicedb_bootstrap.yaml,/mnt/spicedb_bootstrap_relations.yaml"},
		Mounts: []string{
			path.Join(basepath, "../../../schema/spicedb_bootstrap.yaml") + ":/mnt/spicedb_bootstrap.yaml",
			path.Join(basepath, "../../../schema/spicedb_bootstrap_relations.yaml") + ":/mnt/spicedb_bootstrap_relations.yaml",
		},

		ExposedPorts: []string{"50051/tcp", "50052/tcp"},
	})
	if err != nil {
		return nil, err
	}

	return &LocalSpiceDbContainer{
		port:      resource.GetPort("50051/tcp"),
		container: resource,
		pool:      pool,
	}, nil
}

// LocalSpiceDbContainer struct that holds pointers to the container, dockertest pool and exposes the port
type LocalSpiceDbContainer struct {
	port      string
	container *dockertest.Resource
	pool      *dockertest.Pool
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
func (l *LocalSpiceDbContainer) CreateClient() (*SpiceDbAccessRepository, error) {

	randomKey, err := l.NewToken()
	if err != nil {
		return nil, err
	}

	e := &SpiceDbAccessRepository{}
	e.NewConnection("localhost:"+l.port, randomKey, true, false)

	return e, nil
}

// Close purges the container
func (l *LocalSpiceDbContainer) Close() {
	err := l.pool.Purge(l.container)
	if err != nil {
		glog.Error("Could not purge SpiceDB Container from test. Please delete manually.")
	}
}
