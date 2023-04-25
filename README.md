# Authz service

# Start up

run `go run cmd/main.go <-c / --config> <path-to-config-without-file>`

## Example: Running with the provided config.yml in this repo
* from the root directory: `go run cmd/main.go -c .`
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
For smoketests against an environment, run `./scripts/test.sh <BASEURI>` where <BASEURI> is the scheme and authority of the environment where the application is running.
    Ex: ./scripts/test.sh http://localhost:8081
    NOTE: When running locally, it must use the spicedb store.


# Building

run `go build -o authz cmd/main.go` or `make binary`

# Developers

[Authz service development guide](docs/development.md)
