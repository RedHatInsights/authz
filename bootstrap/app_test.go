package bootstrap

import (
	"authz/api"
	core "authz/api/gen/v1alpha"
	"authz/api/grpc"
	"context"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"sync/atomic"
	"testing"

	"github.com/golang/glog"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

// These smoketests should exercise minimal functionality in vertical slices, primarily to ensure correct startup

func TestCheckAccess(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	srv := initializeGrpcServer(t)

	resp, err := srv.CheckPermission(getContext(), &core.CheckPermissionRequest{
		Subject:      "u1",
		Operation:    "access",
		Resourcetype: "license",
		Resourceid:   "o1/smarts",
	})

	assert.NoError(t, err)

	assert.True(t, resp.Result)
}

func TestGetLicense(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	srv := initializeGrpcServer(t)

	resp, err := srv.GetLicense(getContext(), &core.GetLicenseRequest{
		OrgId:     "o1",
		ServiceId: "smarts",
	})

	assert.NoError(t, err)

	assert.EqualValues(t, 9, resp.SeatsAvailable)
	assert.EqualValues(t, 10, resp.SeatsTotal)
}

func TestGetAssigned(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	srv := initializeGrpcServer(t)

	includeUsers := true
	filter := core.SeatFilterType_assigned
	resp, err := srv.GetSeats(getContext(), &core.GetSeatsRequest{
		OrgId:        "o1",
		ServiceId:    "smarts",
		IncludeUsers: &includeUsers,
		Filter:       &filter,
	})

	assert.NoError(t, err)

	assert.EqualValues(t, 1, len(resp.Users))
}

func TestModify(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	srv := initializeGrpcServer(t)

	_, err := srv.ModifySeats(getContext(), &core.ModifySeatsRequest{
		OrgId:     "o1",
		ServiceId: "smarts",
		Assign:    []string{"u2"},
		Unassign:  []string{"u1"},
	})

	assert.NoError(t, err)
}

func getContext() context.Context {
	data := metadata.New(map[string]string{
		"grpcgateway-authorization": "token",
	})

	return metadata.NewIncomingContext(context.Background(), data)
}

var port string

func initializeGrpcServer(t *testing.T) *grpc.Server {
	token, err := serialKey()
	assert.NoError(t, err)

	grpc, _ := initialize("localhost:"+port, token, "spicedb", false)

	return grpc
}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("") // Empty string uses default docker env
	if err != nil {
		glog.Fatalf("Failed to initialize dockertest pool: %s", err)
		os.Exit(1)
	}

	var (
		_, b, _, _ = runtime.Caller(0)
		basepath   = filepath.Dir(b)
	)

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   api.SpicedbImage,
		Tag:          api.SpicedbVersion, // Replace this with an actual version
		Cmd:          []string{"serve-testing", "--load-configs", "/mnt/spicedb_bootstrap.yaml"},
		Mounts:       []string{path.Join(basepath, "../schema/spicedb_bootstrap.yaml") + ":/mnt/spicedb_bootstrap.yaml"},
		ExposedPorts: []string{"50051/tcp"},
	})
	if err != nil {
		glog.Fatalf("Failed to create SpiceDB container: %s", err)
		os.Exit(1)
	}

	port = resource.GetPort("50051/tcp")

	result := m.Run()
	_ = pool.Purge(resource)

	os.Exit(result)
}

var keyData int32 = 1

func serialKey() (string, error) {
	atomic.AddInt32(&keyData, 1)
	return strconv.Itoa(int(keyData)), nil
}
