package bootstrap

import (
	core "authz/api/gen/v1alpha"
	"authz/api/grpc"
	"authz/infrastructure/repository/authzed"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

// These smoketests should exercise minimal functionality in vertical slices, primarily to ensure correct startup

func TestCheckAccess(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	t.Parallel()
	srv := initializeGrpcServer()

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
	srv := initializeGrpcServer()

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
	srv := initializeGrpcServer()

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
	srv := initializeGrpcServer()

	_, err := srv.ModifySeats(getContext(), &core.ModifySeatsRequest{
		OrgId:     "o1",
		ServiceId: "smarts",
		Assign:    []string{"u2"},
		Unassign:  []string{}, //Unassign:  []string{"u1"}, //TODO: bring back when swap is reintroduced
	})

	assert.NoError(t, err)
}

func getContext() context.Context {
	data := metadata.New(map[string]string{
		"grpcgateway-authorization": "token",
	})

	return metadata.NewIncomingContext(context.Background(), data)
}

var container *authzed.LocalSpiceDbContainer

func initializeGrpcServer() *grpc.Server {
	token, err := container.NewToken()
	if err != nil {
		panic(err)
	}

	grpc, _ := initialize("localhost:"+container.Port(), token, "spicedb", false)

	return grpc
}

func TestMain(m *testing.M) {
	factory := authzed.NewLocalSpiceDbContainerFactory()
	var err error
	container, err = factory.CreateContainer()

	if err != nil {
		fmt.Printf("Error initializing Docker container: %s", err)
		os.Exit(-1)
	}

	result := m.Run()

	container.Close()
	os.Exit(result)
}
