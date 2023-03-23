# Authz service

# Start up

## Start using stub access repository:
run `go run cmd/main.go --store=stub`

## Start using spicedb access repository:
run `go run cmd/main.go --endpoint=<endpoint> --token=<token> --store=spicedb`

# Testing

For complete tests, run `go test ./...` (use -count=1 to avoid caching)
For abbreviated tests, run `go test -short ./...`


# Building

run `go build -o authz cmd/main.go` or `make binary`
