# authz service

# Start up

## stub access repository:
run `go run cmd/main.go --store=stub`

## spicedb access repository:
run `go run cmd/main.go --endpoint=<endpoint> --token=<token> --store=spicedb`

# Testing

run `go test ./...` (use -count=1 to avoid caching)

# Building

run `go build build -o authz cmd/main.go` or `make binary`
