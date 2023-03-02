GO := go
GOFMT := gofmt
PROTOC_INSTALLED := $(shell which protoc)
BUF_INSTALLED := $(shell which buf)
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell $(GO) env GOBIN))
GOBIN=$(shell $(GO) env GOPATH)/bin
else
GOBIN=$(shell $(GO) env GOBIN)
endif

DOCKER ?= docker
DOCKER_CONFIG="${PWD}/.docker"
SHELL = bash

.PHONY: binary
binary:
	$(GO) build .

.PHONY: kind-create
kind-create:
	@kind create cluster --name=authz --config=k8s/kind.yml
	@kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
	@echo "Wait for ingress with: kubectl get pods --namespace ingress-nginx"
	@echo "See k8s/README.md for more info."

.PHONY: kind-deploy
kind-deploy:
	@kubectl apply -f k8s/authz.yml
	@echo "Wait for pods with: kubectl get pods,svc"
	@echo "See k8s/README.md for more info."

.PHONY: kind-delete
kind-delete:
	@kind delete cluster --name=authz

.PHONY: tls-cert
tls-cert:
	@echo "creating directory tls/"
	@mkdir -p tls
	@echo "Generating self-signed TLS certs. needs openssl 1.1.1 or newer..."
	@openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes \
               -keyout tls/tls.key -out tls/tls.crt -subj "/CN=localhost" \
               -addext "subjectAltName=DNS:example.com,DNS:www.example.net,IP:10.0.0.1"
	@echo "Success! find your cert files in the tls/ folder"

generate:
	./scripts/generate.sh
.PHONY: generate

# validate the openapi schema
openapi/validate: 
	$(DOCKER) run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli validate -i /local/api/gen/v1/core.swagger.yaml
.PHONY: openapi/validate

# Run Swagger and host the api docs
run/docs:
	$(DOCKER) run --rm --name swagger_ui_docs -d -p 8082:8080 -e URLS="[ \
		{ url: \"/openapi/gen/v1/core.swagger.yaml\", name: \"Authz API\"}]"\
		  -v $(PWD)/api/:/usr/share/nginx/html/openapi:Z swaggerapi/swagger-ui
	@echo "Please open http://localhost:8082/"
.PHONY: run/docs


# Remove Swagger container
run/docs/teardown:
	$(DOCKER) container stop swagger_ui_docs
	$(DOCKER) container rm swagger_ui_docs
.PHONY: run/docs/teardown

# Generate grpc gateway code from proto
proto-gen-local:
	@echo "checking if protoc and buf are installed..."
ifndef PROTOC_INSTALLED
	$(error "protoc is not installed, please install it using e.g. 'brew install protobuf'")
endif
ifndef BUF_INSTALLED
	$(error "Buf is not installed, please install it using e.g. 'brew install buf'")
endif
	@echo "generating grpc gateway code from .proto"
	@cd api && buf generate
.PHONY: proto-gen-local