# C8S - Kubernetes-Native Continuous Integration

[![Go Version](https://img.shields.io/badge/go-1.25-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![CI](https://github.com/org/c8s/workflows/CI/badge.svg)](https://github.com/org/c8s/actions)
[![Documentation](https://img.shields.io/badge/docs-latest-brightgreen.svg)](./docs/README.md)
[![Tests](https://img.shields.io/badge/tests-200%2B-success.svg)](./tests/README.md)

C8S is a Kubernetes-native continuous integration system that runs pipeline steps as isolated container Jobs. It leverages Kubernetes primitives (CRDs, Jobs, Pods) for orchestration, scheduling, and state management, providing a scalable and secure CI platform.

## Quick Links

📚 **Documentation**: [Getting Started](./docs/guides/getting-started.md) • [Architecture](./docs/guides/architecture.md) • [API Reference](./specs/001-build-a-continuous/contracts/openapi.yaml)

🛠️ **Development**: [5-Min Setup](./docs/development/QUICKSTART.md) • [Tilt Workflow](./docs/development/TILT-WORKFLOW.md) • [Contributing](./docs/CONTRIBUTING.md)

🧪 **Testing**: [Test Guide](./tests/README.md) • [E2E Tests](./tests/e2e/README.md) • [Coverage](./coverage.html)

🚀 **Deployment**: [Helm Chart](./chart/c8s/README.md) • [Operator Guide](./docs/operations/operator-guide.md)

## Features

- **Kubernetes-Native**: Built entirely on Kubernetes primitives (CRDs, Jobs, Pods)
- **Isolated Execution**: Each pipeline step runs in its own Kubernetes Job with isolated resources
- **Declarative Pipelines**: YAML-based pipeline definitions with GitOps-friendly configuration
- **DAG Scheduling**: Automatic dependency resolution and parallel step execution
- **Git Integration**: Webhook support for GitHub, GitLab, and Bitbucket
- **Secure Secrets**: Native Kubernetes Secret integration with automatic log masking
- **Object Storage**: S3-compatible storage for logs and build artifacts
- **Resource Limits**: CPU/memory quotas and namespace-scoped access control
- **Real-Time Logs**: Streaming logs via CLI, API, and optional web dashboard
- **Matrix Builds**: Run parallel pipelines across multiple configurations
- **Conditional Execution**: Branch and tag-based conditional steps

## Architecture

```
Developer pushes code
    ↓
GitHub webhook → C8S Webhook Service
    ↓
Creates PipelineRun CRD
    ↓
Controller watches PipelineRun
    ↓
Creates Kubernetes Jobs (one per step)
    ↓
Jobs run in isolated Pods
    ↓
Logs streamed to object storage
    ↓
Status updated in PipelineRun
    ↓
Developer views results via CLI/API/Dashboard
```

### Components

- **Controller**: Watches PipelineRun CRDs, creates Jobs, updates status
- **API Server**: REST API for pipeline management and log retrieval
- **Webhook Service**: Receives Git webhooks, creates PipelineRun resources
- **CLI**: Command-line tool for triggering pipelines and viewing logs
- **Dashboard** (optional): HTMX-based web UI for visual pipeline monitoring

## Quick Start

**For Local Development** (Recommended): Use **Tilt** for rapid local Kubernetes development with live reload:

### Local Development with Tilt (Primary Workflow)

```bash
# Install prerequisites (if needed)
brew install tilt k3d kubectl

# Start development environment (one command!)
tilt up

# Tilt will:
# 1. Create a local k3d cluster
# 2. Install CRDs and configuration
# 3. Build all components
# 4. Deploy controller, api-server, webhook
# 5. Open dashboard at http://localhost:10350

# Edit Go code - changes auto-rebuild and redeploy!
vim cmd/controller/main.go  # Save → auto-rebuild in ~30 seconds

# View logs in Tilt dashboard or via kubectl
kubectl logs -f deployment/c8s-controller -n c8s-system
```

See [docs/development/tilt-setup.md](docs/development/tilt-setup.md) for detailed setup instructions.

**Requirements for local development:**
- Docker (27.3.1+)
- k3d (5.8.3+) or kind
- kubectl (1.28+)
- Go 1.25+ (for building from source)
- Tilt (for live development)

**For Manual Installation or Production Deployment**: See [quickstart.md](./specs/001-build-a-continuous/quickstart.md) for Helm chart installation and manual kubectl deployment.

### Create Pipeline

Create `.c8s.yaml` in your repository:

```yaml
version: v1alpha1
name: my-pipeline
steps:
  - name: test
    image: golang:1.21
    commands:
      - go test ./...
    resources:
      cpu: 1000m
      memory: 2Gi

  - name: build
    image: golang:1.21
    commands:
      - go build -o app
    dependsOn: [test]
    artifacts:
      - app
```

### Run Pipeline

```bash
# Via CLI
c8s run my-pipeline --commit=$(git rev-parse HEAD) --branch=$(git branch --show-current)

# Watch logs
c8s logs my-pipeline-xxxxx --follow
```

## Development

### Prerequisites

**Option 1: Using Devbox (Recommended)**
- [Devbox](https://www.jetify.com/devbox) (install: `curl -fsSL https://get.jetify.com/devbox | bash`)
- Docker (for building images and kind clusters)

**Option 2: Manual Setup**
- Go 1.25+
- Kubernetes cluster (1.24+)
- kubectl with cluster access
- Docker (for building images)

### Setup

**With Devbox (Recommended)**:
```bash
# Clone repository
git clone https://github.com/org/c8s.git
cd c8s

# Enter development environment (installs all tools automatically)
devbox shell

# Run tests
make test

# Build binaries
make build
```

See [docs/devbox-setup.md](./docs/devbox-setup.md) for detailed devbox usage.

**Manual Setup**:
```bash
# Clone repository
git clone https://github.com/org/c8s.git
cd c8s

# Install dependencies
go mod download

# Install development tools
make tools

# Run tests
make test

# Build binaries
make build
```

### Running Locally

```bash
# Install CRDs to cluster
make install-crds

# Run controller locally (requires kubeconfig)
make run-controller

# In another terminal, run API server
make run-api-server

# In another terminal, run webhook service
make run-webhook
```

### Code Generation

```bash
# Generate CRD manifests
make manifests

# Generate DeepCopy methods
make generate
```

## Project Structure

```
c8s/
├── cmd/
│   ├── controller/       # Controller main
│   ├── api-server/       # API server main (includes templates/ and static/)
│   ├── webhook/          # Webhook service main
│   └── c8s/              # CLI main
├── pkg/
│   ├── apis/v1alpha1/    # CRD definitions
│   ├── controller/       # Controller logic
│   ├── parser/           # Pipeline YAML parser
│   ├── scheduler/        # DAG scheduler
│   ├── storage/          # S3 log storage
│   ├── webhook/          # Git webhook handlers
│   ├── api/              # REST API handlers
│   ├── cli/              # CLI commands
│   └── secrets/          # Secret management
├── config/
│   ├── crd/bases/        # Generated CRD YAML
│   ├── rbac/             # RBAC manifests
│   └── samples/          # Sample CRs
├── tests/
│   ├── unit/             # Unit tests
│   ├── integration/      # Integration tests
│   └── e2e/              # Playwright E2E tests
├── chart/c8s/            # Helm chart for deployment
└── deploy/               # Kubernetes manifests (legacy - use Helm + Tilt)
```

## API Reference

REST API documentation available at `/api/v1/docs` when API server is running with `--enable-docs` flag.

See [openapi.yaml](./specs/001-build-a-continuous/contracts/openapi.yaml) for complete API specification.

## Pipeline Configuration Schema

See [pipeline-config-schema.json](./specs/001-build-a-continuous/contracts/pipeline-config-schema.json) for YAML validation schema.

## Examples

See [examples/](./examples/) for complete pipeline configurations:

- **[Simple Go Pipeline](./examples/simple-go-pipeline.yaml)** - Basic test and build
- **[Docker Build Pipeline](./examples/docker-build-pipeline.yaml)** - Container build and push
- **[Matrix Pipeline](./examples/matrix-pipeline.yaml)** - Multi-version/platform testing
- **[Monorepo Pipeline](./examples/monorepo-pipeline.yaml)** - Multi-service builds

Full examples guide with patterns and best practices: [examples/README.md](./examples/README.md)

### Quick Example: Multi-Step Pipeline

```yaml
version: v1alpha1
name: test-build-deploy
steps:
  - name: test
    image: golang:1.25
    commands:
      - go test ./...

  - name: build
    image: golang:1.25
    commands:
      - go build -o app
    dependsOn: [test]
    artifacts:
      - app

  - name: deploy
    image: ubuntu:22.04
    commands:
      - ./deploy.sh production
    dependsOn: [build]
    conditional:
      branch: "main"
```

## Contributing

Contributions are welcome! Please read [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

## License

Apache License 2.0 - See [LICENSE](LICENSE) for details.

## Documentation

📖 **[Complete Documentation Index](./docs/README.md)** - Full documentation catalog

### Essential Resources

**Getting Started**
- [Getting Started Guide](./docs/guides/getting-started.md) - Installation and first pipeline
- [Pipeline Syntax](./docs/guides/pipeline-syntax.md) - YAML configuration reference
- [Examples](./examples/README.md) - Pipeline configuration examples

**Development**
- [5-Minute Quick Start](./docs/development/QUICKSTART.md) - Get developing fast
- [Tilt Workflow Guide](./docs/development/TILT-WORKFLOW.md) - Primary development workflow
- [Testing Guide](./tests/README.md) - Comprehensive test documentation

**Architecture & Reference**
- [System Architecture](./docs/guides/architecture.md) - Complete architecture guide
- [Package Documentation](./pkg/README.md) - Go package structure
- [Specs Framework](./specs/README.md) - Feature specification methodology

**Deployment & Operations**
- [Operator Guide](./docs/operations/operator-guide.md) - Production deployment
- [Helm Chart](./chart/c8s/README.md) - Kubernetes deployment
- [Security Policy](./SECURITY.md) - Security practices and vulnerability reporting

## Community

- **GitHub Issues**: https://github.com/org/c8s/issues
- **Contributing**: See [CONTRIBUTING.md](CONTRIBUTING.md)
