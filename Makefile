# C8S Kubernetes-Native CI System
# Makefile for common development operations

# Variables
PROJECT_NAME := c8s
ORG := github.com/lavigneer
MODULE := $(ORG)/$(PROJECT_NAME)

# Build configuration
BUILD_DIR := bin
CONTROLLER_BINARY := $(BUILD_DIR)/controller
WEBHOOK_BINARY := $(BUILD_DIR)/webhook
CLI_BINARY := $(BUILD_DIR)/c8s

# Docker configuration
DOCKER_REGISTRY ?= ghcr.io/org
CONTROLLER_IMAGE := $(DOCKER_REGISTRY)/c8s-controller
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
	go tool golangci-lint run

.PHONY: test
test: ## Run unit tests
	$(GO) test -timeout $(TEST_TIMEOUT) -race -coverprofile=$(COVERAGE_FILE) ./...

.PHONY: test-unit
test-unit: ## Run unit tests only
	$(GO) test -timeout $(TEST_TIMEOUT) -race ./tests/unit/...

.PHONY: test-integration
test-integration: envtest ## Run integration tests with envtest
	KUBEBUILDER_ASSETS="$(shell setup-envtest use -p path)" $(GO) test -timeout $(TEST_TIMEOUT) ./tests/integration/...

.PHONY: coverage
coverage: test ## Generate coverage report
	$(GO) tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@echo "Coverage report generated: coverage.html"

##@ Build

.PHONY: build
build: build-controller build-webhook build-cli ## Build all binaries

.PHONY: build-controller
build-controller: ## Build controller binary
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(CONTROLLER_BINARY) ./cmd/controller

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
docker-build: docker-build-controller docker-build-webhook ## Build all Docker images

.PHONY: docker-build-controller
docker-build-controller: ## Build controller Docker image
	docker build -t $(CONTROLLER_IMAGE):$(VERSION) --target controller .
	docker tag $(CONTROLLER_IMAGE):$(VERSION) $(CONTROLLER_IMAGE):latest

.PHONY: docker-build-webhook
docker-build-webhook: ## Build webhook Docker image
	docker build -t $(WEBHOOK_IMAGE):$(VERSION) --target webhook .
	docker tag $(WEBHOOK_IMAGE):$(VERSION) $(WEBHOOK_IMAGE):latest

.PHONY: docker-push
docker-push: ## Push Docker images to registry
	docker push $(CONTROLLER_IMAGE):$(VERSION)
	docker push $(CONTROLLER_IMAGE):latest
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
deploy: manifests ## Deploy controller and webhook to cluster
	kubectl create namespace c8s-system --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deploy/install.yaml
	kubectl apply -f deploy/webhook-deployment.yaml
	kubectl apply -f deploy/webhook-service.yaml
	kubectl apply -f deploy/webhook-ingress.yaml

.PHONY: undeploy
undeploy: ## Remove controller and webhook from cluster
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

##@ E2E Testing

.PHONY: test-e2e
test-e2e: ## Run all E2E tests
	devbox run -- npm run test:e2e

.PHONY: test-e2e-ui
test-e2e-ui: ## Run E2E tests with interactive UI
	devbox run -- npm run test:e2e:ui

.PHONY: test-e2e-debug
test-e2e-debug: ## Run E2E tests with Playwright debugger
	devbox run -- npm run test:e2e:debug

.PHONY: test-e2e-report
test-e2e-report: ## View E2E test HTML report
	devbox run -- npm run test:e2e:report

.PHONY: test-e2e-install
test-e2e-install: ## Install E2E test dependencies and browser binaries
	devbox run -- npm install
	devbox run -- npx playwright install

##@ Local Development

.PHONY: run-controller
run-controller: ## Run controller locally (requires kubeconfig)
	$(GO) run ./cmd/controller/main.go

.PHONY: run-webhook
run-webhook: ## Run webhook server locally
	$(GO) run ./cmd/webhook/main.go --port=9443


##@ Tilt (Local K8s Development)

.PHONY: tilt-up
tilt-up: ## Start Tilt with local K8s development environment (creates cluster if needed)
	@command -v tilt >/dev/null 2>&1 || { echo "⚠ Tilt is not installed. Install from https://docs.tilt.dev/install.html"; exit 1; }
	@command -v k3d >/dev/null 2>&1 || { echo "⚠ k3d is not installed"; exit 1; }
	@echo "Creating k3d cluster (if it doesn't exist)..."
	@k3d cluster get c8s-dev > /dev/null 2>&1 || k3d cluster create c8s-dev --registry-create=registry:5000 -p "8080:80@loadbalancer" --servers 1 --agents 2
	@echo "Starting Tilt..."
	tilt up

.PHONY: tilt-down
tilt-down: ## Stop Tilt development environment
	@command -v tilt >/dev/null 2>&1 || { echo "⚠ Tilt is not installed"; exit 1; }
	tilt down

.PHONY: tilt-logs
tilt-logs: ## View Tilt logs
	@command -v tilt >/dev/null 2>&1 || { echo "⚠ Tilt is not installed"; exit 1; }
	tilt logs

.PHONY: tilt-status
tilt-status: ## Check Tilt status
	@command -v tilt >/dev/null 2>&1 || { echo "⚠ Tilt is not installed"; exit 1; }
	tilt status

.PHONY: clean-clusters
clean-clusters: ## Delete all c8s test clusters
	@k3d cluster list -o json 2>/dev/null | grep -o '"name":"c8s-[^"]*"' | cut -d'"' -f4 | xargs -I {} k3d cluster delete {} 2>/dev/null || true
	@echo "All c8s clusters deleted"

##@ Tilt CI

.PHONY: tilt-ci-local
tilt-ci-local: ## Run tilt ci locally with kind cluster
	@bash scripts/tilt-ci-local.sh

.PHONY: tilt-ci-clean
tilt-ci-clean: ## Clean up kind cluster from tilt ci
	@command -v kind >/dev/null 2>&1 || { echo "⚠ kind is not installed"; exit 1; }
	@kind get clusters | grep -q "c8s-ci" && kind delete cluster --name c8s-ci || echo "No c8s-ci cluster found"
	@echo "Cleaned up tilt ci cluster"

##@ Help

.PHONY: dev-help
dev-help: ## Show development commands
	@echo ""
	@echo "Development Workflow (using Tilt):"
	@echo "  make tilt-up               # Start Tilt development environment"
	@echo "  make tilt-down             # Stop Tilt"
	@echo "  make tilt-logs             # View Tilt logs"
	@echo ""
	@echo "Quick Iteration:"
	@echo "  make build                 # Build all binaries"
	@echo "  make test                  # Run all unit tests"
	@echo "  make test-integration      # Run integration tests"
	@echo "  make lint                  # Run linter"
	@echo ""
	@echo "Manual Development (without Tilt):"
	@echo "  make run-controller        # Run controller locally"
	@echo "  make run-webhook           # Run webhook locally"
	@echo ""
	@echo "Cluster Management:"
	@echo "  make clean-clusters        # Delete all c8s test clusters"
	@echo ""

.PHONY: help
help: ## Display this help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
