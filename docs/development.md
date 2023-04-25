# Authz service development guide

The authz service is a golang compiled binary running on the following ports (by default):
* Grpc:  50051
* HTTP:  8081
* HTTPS: 8443

For development purpose, we use HTTP/8081. 

Authz calls out to SpiceDB on grpc/50051 by default if the command line, below, is used.

## Steps for setting up a local development environment
1. `make kind-create` to create the local kind cluster. (`make kind-delete` to delete an existing cluster.)
2. Ensure there is a SpiceDB token string in `.secrets/spice-db-local`. It will be loaded into a secret and used by SpiceDB.
3. `make kind-create-schema-configmap` to make the default schema and test relationships available to SpiceDB from a ConfigMap.
4. `make kind-spicedb-deploy` to deploy SpiceDB in the kind cluster.
5. `kubectl port-forward $(kubectl get pods -oname -l app.kubernetes.io/name=spicedb --field-selector=status.phase==Running) 50051:50051` to port forward from localhost 50051 to the SpiceDB pod on kind running on 50051. (This command does not exit, so continue on another terminal.)
6. `make binary` to build the authz service locally.
7. `./bin/authz --endpoint=localhost:50051 --token=$(cat .secrets/spice-db-local) --store=spicedb` (This command does not exit, so continue on another terminal.)
8. `./scripts/test.sh localhost:8081` to verify the setup with a smoke test.
