# Authz service

# Start up

run `go run cmd/main.go < -c / --config > < path-to-directory-where-config-file-lives >`

## Example: Running with the provided config.yml in this repo
* from the root directory: `go run cmd/main.go -c .` 
  or `go run cmd/maing.go --config $PWD` 

* from the `cmd` directory: go run main.go -c ../

## Configuring the underlying Store
Set 
```
store:
  kind: "stub"
```
in config.yaml. Set it to `"spicedb"` and fill out the other fields to use spiceDB instead.

# Testing

For complete tests, run `go test ./...` (use -count=1 to avoid caching)
For abbreviated tests, run `go test -short ./...`
To run the test suite many times (ex: checking for flaky tests), run `./scripts/repeat-tests.sh`
For smoketests against an environment, run `./scripts/test.sh <BASEURI> <IDP discovery endpoint>` where 
    - <BASEURI> is the scheme and authority of the environment where the application is running.
    - <IDP discovery endpoint> is the OIDC discovery endpoint for the idp to use for login. This ends in: .well-known/openid-configuration
    Ex: ./scripts/test.sh http://localhost:8081 https://example-keycloak.com/auth/realms/public/.well-known/openid-configuration
    NOTE: When running locally, it must use the spicedb store.


# Building

run `go build -o authz cmd/main.go` or `make binary`

# Developers

[Authz service development guide](docs/development.md)
