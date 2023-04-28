package authzed

import (
	"authz/api"
	"crypto/rand"
	"encoding/base64"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ory/dockertest"
)

type LocalSpiceDbContainerFactory struct {
}

func NewLocalSpiceDbContainerFactory() *LocalSpiceDbContainerFactory {
	return &LocalSpiceDbContainerFactory{}
}

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
		Repository:   api.SpicedbImage,
		Tag:          api.SpicedbVersion, // Replace this with an actual version
		Cmd:          []string{"serve-testing", "--load-configs", "/mnt/spicedb_bootstrap.yaml"},
		Mounts:       []string{path.Join(basepath, "../../../schema/spicedb_bootstrap.yaml") + ":/mnt/spicedb_bootstrap.yaml"},
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

type LocalSpiceDbContainer struct {
	port      string
	container *dockertest.Resource
	pool      *dockertest.Pool
}

func (l *LocalSpiceDbContainer) Port() string {
	return l.port
}

func (l *LocalSpiceDbContainer) NewToken() (string, error) {
	buf := make([]byte, 20)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}

func (l *LocalSpiceDbContainer) WaitForQuantizationInterval() {
	time.Sleep(10 * time.Millisecond)
}

func (l *LocalSpiceDbContainer) CreateClient() (*SpiceDbAccessRepository, error) {

	randomKey, err := l.NewToken()
	if err != nil {
		return nil, err
	}

	e := &SpiceDbAccessRepository{}
	e.NewConnection("localhost:"+l.port, randomKey, true, false)

	return e, nil
}

func (l *LocalSpiceDbContainer) Close() {
	l.pool.Purge(l.container)
}
