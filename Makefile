GO := go
GOFMT := gofmt
PROTOC_INSTALLED := $(shell which protoc 2>/dev/null)
BUF_INSTALLED := $(shell which buf 2>/dev/null)
GEN_GW_INSTALLED := $(shell which protoc-gen-grpc-gateway 2>/dev/null)
GEN_OPENAPI_INSTALLED := $(shell which protoc-gen-openapiv2 2>/dev/null)
GEN_GO_GRPC_INSTALLED := $(shell which protoc-gen-go-grpc 2>/dev/null)
GEN_GO_INSTALLED := $(shell which protoc-gen-go 2>/dev/null)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell $(GO) env GOBIN))
GOBIN=$(shell $(GO) env GOPATH)/bin
else
GOBIN=$(shell $(GO) env GOBIN)
endif

DOCKER ?= docker
DOCKER_CONFIG="${PWD}/.docker"
SHELL = bash

# builds the binary inside the bin folder
.PHONY: binary
binary:
	@echo "Building the service..."
	@$(GO) build -o bin/authz cmd/main.go

# builds the binary inside the bin folder
.PHONY: binary-delete
binary-delete:
	@echo "removing binary"
	@rm -rf bin/
	@echo "binary successfully removed"

# starts a kind cluster
.PHONY: kind-create
kind-create:
	@kind create cluster --name=authz --config=k8s/kind.yml
	@kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	# make kind-deploy (below) fails (admission controller fails on ingress) unless nginx controller pod is ready.
	@echo "Waiting for nginx to become available..."
	@kubectl wait --for condition=Available=True deployment.apps/ingress-nginx-controller -n ingress-nginx --timeout 20s
	@kubectl wait --for=condition=complete job/ingress-nginx-admission-create -n ingress-nginx --timeout 60s
	@kubectl wait --for=condition=complete job/ingress-nginx-admission-patch -n ingress-nginx --timeout 60s
	@kubectl wait --for condition=Ready=True pod -l app.kubernetes.io/component=controller -n ingress-nginx --timeout 120s

# deploys the authz server to the kind cluster
.PHONY: kind-deploy
kind-deploy:
	@kubectl apply -f k8s/authz.yml
	@echo "Wait for pods with: kubectl get pods,svc"
	@echo "See k8s/README.md for more info."

# adds the spiceDB schema to the kind clusters configmap
.PHONY: kind-create-schema-configmap
kind-create-schema-configmap:
	@kubectl create configmap spicedb-schema --from-file=schema/spicedb_bootstrap.yaml

# deploys spiceDB
.PHONY: kind-spicedb-deploy
kind-spicedb-deploy:
	@kubectl create secret generic spicedb --from-file=SPICEDB_GRPC_PRESHARED_KEY=.secrets/spice-db-local
	@kubectl apply -f k8s/spicedb.yaml

# deletes the kind cluster
.PHONY: kind-delete
kind-delete:
	@kind delete cluster --name=authz

# creates tls certificate and key in the tls dir
.PHONY: tls-cert
tls-cert:
	@echo "creating directory tls/"
	@mkdir -p tls
	@echo "Generating self-signed TLS certs. needs openssl 1.1.1 or newer..."
	@openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes \
               -keyout tls/tls.key -out tls/tls.crt -subj "/CN=localhost" \
               -addext "subjectAltName=DNS:example.com,DNS:www.example.net,IP:10.0.0.1"
	@echo "Success! find your cert files in the tls/ folder"

# delete the tls directory and certs
.PHONY: tls-delete
tls-delete:
	@echo "removing generated tls certs"
	@rm -rf tls/
	@echo "TLS certs successfully removed"

# validate the openapi schema
.PHONY: apigen-validate
apigen-validate:
	$(DOCKER) run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli validate -i /local/api/v1alpha/openapi-authz-v1alpha.yaml

# Run Swagger UI and host the api docs
.PHONY: apidocs-start
apidocs-ui:
	$(DOCKER) run --rm --name swagger_ui_docs -d -p 8082:8080 -e URLS="[ \
		{ url: \"/openapi/v1alpha/openapi-authz-v1alpha.yaml\", name: \"Authz API\"}]"\
		  -v $(PWD)/api/:/usr/share/nginx/html/openapi:Z swaggerapi/swagger-ui
	@echo "Please open http://localhost:8082/"

# Remove Swagger container
.PHONY: apidocs-stop
apidocs-stop:
	$(DOCKER) container stop swagger_ui_docs
	$(DOCKER) container rm swagger_ui_docs

#convert grpc gateway openapiv2 spec to openapi v3
.PHONY: apigen-v3
apigen-v3:
	@echo "generating v3 openapi.yaml from grpc-gateway v2 yaml..."
	$(DOCKER) run --rm --name swagger_codegen \
		  -v $(PWD)/api/gen/v1alpha/:/opt/mnt:Z -w /opt/mnt swaggerapi/swagger-codegen-cli-v3:3.0.41 generate -i ./core.swagger.yaml -l openapi-yaml -o .
	@echo "generating v3 openapi.json from grpc-gateway v2 json..."
	$(DOCKER) run --rm --name swagger_codegen \
		  -v $(PWD)/api/gen/v1alpha/:/opt/mnt:Z -w /opt/mnt swaggerapi/swagger-codegen-cli-v3:3.0.41 generate -i ./core.swagger.json -l openapi -o .
	@echo "Remove unnecessary generated artifacts"
	@make apigen-v2-delete
	@echo "move and rename files to v1alpha directory"
	@cd api/gen/v1alpha && mv openapi.yaml ../../v1alpha/openapi-authz-v1alpha.yaml
	@cd api/gen/v1alpha && mv openapi.json ../../v1alpha/openapi-authz-v1alpha.json

# remove generated openAPI artifacts
.PHONY: apigen-v2-delete
apigen-v2-delete:
	@echo "removing openapi v2 artifacts"
	@cd api/gen/v1alpha && rm -rf .swagger-codegen/ && rm -f README.md && rm -f .swagger-codegen-ignore
	@echo "opanpi v2 artifacts successfully removed"


# Generate grpc gateway code from proto via buf
.PHONY: buf-gen
buf-gen:
#check if protoc is installed
ifndef PROTOC_INSTALLED
	$(error "protoc is not installed, please install protoc - see https://grpc.io/docs/protoc-installation/ ")
endif
#check if buf is installed
ifndef BUF_INSTALLED
	$(error "Buf is not installed, please install buf - see https://docs.buf.build/installation")
endif
# install dependencies if not installed yet. see https://github.com/grpc-ecosystem/grpc-gateway#installation - versions are derived from go.mod via tools.go
ifndef GEN_GW_INSTALLED
	@echo "Installing protoc grpc gateway plugin to gobin"
	@go mod tidy && go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
endif
ifndef GEN_OPENAPI_INSTALLED
	@echo "Installing protoc openapi plugin to gobin"
	@go mod tidy && go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
endif
ifndef GEN_GO_GRPC_INSTALLED
	@echo "Installing protoc go-grpc plugin to gobin"
	@go mod tidy && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
endif
ifndef GEN_GO_INSTALLED
	@echo "Installing protoc go plugin to gobin"
	@go mod tidy && go install google.golang.org/protobuf/cmd/protoc-gen-go
endif
#run buf from api dir when everything is ok
	@echo "Generating grpc gateway code from .proto files"
	@cd api && buf generate

# generate go code and openAPI v2 spec, then converts the v2 spec to v3, then validates it.
.PHONY: apigen
apigen: buf-gen apigen-v3 apigen-validate

# remove all generated files
.PHONY: clean
clean: tls-delete apigen-v2-delete binary-delete
	@echo "All generated artifacts removed."

# run go linter with the repositories lint config
.PHONY: lint
lint:
	@echo ""
	@echo "Linting code."
	@$(DOCKER) run -t --rm -v $(PWD):/app -w /app golangci/golangci-lint golangci-lint run -v

# run short subset of tests
.PHONY: test-short
test-short:
	$(GO) test -short $(PWD)/...

# run all tests
.PHONY: test
test:
	$(GO) test $(PWD)/...

# mimics the CI that runs on PR
.PHONY: pr-check
pr-check: gmtidy arch-check test lint binary

# runs go mod tidy
.PHONY: gmtidy
gmtidy:
	@echo "Tidying dependencies using go mod tidy..."
	@$(GO) mod tidy

# describes current architectural rules setup in arch-go.yml
.PHONY: arch-describe
arch-describe:
	@echo "Current architecture rules:"
	@$(DOCKER) run --rm -v $(PWD):/app -w /app quay.io/archgo/arch-go-test:latest describe

# checks if architectural rules are met
.PHONY: arch-check
arch-check:
	@echo "Checking changes against architecture rules defined in arch-go.yml:"
	@$(DOCKER) run --rm -v $(PWD):/app -w /app quay.io/archgo/arch-go-test:latest
