# C8S Cheat Sheet

Quick reference for common C8S operations.

## 🚀 Quick Start

```bash
# Start local development
tilt up

# Or with Helm
helm install c8s ./chart/c8s -n c8s-system --create-namespace
```

## 📝 Pipeline Configuration

### Minimal Pipeline

```yaml
version: v1alpha1
name: my-pipeline
steps:
  - name: test
    image: golang:1.25
    commands:
      - go test ./...
```

### With Dependencies

```yaml
steps:
  - name: test
    image: golang:1.25
    commands: [go test ./...]

  - name: build
    image: golang:1.25
    commands: [go build -o app]
    dependsOn: [test]  # Runs after test
```

### With Resources

```yaml
steps:
  - name: build
    image: golang:1.25
    commands: [go build]
    resources:
      cpu: 2000m      # 2 CPU cores
      memory: 4Gi     # 4 GB RAM
```

### With Secrets

```yaml
steps:
  - name: deploy
    image: ubuntu:22.04
    commands: [./deploy.sh]
    secrets:
      - secretRef: my-secret
        key: API_TOKEN
        envVar: API_TOKEN
```

### Conditional Execution

```yaml
steps:
  - name: deploy
    image: ubuntu:22.04
    commands: [./deploy.sh]
    conditional:
      branch: "main"      # Only on main branch
      onSuccess: true     # Only if previous steps succeeded
```

### Matrix Builds

```yaml
matrix:
  dimensions:
    go_version: ["1.24", "1.25"]
    os: ["ubuntu", "alpine"]

steps:
  - name: test
    image: golang:${{ matrix.go_version }}-${{ matrix.os }}
    commands: [go test ./...]
```

## 🛠️ CLI Commands

### Run Pipeline

```bash
# Trigger pipeline
c8s run my-pipeline \
  --commit=$(git rev-parse HEAD) \
  --branch=$(git branch --show-current)
```

### View Pipelines

```bash
# List all pipelines
c8s pipelines list

# Get details
c8s pipelines get my-pipeline

# Delete pipeline
c8s pipelines delete my-pipeline
```

### View Runs

```bash
# List runs
c8s runs list

# List runs for specific pipeline
c8s runs list --pipeline=my-pipeline

# Get run details
c8s runs get my-pipeline-xxxxx

# Watch runs
c8s runs list --watch
```

### Logs

```bash
# Follow logs
c8s logs my-pipeline-xxxxx --follow

# View specific step
c8s logs my-pipeline-xxxxx --step=build

# Save to file
c8s logs my-pipeline-xxxxx > build.log
```

### Artifacts

```bash
# List artifacts
c8s artifacts list my-pipeline-xxxxx

# Download artifact
c8s artifacts get my-pipeline-xxxxx app.tar.gz

# Download all
c8s artifacts download my-pipeline-xxxxx --all
```

## 🐳 kubectl Commands

### View Resources

```bash
# PipelineConfigs
kubectl get pipelineconfigs
kubectl get pc  # Short form

# PipelineRuns
kubectl get pipelineruns
kubectl get pr  # Short form

# Jobs (pipeline steps)
kubectl get jobs -l c8s.dev/pipeline=my-pipeline

# Pods
kubectl get pods -l c8s.dev/run=my-pipeline-xxxxx
```

### View Details

```bash
# PipelineConfig
kubectl describe pipelineconfig my-pipeline

# PipelineRun
kubectl describe pipelinerun my-pipeline-xxxxx

# Logs from specific step
kubectl logs job/my-pipeline-xxxxx-test
```

### Create from File

```bash
# Create PipelineConfig
kubectl apply -f .c8s.yaml

# Trigger pipeline run
kubectl apply -f - <<EOF
apiVersion: c8s.dev/v1alpha1
kind: PipelineRun
metadata:
  name: my-pipeline-$(date +%s)
spec:
  pipelineConfigRef: my-pipeline
  commit: abc123
  branch: main
EOF
```

### Delete Resources

```bash
# Delete PipelineConfig
kubectl delete pipelineconfig my-pipeline

# Delete PipelineRun
kubectl delete pipelinerun my-pipeline-xxxxx

# Delete all runs for pipeline
kubectl delete pr -l c8s.dev/pipeline=my-pipeline
```

## 🔨 Development

### Local Development

```bash
# Start Tilt
tilt up

# Open Tilt dashboard
# http://localhost:10350

# Edit code - auto-rebuilds!
vim cmd/controller/main.go
```

### Build

```bash
# Build all binaries
make build

# Build specific binary
make build-controller
make build-webhook
make build-cli

# Build Docker images
make docker-build
```

### Testing

```bash
# Run all tests
make test-all

# Unit tests only
make test-unit

# Integration tests
make test-integration

# E2E tests
make test-e2e

# With coverage
make coverage
open coverage.html
```

### Code Generation

```bash
# Generate CRDs
make manifests

# Generate DeepCopy methods
make generate

# Install CRDs to cluster
make install-crds
```

### Linting

```bash
# Run linter
make lint

# Format code
make fmt

# Vet code
make vet
```

## 🔍 Debugging

### Controller Issues

```bash
# Controller logs
kubectl logs -f deployment/c8s-controller -n c8s-system

# Controller events
kubectl get events -n c8s-system --sort-by='.lastTimestamp'

# Controller status
kubectl get deployment c8s-controller -n c8s-system -o wide
```

### Pipeline Issues

```bash
# PipelineRun status
kubectl get pr my-pipeline-xxxxx -o yaml

# Step Jobs
kubectl get jobs -l c8s.dev/run=my-pipeline-xxxxx

# Step Pods
kubectl get pods -l c8s.dev/run=my-pipeline-xxxxx

# Step logs
kubectl logs -l c8s.dev/run=my-pipeline-xxxxx --all-containers

# Failed pods
kubectl get pods -l c8s.dev/run=my-pipeline-xxxxx --field-selector=status.phase=Failed
```

### Webhook Issues

```bash
# Webhook logs
kubectl logs -f deployment/c8s-webhook -n c8s-system

# Webhook service
kubectl get svc c8s-webhook -n c8s-system

# Test webhook locally
curl -X POST http://localhost:9090/webhook/github \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -d @payload.json
```

### API Server Issues

```bash
# API server logs
kubectl logs -f deployment/c8s-api-server -n c8s-system

# API server service
kubectl get svc c8s-api-server -n c8s-system

# Port forward for local testing
kubectl port-forward svc/c8s-api-server 8080:8080 -n c8s-system

# Test API
curl http://localhost:8080/api/v1/pipelines
```

## 🌐 API Endpoints

### Pipelines

```bash
# List pipelines
GET /api/v1/pipelines

# Get pipeline
GET /api/v1/pipelines/:name

# Create pipeline
POST /api/v1/pipelines

# Delete pipeline
DELETE /api/v1/pipelines/:name
```

### Runs

```bash
# List runs
GET /api/v1/runs

# Get run
GET /api/v1/runs/:id

# Trigger run
POST /api/v1/runs

# Stream logs (SSE)
GET /api/v1/runs/:id/logs
GET /api/v1/runs/:id/logs?step=build
```

### Artifacts

```bash
# List artifacts
GET /api/v1/runs/:id/artifacts

# Download artifact
GET /api/v1/runs/:id/artifacts/:name
```

## 🔐 Secrets Management

### Create Secret

```bash
# Create secret
kubectl create secret generic my-secret \
  --from-literal=API_TOKEN=abc123 \
  -n c8s-system

# From file
kubectl create secret generic my-secret \
  --from-file=key.json \
  -n c8s-system

# Docker registry secret
kubectl create secret docker-registry docker-creds \
  --docker-server=ghcr.io \
  --docker-username=myuser \
  --docker-password=mypass \
  -n c8s-system
```

### Use in Pipeline

```yaml
steps:
  - name: deploy
    image: ubuntu:22.04
    commands:
      - echo "Token: $API_TOKEN"
    secrets:
      - secretRef: my-secret
        key: API_TOKEN
        envVar: API_TOKEN
```

## 📦 Helm

### Install

```bash
# Install C8S
helm install c8s ./chart/c8s -n c8s-system --create-namespace

# With custom values
helm install c8s ./chart/c8s \
  -n c8s-system \
  -f my-values.yaml

# Dry run
helm install c8s ./chart/c8s \
  -n c8s-system \
  --dry-run --debug
```

### Upgrade

```bash
# Upgrade
helm upgrade c8s ./chart/c8s -n c8s-system

# With new values
helm upgrade c8s ./chart/c8s \
  -n c8s-system \
  --set controller.replicas=3
```

### Uninstall

```bash
# Uninstall
helm uninstall c8s -n c8s-system

# Keep CRDs
helm uninstall c8s -n c8s-system --keep-history
```

## 📊 Monitoring

### Resource Usage

```bash
# Pod resources
kubectl top pods -n c8s-system

# Node resources
kubectl top nodes

# Pipeline step resources
kubectl top pods -l c8s.dev/run=my-pipeline-xxxxx
```

### Metrics

```bash
# Prometheus metrics (if enabled)
curl http://c8s-api-server:8080/metrics
curl http://c8s-controller:8080/metrics
```

## 🔄 Environment Variables

### Controller

```bash
KUBECONFIG=/path/to/config    # Kubeconfig path
NAMESPACE=c8s-system          # Namespace to watch
LOG_LEVEL=debug               # Log level
S3_ENDPOINT=http://minio:9000 # S3 endpoint
S3_BUCKET=c8s-logs            # S3 bucket name
```

### API Server

```bash
PORT=8080                     # HTTP port
KUBECONFIG=/path/to/config    # Kubeconfig path
S3_ENDPOINT=http://minio:9000 # S3 endpoint
S3_BUCKET=c8s-logs            # S3 bucket
JWT_SECRET=your-secret        # JWT secret
```

### CLI

```bash
C8S_ENDPOINT=http://localhost:8080  # API endpoint
C8S_TOKEN=your-token                 # Auth token
C8S_OUTPUT=json                      # Output format
```

## 📚 Resources

- **Documentation**: [docs/](./docs/)
- **Examples**: [examples/](./examples/)
- **API Spec**: [specs/001-build-a-continuous/contracts/openapi.yaml](./specs/001-build-a-continuous/contracts/openapi.yaml)
- **Troubleshooting**: [docs/guides/troubleshooting.md](./docs/guides/troubleshooting.md)

---

**More Help**: See [SUPPORT.md](./.github/SUPPORT.md) for getting support and asking questions.
