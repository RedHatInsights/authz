# Authz service

# Start up

## Start using stub access repository:
run `go run cmd/main.go --store=stub`

## Start using spicedb access repository:
run `go run cmd/main.go --endpoint=<endpoint> --token=<token> --store=spicedb --useTLS=false`

# Testing

For complete tests, run `go test ./...` (use -count=1 to avoid caching)
For abbreviated tests, run `go test -short ./...`
For smoketests against an environment, run `./scripts/test.sh <BASEURI>` where <BASEURI> is the scheme and authority of the environment where the application is running.
    Ex: ./scripts/test.sh http://localhost:8081
    NOTE: When running locally, it must use the spicedb store.
To run the test suite many times (ex: checking for flaky tests), run `./scripts/repeat-tests.sh`


# Building

run `go build -o authz cmd/main.go` or `make binary`

# Developers

[Authz service development guide](docs/development.md)
