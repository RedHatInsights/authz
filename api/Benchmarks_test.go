package api

import (
	core "authz/api/gen/v1alpha"
	"context"
	"net/http"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func BenchmarkCheckAccessHTTP(b *testing.B) {
	//NOTE: make sure authz is running as a standalone process before running

	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8081/v1alpha/check", strings.NewReader(`{"subject": "alice", "operation": "access", "resourcetype": "license", "resourceid": "foo"}`))
		if err != nil {
			panic(err)
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "token")

		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			panic(err)
		}

		if resp.StatusCode != http.StatusOK {
			panic(resp.StatusCode)
		}
	}
}

func BenchmarkCheckAccessGRPC(b *testing.B) {
	// NOTE: make sure authz is running as a standalone process before running
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	client := core.NewCheckPermissionClient(conn)

	for i := 0; i < b.N; i++ {
		req := &core.CheckPermissionRequest{
			Subject:      "alice",
			Operation:    "access",
			Resourcetype: "license",
			Resourceid:   "foo",
		}

		ctx := metadata.AppendToOutgoingContext(context.Background(), "bearer-token", "token")
		_, err := client.CheckPermission(ctx, req)

		if err != nil {
			panic(err)
		}
	}
}
