GO := go
GOFMT := gofmt
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell $(GO) env GOBIN))
GOBIN=$(shell $(GO) env GOPATH)/bin
else
GOBIN=$(shell $(GO) env GOBIN)
endif

DOCKER ?= docker
DOCKER_CONFIG="${PWD}/.docker"

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
	$(GO) build .
.PHONY: binary
