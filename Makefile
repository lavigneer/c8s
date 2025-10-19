# C8S Kubernetes-Native CI System
# Makefile for common development operations

# Variables
PROJECT_NAME := c8s
ORG := github.com/org
MODULE := $(ORG)/$(PROJECT_NAME)

# Build configuration
BUILD_DIR := bin
CONTROLLER_BINARY := $(BUILD_DIR)/controller
API_SERVER_BINARY := $(BUILD_DIR)/api-server
WEBHOOK_BINARY := $(BUILD_DIR)/webhook
CLI_BINARY := $(BUILD_DIR)/c8s

# Docker configuration
DOCKER_REGISTRY ?= ghcr.io/org
CONTROLLER_IMAGE := $(DOCKER_REGISTRY)/c8s-controller
API_SERVER_IMAGE := $(DOCKER_REGISTRY)/c8s-api-server
WEBHOOK_IMAGE := $(DOCKER_REGISTRY)/c8s-webhook
VERSION ?= $(shell git describe --tags --always --dirty)

# Go configuration
GO := go
GOFLAGS := -v
LDFLAGS := -ldflags "-X $(MODULE)/pkg/version.Version=$(VERSION)"

# CRD and code generation
CONTROLLER_GEN := $(shell which controller-gen)
CRD_OPTIONS ?= crd:allowDangerousTypes=true

# Test configuration
TEST_TIMEOUT := 10m
COVERAGE_FILE := coverage.out

.PHONY: all
all: fmt vet test build

##@ Development

.PHONY: fmt
fmt: ## Run go fmt against code
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet against code
	$(GO) vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: test
test: ## Run unit tests
	$(GO) test -timeout $(TEST_TIMEOUT) -race -coverprofile=$(COVERAGE_FILE) ./...

.PHONY: test-unit
test-unit: ## Run unit tests only
	$(GO) test -timeout $(TEST_TIMEOUT) -race ./tests/unit/...

.PHONY: test-integration
test-integration: envtest ## Run integration tests with envtest
	KUBEBUILDER_ASSETS="$(shell setup-envtest use -p path)" $(GO) test -timeout $(TEST_TIMEOUT) ./tests/integration/...

.PHONY: test-contract
test-contract: ## Run API contract tests (requires Docker and k3d)
	$(GO) test -v -timeout 30m ./tests/contract/...

.PHONY: test-contract-short
test-contract-short: ## Run contract tests (short version)
	$(GO) test -v -short -timeout $(TEST_TIMEOUT) ./tests/contract/...

.PHONY: coverage
coverage: test ## Generate coverage report
	$(GO) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated: coverage.html"

##@ Build

.PHONY: build
build: build-controller build-api-server build-webhook build-cli ## Build all binaries

.PHONY: build-controller
build-controller: ## Build controller binary
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(CONTROLLER_BINARY) ./cmd/controller

.PHONY: build-api-server
build-api-server: ## Build API server binary
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(API_SERVER_BINARY) ./cmd/api-server

.PHONY: build-webhook
build-webhook: ## Build webhook binary
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(WEBHOOK_BINARY) ./cmd/webhook

.PHONY: build-cli
build-cli: ## Build CLI binary
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(CLI_BINARY) ./cmd/c8s

.PHONY: install
install: build-cli ## Install CLI to $GOPATH/bin
	$(GO) install $(LDFLAGS) ./cmd/c8s

.PHONY: clean
clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f $(COVERAGE_FILE) coverage.html

.PHONY: clean-all
clean-all: clean clean-clusters ## Clean everything including test clusters

##@ Docker

.PHONY: docker-build
docker-build: docker-build-controller docker-build-api-server docker-build-webhook ## Build all Docker images

.PHONY: docker-build-controller
docker-build-controller: ## Build controller Docker image
	docker build -t $(CONTROLLER_IMAGE):$(VERSION) --target controller .
	docker tag $(CONTROLLER_IMAGE):$(VERSION) $(CONTROLLER_IMAGE):latest

.PHONY: docker-build-api-server
docker-build-api-server: ## Build API server Docker image
	docker build -t $(API_SERVER_IMAGE):$(VERSION) --target api-server .
	docker tag $(API_SERVER_IMAGE):$(VERSION) $(API_SERVER_IMAGE):latest

.PHONY: docker-build-webhook
docker-build-webhook: ## Build webhook Docker image
	docker build -t $(WEBHOOK_IMAGE):$(VERSION) --target webhook .
	docker tag $(WEBHOOK_IMAGE):$(VERSION) $(WEBHOOK_IMAGE):latest

.PHONY: docker-push
docker-push: ## Push Docker images to registry
	docker push $(CONTROLLER_IMAGE):$(VERSION)
	docker push $(CONTROLLER_IMAGE):latest
	docker push $(API_SERVER_IMAGE):$(VERSION)
	docker push $(API_SERVER_IMAGE):latest
	docker push $(WEBHOOK_IMAGE):$(VERSION)
	docker push $(WEBHOOK_IMAGE):latest

##@ Code Generation

.PHONY: generate
generate: controller-gen ## Generate code (DeepCopy, client, etc.)
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: manifests
manifests: controller-gen ## Generate CRD manifests
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=controller-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

##@ Deployment

.PHONY: install-crds
install-crds: manifests ## Install CRDs to cluster
	kubectl apply -f config/crd/bases

.PHONY: uninstall-crds
uninstall-crds: manifests ## Uninstall CRDs from cluster
	kubectl delete -f config/crd/bases

.PHONY: deploy
deploy: manifests ## Deploy controller, webhook, and API server to cluster
	kubectl apply -f deploy/

.PHONY: undeploy
undeploy: ## Remove controller, webhook, and API server from cluster
	kubectl delete -f deploy/

##@ Tools

.PHONY: controller-gen
controller-gen: ## Ensure controller-gen is installed
	@which controller-gen > /dev/null || $(GO) install sigs.k8s.io/controller-tools/cmd/controller-gen@latest

.PHONY: envtest
envtest: ## Ensure setup-envtest is installed
	@which setup-envtest > /dev/null || $(GO) install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: golangci-lint
golangci-lint: ## Ensure golangci-lint is installed
	@which golangci-lint > /dev/null || $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: tools
tools: controller-gen envtest golangci-lint ## Install all development tools

.PHONY: check-deps
check-deps: ## Check if required dependencies are installed
	@echo "Checking dependencies..."
	@command -v docker >/dev/null 2>&1 || { echo "⚠ Docker is not installed"; exit 1; }
	@echo "  ✓ Docker installed"
	@docker info >/dev/null 2>&1 || { echo "⚠ Docker daemon is not running"; exit 1; }
	@echo "  ✓ Docker daemon running"
	@command -v kubectl >/dev/null 2>&1 || { echo "⚠ kubectl is not installed"; exit 1; }
	@echo "  ✓ kubectl installed"
	@command -v k3d >/dev/null 2>&1 || { echo "⚠ k3d is not installed"; exit 1; }
	@echo "  ✓ k3d installed"
	@echo "All dependencies are installed"

##@ Local Development

.PHONY: run-controller
run-controller: ## Run controller locally (requires kubeconfig)
	$(GO) run ./cmd/controller/main.go

.PHONY: run-api-server
run-api-server: ## Run API server locally
	$(GO) run ./cmd/api-server/main.go --port=8080

.PHONY: run-webhook
run-webhook: ## Run webhook server locally
	$(GO) run ./cmd/webhook/main.go --port=9443

.PHONY: dev-cluster-create
dev-cluster-create: build-cli ## Create local test cluster
	$(CLI_BINARY) dev cluster create c8s-dev

.PHONY: dev-cluster-delete
dev-cluster-delete: build-cli ## Delete local test cluster
	$(CLI_BINARY) dev cluster delete c8s-dev --force

.PHONY: dev-cluster-status
dev-cluster-status: build-cli ## Show local test cluster status
	$(CLI_BINARY) dev cluster status c8s-dev

.PHONY: dev-cluster-list
dev-cluster-list: build-cli ## List all local clusters
	$(CLI_BINARY) dev cluster list

.PHONY: dev-cluster-reset
dev-cluster-reset: dev-cluster-delete dev-cluster-create ## Reset local test cluster

.PHONY: dev-deploy
dev-deploy: build-cli ## Deploy operator to local cluster (Phase 4 - not yet implemented)
	@echo "⚠️  Operator deployment not yet implemented (Phase 4)"
	@echo "Coming soon: $(CLI_BINARY) dev deploy operator"

.PHONY: dev-test
dev-test: build-cli ## Run end-to-end tests on local cluster (Phase 5 - not yet implemented)
	@echo "⚠️  E2E testing not yet implemented (Phase 5)"
	@echo "Coming soon: $(CLI_BINARY) dev test run"

.PHONY: dev-reload
dev-reload: build-cli ## Quick iteration: rebuild CLI (deploy/test coming in later phases)
	@echo "✓ CLI rebuilt successfully"
	@echo ""
	@echo "⚠️  Full reload workflow not yet available"
	@echo "Available now:"
	@echo "  - make dev-cluster-create"
	@echo "  - make dev-cluster-status"
	@echo "  - make dev-cluster-delete"
	@echo ""
	@echo "Coming in Phase 4:"
	@echo "  - make dev-deploy (operator deployment)"
	@echo ""
	@echo "Coming in Phase 5:"
	@echo "  - make dev-test (end-to-end tests)"

.PHONY: clean-clusters
clean-clusters: ## Delete all c8s test clusters
	@k3d cluster list -o json 2>/dev/null | grep -o '"name":"c8s-[^"]*"' | cut -d'"' -f4 | xargs -I {} k3d cluster delete {} 2>/dev/null || true
	@echo "All c8s clusters deleted"

##@ Help

.PHONY: dev-help
dev-help: ## Show development commands
	@echo ""
	@echo "Development Workflow:"
	@echo "  make dev-cluster-create    # Create local test cluster"
	@echo "  make dev-deploy            # Deploy operator (Phase 4 - not yet implemented)"
	@echo "  make dev-test              # Run end-to-end tests (Phase 5 - not yet implemented)"
	@echo "  make dev-cluster-reset     # Reset cluster (delete and recreate)"
	@echo "  make dev-cluster-delete    # Delete test cluster"
	@echo ""
	@echo "Quick Iteration:"
	@echo "  make build                 # Build all binaries"
	@echo "  make test                  # Run all unit tests"
	@echo "  make test-contract-short   # Run contract tests (fast)"
	@echo "  make lint                  # Run linter"
	@echo ""
	@echo "Documentation:"
	@echo "  open docs/local-testing.md # View local testing guide"
	@echo ""

.PHONY: help
help: ## Display this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
