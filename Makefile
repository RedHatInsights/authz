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
openapi/validate: openapi-generator
	$(OPENAPI_GENERATOR) validate -i api/v1alpha/openapi-authz-v1_alpha.yaml
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
	@go mod tidy && install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
endif
ifndef GEN_GO_GRPC_INSTALLED
	@echo "Installing protoc go-grpc plugin to gobin"
	@go mod tidy && install google.golang.org/grpc/cmd/protoc-gen-go-grpc
endif
ifndef GEN_GO_INSTALLED
	@echo "Installing protoc go plugin to gobin"
	@go mod tidy && install google.golang.org/protobuf/cmd/protoc-gen-go
endif
#run buf from api dir when everything is ok
	@echo "Generating grpc gateway code from .proto files"
	@cd api && buf generate

.PHONY: clean-tls
clean-tls:
	@echo "removing generated tls certs"
	@rm -rf tls/

.PHONY: clean
clean-all:
	@echo "removing all generated artifacts "
	@rm -rf tls/
