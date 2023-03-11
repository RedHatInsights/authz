// Package openfga contains the openfga technical implementation.
package openfga

import (
	"authz/domain/model"
	"context"

	"github.com/golang/glog"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/credentials"
)

// FgaAuthzEngine -
type FgaAuthzEngine struct{}

// FgaClient - Authz client struct
type FgaClient struct {
	client *openfga.APIClient
	ctx    context.Context
}

var openfgaConn *FgaClient

// CheckAccess -
func (o FgaAuthzEngine) CheckAccess(principal model.Principal, operation string, resource model.Resource) (bool, error) {
	trace := false

	body := openfga.CheckRequest{TupleKey: openfga.TupleKey{
		Object:   openfga.PtrString("foo"),
		Relation: openfga.PtrString("bar"),
		User:     openfga.PtrString("baz"),
	}, ContextualTuples: nil, AuthorizationModelId: openfga.PtrString("foo"), Trace: &trace}

	result, _, err := openfgaConn.client.OpenFgaApi.Check(context.Background()).Body(body).Execute()

	if err != nil {
		glog.Errorf("Error checking assertion tuple (%s): %v", body.TupleKey.GetObject(), err)
		return false, err
	}
	return result.GetAllowed(), nil
}

// NewConnection initializes a new openfga client
func (o FgaAuthzEngine) NewConnection(endpoint string, token string) {

	configuration, err := openfga.NewConfiguration(openfga.Configuration{
		ApiScheme: "http", //TODO: derive from endpoint or cfg
		ApiHost:   endpoint,
		StoreId:   "foo", // TODO, dynamic - see experiments example. may result in leaky abstraction though, different to spiceDB. perhaps an engineConfig struct helps
		Credentials: &credentials.Credentials{
			Method: credentials.CredentialsMethodApiToken,
			Config: &credentials.Config{
				ApiToken: token,
			},
		},
	})

	if err != nil {
		glog.Fatalf("unable to initialize client: %s", err)
	}

	client := openfga.NewAPIClient(configuration)

	openfgaConn = &FgaClient{
		client: client,
		ctx:    context.Background(),
	}
}
