# cmd/

Main application entry points for C8S components.

## Overview

The `cmd/` directory contains the main entry points for all C8S executables. Each subdirectory represents a separate binary that can be built and deployed independently.

## Components

```
cmd/
├── controller/        # Pipeline orchestration controller
├── api-server/        # REST API and web dashboard
├── webhook/          # Git webhook receiver
└── c8s/              # Command-line interface (CLI)
```

---

## controller/

**Purpose**: Kubernetes controller that watches PipelineRun CRs and orchestrates pipeline execution.

**Binary**: `bin/controller`

**What It Does**:
- Watches for PipelineRun custom resources
- Creates Kubernetes Jobs for each pipeline step
- Manages step dependencies and parallel execution
- Updates PipelineRun status
- Handles retries and failure scenarios

**Key Features**:
- DAG-based scheduling (respects `dependsOn` relationships)
- Parallel step execution when no dependencies
- Resource limit enforcement (CPU, memory)
- Secret injection into pipeline steps
- Log and artifact upload to S3-compatible storage

**Running Locally**:
```bash
# Build
make build-controller

# Run (requires kubeconfig)
./bin/controller \
  --kubeconfig=$HOME/.kube/config \
  --namespace=c8s-system \
  --log-level=debug
```

**Running in Kubernetes**:
```bash
# Via Helm
helm install c8s ./chart/c8s -n c8s-system

# Via Tilt (development)
tilt up
```

**Environment Variables**:
- `KUBECONFIG` - Path to kubeconfig file (default: in-cluster config)
- `NAMESPACE` - Namespace to watch (default: all namespaces)
- `LOG_LEVEL` - Logging level (debug, info, warn, error)
- `S3_ENDPOINT` - S3-compatible storage endpoint
- `S3_BUCKET` - Bucket name for logs and artifacts

**Permissions Required** (RBAC):
- Watch/update PipelineConfigs and PipelineRuns
- Create/delete/watch Jobs
- Create/delete/watch Pods
- Get/list Secrets (for pipeline secret injection)
- Get/list ConfigMaps

**Source Files**:
```
controller/
└── main.go                # Entry point and controller setup
```

**Related**:
- [pkg/controller/](../pkg/controller/) - Controller implementation logic
- [config/manager/](../config/manager/) - Deployment manifests

---

## api-server/

**Purpose**: REST API server and optional web dashboard for pipeline management.

**Binary**: `bin/api-server`

**What It Does**:
- Exposes REST API for pipeline operations
- Serves web dashboard (HTMX-based UI)
- Streams pipeline logs from S3 storage
- Provides artifact download endpoints
- Handles authentication (JWT, API keys)

**Key Features**:
- RESTful API for CRUD operations on pipelines
- Real-time log streaming via Server-Sent Events (SSE)
- Artifact management (upload/download)
- Project and webhook management UI
- Responsive web dashboard with Tailwind CSS
- Keyboard shortcuts for power users

**Running Locally**:
```bash
# Build
make build-api-server

# Run (requires kubeconfig)
./bin/api-server \
  --port=8080 \
  --kubeconfig=$HOME/.kube/config \
  --s3-endpoint=http://localhost:9000 \
  --s3-bucket=c8s-logs \
  --log-level=debug
```

**API Endpoints**:

**Pipelines**:
- `GET /api/v1/pipelines` - List all pipelines
- `GET /api/v1/pipelines/:name` - Get pipeline details
- `POST /api/v1/pipelines` - Create pipeline
- `DELETE /api/v1/pipelines/:name` - Delete pipeline

**Pipeline Runs**:
- `GET /api/v1/runs` - List all runs
- `GET /api/v1/runs/:id` - Get run details
- `POST /api/v1/runs` - Trigger pipeline run
- `GET /api/v1/runs/:id/logs` - Stream logs (SSE)
- `GET /api/v1/runs/:id/logs?step=name` - Stream step logs

**Artifacts**:
- `GET /api/v1/runs/:id/artifacts` - List artifacts
- `GET /api/v1/runs/:id/artifacts/:name` - Download artifact

**Projects** (Dashboard):
- `GET /api/v1/projects` - List projects
- `POST /api/v1/projects` - Create project
- `PUT /api/v1/projects/:id` - Update project
- `DELETE /api/v1/projects/:id` - Delete project

**Web Dashboard**:
- `/` - Dashboard home
- `/pipelines` - Pipeline list
- `/pipelines/:name` - Pipeline details
- `/runs` - Run history
- `/runs/:id` - Run details with logs
- `/projects` - Project management
- `/webhooks` - Webhook management

**Environment Variables**:
- `PORT` - HTTP server port (default: 8080)
- `KUBECONFIG` - Path to kubeconfig
- `S3_ENDPOINT` - S3-compatible storage endpoint
- `S3_BUCKET` - Bucket for logs/artifacts
- `S3_ACCESS_KEY` - S3 access key
- `S3_SECRET_KEY` - S3 secret key
- `JWT_SECRET` - Secret for JWT token signing
- `LOG_LEVEL` - Logging level

**Source Files**:
```
api-server/
├── main.go                     # Entry point and HTTP server
├── handlers/                   # HTTP request handlers
│   ├── pipelines.go
│   ├── runs.go
│   ├── logs.go
│   ├── artifacts.go
│   ├── projects.go
│   └── webhooks.go
├── templates/                  # HTMX templates
│   ├── base.html
│   ├── dashboard.html
│   ├── pipelines.html
│   └── runs.html
└── static/                     # CSS, JavaScript
    ├── css/
    └── js/
```

**Related**:
- [pkg/api/](../pkg/api/) - API handler implementations
- [docs/guides/dashboard-features.md](../docs/guides/dashboard-features.md) - Dashboard guide
- [specs/001-build-a-continuous/contracts/openapi.yaml](../specs/001-build-a-continuous/contracts/openapi.yaml) - API specification

---

## webhook/

**Purpose**: Receives Git webhooks and creates PipelineRun resources.

**Binary**: `bin/webhook`

**What It Does**:
- Receives webhooks from GitHub, GitLab, Bitbucket
- Validates webhook signatures
- Parses commit, branch, and PR information
- Matches repository to PipelineConfig
- Creates PipelineRun CR to trigger pipeline

**Key Features**:
- Multi-provider support (GitHub, GitLab, Bitbucket)
- Signature verification for security
- Branch filtering (only trigger on configured branches)
- PR event support
- Webhook secret validation

**Running Locally**:
```bash
# Build
make build-webhook

# Run (requires kubeconfig)
./bin/webhook \
  --port=9090 \
  --kubeconfig=$HOME/.kube/config \
  --namespace=c8s-system \
  --log-level=debug
```

**Webhook Endpoints**:
- `POST /webhook/github` - GitHub webhooks
- `POST /webhook/gitlab` - GitLab webhooks
- `POST /webhook/bitbucket` - Bitbucket webhooks
- `GET /health` - Health check

**Webhook Payload Processing**:
1. Receive webhook POST
2. Verify signature using webhook secret
3. Parse event type (push, pull_request, etc.)
4. Extract commit SHA, branch, repository URL
5. Find matching PipelineConfig by repository
6. Check if branch matches config
7. Create PipelineRun CR with commit/branch info
8. Return 200 OK or error

**Environment Variables**:
- `PORT` - HTTP server port (default: 9090)
- `KUBECONFIG` - Path to kubeconfig
- `NAMESPACE` - Namespace to create PipelineRuns
- `WEBHOOK_SECRET` - Secret for validating webhooks
- `LOG_LEVEL` - Logging level

**GitHub Configuration**:
1. Go to repository Settings → Webhooks
2. Add webhook:
   - **Payload URL**: `https://your-domain.com/webhook/github`
   - **Content type**: `application/json`
   - **Secret**: (your webhook secret)
   - **Events**: Push, Pull request
3. Save webhook

**Local Development with ngrok**:
```bash
# Start webhook service
./bin/webhook --port=9090

# In another terminal, start ngrok
ngrok http 9090

# Use ngrok URL in GitHub webhook configuration
# https://abc123.ngrok.io/webhook/github
```

**Source Files**:
```
webhook/
└── main.go                # Entry point and webhook handlers
```

**Related**:
- [pkg/webhook/](../pkg/webhook/) - Webhook handler implementations
- [docs/development/c8s-dogfooding.md](../docs/development/c8s-dogfooding.md) - Webhook setup guide

---

## c8s/

**Purpose**: Command-line interface for C8S pipeline management.

**Binary**: `bin/c8s`

**What It Does**:
- Trigger pipeline runs manually
- View pipeline status and history
- Stream logs from running pipelines
- Download artifacts
- Manage projects and webhooks

**Key Features**:
- Interactive TUI for pipeline selection
- Live log streaming with colors
- Tab completion for Bash/Zsh
- JSON output mode for scripting
- Multiple output formats (table, JSON, YAML)

**Installation**:
```bash
# Build from source
make build-cli

# Install to PATH
sudo cp bin/c8s /usr/local/bin/

# Or use go install
go install github.com/lavigneer/c8s/cmd/c8s@latest
```

**Usage**:

**Run a pipeline**:
```bash
# Trigger pipeline run
c8s run my-pipeline \
  --commit=$(git rev-parse HEAD) \
  --branch=$(git branch --show-current)

# With custom parameters
c8s run my-pipeline \
  --commit=abc123 \
  --branch=main \
  --param="ENV=production"
```

**View pipelines**:
```bash
# List all pipelines
c8s pipelines list

# Get pipeline details
c8s pipelines get my-pipeline

# Watch pipeline runs
c8s runs list --watch
```

**Stream logs**:
```bash
# Follow logs from latest run
c8s logs my-pipeline-xxxxx --follow

# View specific step logs
c8s logs my-pipeline-xxxxx --step=build

# Save logs to file
c8s logs my-pipeline-xxxxx > build.log
```

**Manage artifacts**:
```bash
# List artifacts
c8s artifacts list my-pipeline-xxxxx

# Download artifact
c8s artifacts get my-pipeline-xxxxx app.tar.gz

# Download all artifacts
c8s artifacts download my-pipeline-xxxxx --all
```

**Configuration**:
```bash
# Set default API endpoint
c8s config set endpoint https://c8s.example.com

# Set default output format
c8s config set output json

# View current config
c8s config list
```

**Environment Variables**:
- `C8S_ENDPOINT` - API server endpoint (default: http://localhost:8080)
- `C8S_TOKEN` - API authentication token
- `C8S_OUTPUT` - Output format (table, json, yaml)
- `C8S_NAMESPACE` - Kubernetes namespace (default: c8s-system)

**Shell Completion**:
```bash
# Bash
c8s completion bash > /etc/bash_completion.d/c8s

# Zsh
c8s completion zsh > "${fpath[1]}/_c8s"

# Fish
c8s completion fish > ~/.config/fish/completions/c8s.fish
```

**Source Files**:
```
c8s/
├── main.go                # Entry point and CLI setup
└── commands/              # Command implementations (if using cobra)
    ├── run.go
    ├── logs.go
    ├── pipelines.go
    └── artifacts.go
```

**Related**:
- [pkg/cli/](../pkg/cli/) - CLI command implementations
- [specs/002-i-want-to/](../specs/002-i-want-to/) - CLI specification

---

## Building

### Build All Components

```bash
# Build everything
make build

# Binaries output to:
# - bin/controller
# - bin/api-server
# - bin/webhook
# - bin/c8s
```

### Build Individual Components

```bash
# Controller
make build-controller

# API Server
make build-api-server

# Webhook
make build-webhook

# CLI
make build-cli
```

### Cross-Compilation

```bash
# Build for Linux (from Mac)
GOOS=linux GOARCH=amd64 go build -o bin/controller-linux ./cmd/controller

# Build for ARM64
GOOS=linux GOARCH=arm64 go build -o bin/controller-arm64 ./cmd/controller
```

---

## Docker Images

### Build Images

```bash
# Build all images
make docker-build

# Build specific image
docker build -f Dockerfile -t c8s-controller:latest .
docker build -f Dockerfile -t c8s-api-server:latest .
docker build -f Dockerfile -t c8s-webhook:latest .
```

### Push to Registry

```bash
# Push all images
make docker-push

# Push specific image
docker tag c8s-controller:latest ghcr.io/org/c8s-controller:v1.0.0
docker push ghcr.io/org/c8s-controller:v1.0.0
```

---

## Development

### Running Components Locally

**Controller**:
```bash
go run ./cmd/controller \
  --kubeconfig=$HOME/.kube/config \
  --log-level=debug
```

**API Server**:
```bash
go run ./cmd/api-server \
  --port=8080 \
  --kubeconfig=$HOME/.kube/config \
  --log-level=debug
```

**Webhook**:
```bash
go run ./cmd/webhook \
  --port=9090 \
  --kubeconfig=$HOME/.kube/config \
  --log-level=debug
```

**CLI**:
```bash
go run ./cmd/c8s pipelines list
```

### Hot Reload with Tilt

For the best development experience, use Tilt:
```bash
# Start Tilt (builds and deploys all components)
tilt up

# Edit any Go file, Tilt auto-rebuilds and redeploys
vim cmd/controller/main.go
```

See [docs/development/TILT-WORKFLOW.md](../docs/development/TILT-WORKFLOW.md) for details.

---

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Test specific component
go test ./cmd/controller/...
go test ./cmd/api-server/...
```

### Integration Tests

```bash
# Run integration tests (requires cluster)
make test-integration
```

### E2E Tests

```bash
# Run Playwright E2E tests
make test-e2e
```

See [tests/README.md](../tests/README.md) for comprehensive testing guide.

---

## Troubleshooting

### Controller Not Starting
```bash
# Check logs
kubectl logs -f deployment/c8s-controller -n c8s-system

# Common issues:
# - CRDs not installed: make install-crds
# - RBAC issues: check config/rbac/
# - S3 connection: verify S3_ENDPOINT and credentials
```

### API Server Connection Refused
```bash
# Check if server is running
kubectl get pods -n c8s-system | grep api-server

# Check service
kubectl get svc c8s-api-server -n c8s-system

# Port forward for local testing
kubectl port-forward svc/c8s-api-server 8080:8080 -n c8s-system
```

### Webhook Not Receiving Events
```bash
# Check webhook logs
kubectl logs -f deployment/c8s-webhook -n c8s-system

# Verify webhook configuration in GitHub
# Check webhook secret matches
# Verify URL is publicly accessible (use ngrok for local testing)
```

---

## Related

- [Makefile](../Makefile) - Build automation
- [pkg/](../pkg/README.md) - Shared package implementations
- [config/](../config/README.md) - Kubernetes manifests
- [Dockerfile](../Dockerfile) - Multi-stage Docker build
- [Tiltfile](../Tiltfile) - Local development workflow
