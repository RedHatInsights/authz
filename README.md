# Authz Service

# Build Commands
`make binary`

# Run 
## For using stub store
`./authz serve --endpoint=<endpoint>:50051 --token=<token> --store=stub`
## For using spicedb store
`./authz serve --endpoint=<endpoint>:50051 --token=<token> --store=spicedb`