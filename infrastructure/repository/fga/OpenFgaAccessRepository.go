// Package fga contains the fga technical implementation.
package fga

import (
	"authz/domain/model"
	vo "authz/domain/valueobjects"
	"context"

	"github.com/golang/glog"
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/credentials"
)

// OpenFgaAccessRepository -
type OpenFgaAccessRepository struct{}

// OpenFgaClient - Authz client struct
type OpenFgaClient struct {
	client *openfga.APIClient
	ctx    context.Context
}

var openfgaConn *OpenFgaClient

// CheckAccess -
func (o OpenFgaAccessRepository) CheckAccess(subjectID vo.SubjectID, operation string, resource model.Resource) (vo.AccessDecision, error) {
	trace := false

	body := openfga.CheckRequest{TupleKey: openfga.TupleKey{
		Object:   openfga.PtrString(resource.ID),
		Relation: openfga.PtrString(operation),
		User:     openfga.PtrString(string(subjectID)),
	}, ContextualTuples: nil, AuthorizationModelId: openfga.PtrString("foo"), Trace: &trace}

	result, _, err := openfgaConn.client.OpenFgaApi.Check(context.Background()).Body(body).Execute()

	if err != nil {
		glog.Errorf("Error checking assertion tuple (%s): %v", body.TupleKey.GetObject(), err)
		return false, err
	}
	return vo.AccessDecision(result.GetAllowed()), nil
}

// NewConnection initializes a new openfga client
func (o OpenFgaAccessRepository) NewConnection(endpoint string, token string, _ bool, _ bool) {

	configuration, err := openfga.NewConfiguration(openfga.Configuration{
		ApiScheme: "http", //TODO: derive from endpoint or cfg
		ApiHost:   endpoint,
		StoreId:   "foo", // TODO, dynamic - see experiments example. may result in leaky abstraction though, different to spiceDB. perhaps an accessRepoConfig struct helps
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

	openfgaConn = &OpenFgaClient{
		client: client,
		ctx:    context.Background(),
	}
}
