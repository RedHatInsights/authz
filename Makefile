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

.PHONY: tls-cert
tls-cert:
	@echo "creating directory tls/"
	@mkdir -p tls
	@echo "Generating self-signed TLS certs. needs openssl 1.1.1 or newer..."
	@openssl req -x509 -newkey rsa:4096 -sha256 -days 365 -nodes \
               -keyout tls/tls.key -out tls/tls.crt -subj "/CN=localhost" \
               -addext "subjectAltName=DNS:example.com,DNS:www.example.net,IP:10.0.0.1"
	@echo "Success! find your cert files in the tls/ folder"


