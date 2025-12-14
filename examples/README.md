# C8S Pipeline Examples

Example pipeline configurations demonstrating various C8S features and use cases.

## Quick Start

Copy any example to your repository as `.c8s.yaml`:

```bash
# Copy simple Go pipeline
cp examples/simple-go-pipeline.yaml .c8s.yaml

# Customize for your project
vim .c8s.yaml

# Commit and push
git add .c8s.yaml
git commit -m "Add C8S pipeline"
git push
```

## Examples Overview

### 1. Simple Go Pipeline (`simple-go-pipeline.yaml`)

**Use Case**: Basic Go application with tests and build

**Features**:
- Parallel execution (test and lint run concurrently)
- Dependency-based ordering (build waits for test)
- Artifact collection
- Resource limits

**Best For**: Simple Go projects, learning C8S basics

### 2. Docker Build Pipeline (`docker-build-pipeline.yaml`)

**Use Case**: Build and push Docker images to a registry

**Features**:
- Docker-in-Docker setup
- Secret management (registry credentials)
- Image tagging (commit SHA + latest)
- Conditional execution (only push on main branch)

**Best For**: Containerized applications, microservices

**Prerequisites**:
- Docker-in-Docker enabled in cluster
- Kubernetes Secret with Docker credentials:
  ```bash
  kubectl create secret generic docker-credentials \
    --from-literal=username=myuser \
    --from-literal=password=mypass \
    -n c8s-system
  ```

### 3. Matrix Pipeline (`matrix-pipeline.yaml`)

**Use Case**: Test across multiple versions and platforms

**Features**:
- Matrix builds (multiple Go versions × OS platforms)
- Matrix variable substitution
- Exclude specific combinations
- Parallel execution of matrix variants

**Best For**: Libraries, multi-platform tools

**Note**: Generates N × M pipelines (e.g., 2 Go versions × 2 OS = 4 parallel runs)

### 4. Monorepo Pipeline (`monorepo-pipeline.yaml`)

**Use Case**: Build multiple services in a monorepo

**Features**:
- Shared dependency installation
- Parallel service builds
- Integration tests after all services tested
- Conditional deployment
- Extended timeout

**Best For**: Monorepos, microservices, complex projects

## Common Patterns

### Parallel Execution

Run independent steps in parallel:

```yaml
steps:
  - name: test
    image: node:20
    commands: [npm test]

  - name: lint
    image: node:20
    commands: [npm run lint]

  # Both test and lint run in parallel (no dependencies)
```

### Sequential Execution

Chain steps with dependencies:

```yaml
steps:
  - name: test
    image: node:20
    commands: [npm test]

  - name: build
    image: node:20
    commands: [npm run build]
    dependsOn: [test]  # Waits for test to complete

  - name: deploy
    image: ubuntu:22.04
    commands: [./deploy.sh]
    dependsOn: [build]  # Waits for build to complete
```

### Conditional Execution

Run steps conditionally:

```yaml
steps:
  - name: deploy-prod
    image: ubuntu:22.04
    commands: [./deploy.sh production]
    conditional:
      branch: "main"        # Only on main branch
      onSuccess: true       # Only if previous steps succeeded

  - name: deploy-staging
    image: ubuntu:22.04
    commands: [./deploy.sh staging]
    conditional:
      branch: "develop"     # Only on develop branch
```

### Using Secrets

Inject secrets from Kubernetes:

```yaml
steps:
  - name: deploy
    image: ubuntu:22.04
    commands:
      - ./deploy.sh --token=$API_TOKEN
    secrets:
      - secretRef: deploy-credentials  # Kubernetes Secret name
        key: API_TOKEN                 # Key in the Secret
        envVar: API_TOKEN              # Environment variable name
```

Create the secret:
```bash
kubectl create secret generic deploy-credentials \
  --from-literal=API_TOKEN=abc123 \
  -n c8s-system
```

### Collecting Artifacts

Save build artifacts:

```yaml
steps:
  - name: build
    image: golang:1.25
    commands:
      - go build -o myapp
    artifacts:
      - myapp              # Uploaded to S3
      - dist/              # Can specify directories
      - "*.tar.gz"         # Glob patterns supported
```

Artifacts are uploaded to S3 and available via API:
```bash
# Download artifact
curl http://c8s-api/api/v1/runs/{run-id}/artifacts/myapp > myapp
```

### Resource Limits

Set CPU and memory limits:

```yaml
steps:
  - name: build
    image: golang:1.25
    commands: [go build]
    resources:
      cpu: 2000m       # 2 CPU cores
      memory: 4Gi      # 4 GB memory
```

## Environment Variables

C8S automatically provides these environment variables:

| Variable | Description | Example |
|----------|-------------|---------|
| `COMMIT_SHA` | Git commit SHA | `abc123def456` |
| `BRANCH` | Git branch name | `main` |
| `REPO_URL` | Repository URL | `https://github.com/org/repo` |
| `PIPELINE_NAME` | Pipeline name | `my-pipeline` |
| `RUN_ID` | Unique run ID | `my-pipeline-1234567890` |
| `STEP_NAME` | Current step name | `build` |

Usage:
```yaml
steps:
  - name: build
    image: golang:1.25
    commands:
      - echo "Building commit $COMMIT_SHA on branch $BRANCH"
      - go build -ldflags "-X main.version=$COMMIT_SHA"
```

## Best Practices

### 1. Use Specific Image Tags

❌ **Bad**:
```yaml
image: node:latest  # Can break builds when new version released
```

✅ **Good**:
```yaml
image: node:20      # Specific major version
image: node:20.10   # Even more specific
```

### 2. Optimize Dependencies

Share dependency installation:

```yaml
steps:
  # Install once
  - name: install
    image: node:20
    commands: [npm ci]

  # Reuse in subsequent steps
  - name: test
    image: node:20
    commands: [npm test]
    dependsOn: [install]

  - name: build
    image: node:20
    commands: [npm run build]
    dependsOn: [install]
```

### 3. Set Appropriate Timeouts

```yaml
timeout: 30m  # Default: 30 minutes

steps:
  - name: quick-test
    timeout: 5m   # Override per step
    # ...
```

### 4. Use Retry Policies

```yaml
retryPolicy:
  maxRetries: 3
  backoffSeconds: 30

steps:
  # Flaky tests will retry automatically
  - name: integration-tests
    # ...
```

### 5. Keep Steps Small

Break large steps into smaller ones for:
- Better parallelization
- Easier debugging
- Faster failure feedback

## Validation

Validate your pipeline before committing:

```bash
# Install C8S CLI
# See: docs/guides/getting-started.md

# Validate pipeline YAML
c8s validate .c8s.yaml

# Dry run (doesn't execute, just validates)
c8s run --dry-run
```

## More Examples

See the full pipeline configuration reference:
- [Pipeline Syntax Guide](../docs/guides/pipeline-syntax.md)
- [Configuration Schema](../specs/001-build-a-continuous/contracts/pipeline-config-schema.json)

## Contributing Examples

Have a useful pipeline example? Share it!

1. Create a new example file in `examples/`
2. Add description to this README
3. Submit a PR

Example naming: `{language}-{feature}-pipeline.yaml` (e.g., `python-ml-pipeline.yaml`)

## Questions?

- **Documentation**: [Pipeline Syntax](../docs/guides/pipeline-syntax.md)
- **Troubleshooting**: [Troubleshooting Guide](../docs/guides/troubleshooting.md)
- **Discussions**: [GitHub Discussions](https://github.com/lavigneer/c8s/discussions)
