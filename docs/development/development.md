# Development Workflow

This guide describes the development workflow for contributing to C8S.

## Prerequisites

- Go 1.25+
- Docker (for building images)
- kubectl with access to a Kubernetes cluster (1.24+)
- make

## Initial Setup

```bash
# Clone repository
git clone https://github.com/org/c8s.git
cd c8s

# Install dependencies
go mod download

# Install development tools (controller-gen, envtest, golangci-lint)
make tools

# Verify installation
make fmt vet lint
```

## Development Cycle

### 1. Make Changes

Edit source files in `cmd/` or `pkg/` directories.

### 2. Run Tests

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Generate coverage report
make coverage
```

### 3. Format and Lint

```bash
# Format code
make fmt

# Run go vet
make vet

# Run golangci-lint
make lint
```

### 4. Generate Code

If you modify CRD types in `pkg/apis/v1alpha1/`:

```bash
# Generate DeepCopy methods
make generate

# Generate CRD manifests
make manifests

# Verify changes
git diff config/crd/bases/
```

### 5. Build Binaries

```bash
# Build all binaries
make build

# Build specific binary
make build-controller
make build-api-server
make build-webhook
make build-cli

# Install CLI to $GOPATH/bin
make install
```

### 6. Build Docker Images

```bash
# Build all images
make docker-build

# Build specific image
make docker-build-controller
make docker-build-api-server
make docker-build-webhook

# Push to registry
make docker-push
```

## Running Locally

### Option 1: Run Components Locally (Recommended for Development)

```bash
# Terminal 1: Install CRDs to cluster
make install-crds

# Terminal 2: Run controller locally
make run-controller

# Terminal 3: Run API server locally
make run-api-server

# Terminal 4: Run webhook service locally
make run-webhook
```

This runs the C8S components locally while connecting to a real Kubernetes cluster. Jobs will still be created in the cluster.

### Option 2: Deploy to Cluster

```bash
# Build and push images
make docker-build docker-push

# Deploy to cluster
make deploy

# Check deployment
kubectl get pods -n c8s-system

# View logs
kubectl logs -n c8s-system -l app=c8s-controller -f
```

## Testing Changes

### Unit Tests

Unit tests are in `tests/unit/`. They test individual functions without external dependencies.

```bash
# Run unit tests
make test-unit

# Run specific test
go test ./tests/unit/... -run TestScheduler
```

### Integration Tests

Integration tests are in `tests/integration/`. They use envtest to run against a real Kubernetes API.

```bash
# Run integration tests
make test-integration

# Run specific test
KUBEBUILDER_ASSETS="$(shell setup-envtest use -p path)" go test ./tests/integration/... -run TestControllerReconcile
```

### Contract Tests

Contract tests are in `tests/contract/`. They validate the REST API contracts.

```bash
# Run contract tests
make test-contract
```

### End-to-End Manual Testing

1. Create a test pipeline:

```bash
kubectl apply -f config/samples/pipelineconfig_example.yaml
```

2. Trigger a pipeline run:

```bash
kubectl apply -f config/samples/pipelinerun_example.yaml
```

3. Watch execution:

```bash
kubectl get pipelinerun -w
kubectl get jobs
kubectl logs -f job/example-go-pipeline-abc123-test
```

## Debugging

### View Controller Logs

```bash
# If running locally
# Logs appear in terminal where you ran `make run-controller`

# If deployed to cluster
kubectl logs -n c8s-system -l app=c8s-controller -f
```

### View API Server Logs

```bash
# If running locally
# Logs appear in terminal where you ran `make run-api-server`

# If deployed to cluster
kubectl logs -n c8s-system -l app=c8s-api -f
```

### Debug CRD Issues

```bash
# Get CRD definitions
kubectl get crd pipelineconfigs.c8s.dev -o yaml
kubectl get crd pipelineruns.c8s.dev -o yaml

# Describe a resource to see events
kubectl describe pipelinerun example-go-pipeline-abc123

# Check validation errors
kubectl get pipelineconfig example-go-pipeline -o yaml
```

### Debug Job Issues

```bash
# List jobs created by controller
kubectl get jobs -l c8s.dev/pipeline-run=example-go-pipeline-abc123

# Describe job to see events
kubectl describe job example-go-pipeline-abc123-test

# View pod logs
kubectl logs job/example-go-pipeline-abc123-test
```

## Making Changes to APIs

When modifying CRD types (`pkg/apis/v1alpha1/*.go`):

1. **Edit the Go structs** with your changes

2. **Add kubebuilder markers** for validation:
   ```go
   // +kubebuilder:validation:Required
   // +kubebuilder:validation:MinLength=1
   Name string `json:"name"`
   ```

3. **Generate manifests and code**:
   ```bash
   make generate manifests
   ```

4. **Update samples** in `config/samples/` if needed

5. **Add tests** for new fields

6. **Run tests**:
   ```bash
   make test
   ```

7. **Commit changes** including generated files:
   ```bash
   git add pkg/apis/ config/crd/bases/
   git commit -m "Add new field X to PipelineConfig"
   ```

## Common Issues

### "controller-gen: command not found"

```bash
make controller-gen
# or
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
```

### "setup-envtest: command not found"

```bash
make envtest
# or
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
```

### Integration tests failing with "no matches for kind"

CRDs may not be installed or generated:

```bash
make manifests
make install-crds
```

### "error: unable to recognize \"config/crd/bases/...\": no matches for kind"

CRDs may have validation errors. Check with:

```bash
kubectl apply --dry-run=server -f config/crd/bases/
```

## Code Style Guidelines

- Follow standard Go conventions (gofmt, goimports)
- Write descriptive comments for exported functions
- Keep functions focused and testable
- Use meaningful variable names
- Prefer standard library over external dependencies
- Add tests for new functionality

## Git Workflow

1. Create a feature branch:
   ```bash
   git checkout -b feature/my-feature
   ```

2. Make changes and commit:
   ```bash
   git add .
   git commit -m "Add feature X"
   ```

3. Run tests before pushing:
   ```bash
   make test lint
   ```

4. Push and create PR:
   ```bash
   git push origin feature/my-feature
   # Create PR on GitHub
   ```

## CI/CD

GitHub Actions runs on every push and PR:

- **Lint**: golangci-lint checks
- **Test**: Unit and integration tests
- **Build**: All binaries and Docker images
- **Verify**: CRD manifests are up-to-date

Check `.github/workflows/ci.yaml` for details.

## Release Process

(To be defined as project matures)

1. Create release branch: `git checkout -b release-v0.1.0`
2. Update version numbers
3. Generate changelog
4. Tag release: `git tag v0.1.0`
5. Build and push release images
6. Create GitHub release

## Getting Help

- **GitHub Discussions**: https://github.com/org/c8s/discussions
- **GitHub Issues**: https://github.com/org/c8s/issues
- **Slack**: https://c8s.slack.com

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for contribution guidelines.
