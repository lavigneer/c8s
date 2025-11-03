# C8S Pipeline Syntax Reference

**Version**: 1.0
**Audience**: Pipeline developers, DevOps engineers
**Last Updated**: 2025-11-02

Complete reference for the C8S Pipeline YAML syntax and configuration options.

## Table of Contents

- [Overview](#overview)
- [Pipeline Structure](#pipeline-structure)
- [PipelineConfig Reference](#pipelineconfig-reference)
- [Pipeline Steps](#pipeline-steps)
- [Conditional Execution](#conditional-execution)
- [Matrix Strategy](#matrix-strategy)
- [Resource Configuration](#resource-configuration)
- [Secrets and Environment Variables](#secrets-and-environment-variables)
- [Timeouts and Retries](#timeouts-and-retries)
- [Examples](#examples)
- [Validation](#validation)

---

## Overview

C8S pipelines are defined using Kubernetes Custom Resources (CRDs). Each pipeline is a `PipelineConfig` resource that describes:

- **What to build**: Git repository and branches
- **How to build**: Steps with container images and commands
- **When to build**: Triggers and conditional execution
- **How to scale**: Matrix strategies for multi-configuration builds
- **Error handling**: Retries and fallback strategies

### Key Concepts

- **PipelineConfig**: The pipeline definition (persistent configuration)
- **PipelineRun**: A single execution of a pipeline (created automatically on trigger)
- **Step**: An individual task in the pipeline (runs in a container)
- **Matrix**: Multi-dimensional build strategy for testing across configurations
- **Artifacts**: Files captured after step execution for download/inspection

---

## Pipeline Structure

### Minimal Pipeline

```yaml
apiVersion: c8s.dev/v1alpha1
kind: PipelineConfig
metadata:
  name: my-pipeline
  namespace: default
spec:
  repository: https://github.com/org/repo.git
  branches: ["main"]
  steps:
    - name: test
      image: golang:1.25
      commands:
        - go test ./...
```

### Complete Pipeline

```yaml
apiVersion: c8s.dev/v1alpha1
kind: PipelineConfig
metadata:
  name: complete-pipeline
  namespace: default
spec:
  repository: https://github.com/org/repo.git
  branches: ["main", "develop", "release/*"]
  timeout: "2h"

  steps:
    - name: test
      image: golang:1.25
      commands:
        - go test ./...
      resources:
        cpu: "500m"
        memory: "512Mi"
      timeout: "30m"

    - name: build
      image: golang:1.25
      commands:
        - go build -o bin/app .
      dependsOn: ["test"]
      resources:
        cpu: "1000m"
        memory: "1Gi"
      artifacts:
        - "bin/*"

    - name: deploy
      image: kubernetes:latest
      commands:
        - kubectl apply -f deploy/
      dependsOn: ["build"]
      secrets:
        - secretRef: deploy-credentials
          key: kubeconfig
          envVar: KUBECONFIG
      conditional:
        branch: "main"
        onSuccess: true

  matrix:
    dimensions:
      go_version: ["1.24", "1.25"]
      os: ["ubuntu", "alpine"]

  retryPolicy:
    maxRetries: 2
    backoffSeconds: 30
```

---

## PipelineConfig Reference

### Spec Fields

#### `repository` (Required)

Git repository URL for the pipeline source code.

**Format**: HTTP(S) or SSH URL
**Validation**: Must match `^(https?|git|ssh)://.*`

```yaml
spec:
  repository: https://github.com/org/repo.git
  # or
  repository: git@github.com:org/repo.git
  # or
  repository: ssh://git@github.com/org/repo.git
```

**Notes**:
- HTTP(S) URLs are recommended for public repositories
- SSH URLs require SSH key setup in the cluster
- Must be accessible from the C8S controller pods

#### `branches` (Optional)

Branch filter patterns for automatic triggering.

**Format**: Array of glob patterns
**Default**: `["*"]` (all branches)
**Validation**: Standard glob patterns

```yaml
spec:
  branches:
    - "main"                    # Exact match
    - "develop"                 # Exact match
    - "release/*"              # Glob pattern
    - "feature/*/backend"      # Multi-level glob
    - "*"                       # All branches (default)
```

**Matching Rules**:
- `*` matches any single path component
- `**` matches any characters including `/`
- `?` matches any single character
- `[abc]` matches any character in brackets

**Examples**:
```yaml
branches:
  - "main"                              # Only main branch
  - "release/*"                         # release/1.0, release/2.0, etc.
  - "feature/*"                         # feature/auth, feature/api, etc.
  - "hotfix/*-urgent"                   # hotfix/bug-123-urgent, etc.
  - "refs/tags/v*"                      # Tags matching v*
```

#### `steps` (Required)

Array of pipeline steps to execute sequentially or in parallel (based on `dependsOn`).

**Validation**: Minimum 1 step required
**Format**: Array of `PipelineStep` objects

See [Pipeline Steps](#pipeline-steps) section for details.

#### `timeout` (Optional)

Pipeline-level timeout. If any step exceeds this duration, the entire pipeline fails.

**Format**: Duration string with unit (s, m, h)
**Default**: `"1h"` (1 hour)
**Validation**: Pattern `^[0-9]+(s|m|h)$`

```yaml
spec:
  timeout: "30m"      # 30 minutes
  timeout: "2h"       # 2 hours
  timeout: "3600s"    # 3600 seconds
```

**Behavior**:
- If any step runs longer than this timeout, the step is killed
- Timeout starts when the first step begins execution
- Individual step timeouts override this (if shorter)

#### `matrix` (Optional)

Matrix strategy for parallel execution across multiple configurations.

**Format**: `MatrixStrategy` object

See [Matrix Strategy](#matrix-strategy) section for details.

#### `retryPolicy` (Optional)

Retry behavior for failed steps.

**Format**: `RetryPolicy` object

See [Timeouts and Retries](#timeouts-and-retries) section for details.

---

## Pipeline Steps

### Step Fields

#### `name` (Required)

Unique identifier for the step within the pipeline.

**Validation**: Pattern `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
**Requirements**:
- Lowercase letters, numbers, hyphens
- Start/end with letter or number
- Unique within the pipeline

```yaml
steps:
  - name: test              # Valid
  - name: build-binary      # Valid
  - name: deploy-to-prod    # Valid
  - name: Validate          # Invalid (uppercase)
  - name: -test             # Invalid (starts with hyphen)
  - name: test_unit         # Invalid (underscore not allowed)
```

#### `image` (Required)

Container image for step execution.

**Format**: Docker image reference
**Requirements**: Image must be accessible from cluster nodes

```yaml
steps:
  - name: build
    image: golang:1.25                      # Docker Hub
    image: gcr.io/my-project/builder:v1     # Google Container Registry
    image: quay.io/my-org/tool:latest       # Quay.io
    image: private-registry:5000/app:v1.0   # Private registry
```

**Image Pull Secrets**:
For private registries, configure `imagePullSecrets` in controller configuration:

```yaml
# In controller config
imagePullSecrets:
  - name: dockercfg-secret
  - name: gcr-secret
```

#### `commands` (Required)

Shell commands to execute within the container.

**Format**: Array of shell command strings
**Validation**: Minimum 1 command required
**Shell**: Commands are executed in `/bin/sh` by default

```yaml
steps:
  - name: test
    image: golang:1.25
    commands:
      - go mod download
      - go test ./...
      - go test ./... -race
```

**Command Behavior**:
- Commands run sequentially (stop on first failure)
- Working directory is the cloned repository
- Environment variables are accessible
- Standard output/error is captured as logs

**Multi-line Commands**:
```yaml
commands:
  - |
    if [ -f Makefile ]; then
      make test
    else
      go test ./...
    fi
```

#### `dependsOn` (Optional)

Step names that must complete before this step executes.

**Format**: Array of step names
**Default**: None (execute in dependency order)

```yaml
steps:
  - name: test
    image: golang:1.25
    commands:
      - go test ./...

  - name: build
    image: golang:1.25
    commands:
      - go build -o bin/app .
    dependsOn: ["test"]        # Build only after test succeeds

  - name: package
    image: alpine:latest
    commands:
      - tar -czf app.tar.gz bin/
    dependsOn: ["build"]       # Package only after build succeeds

  - name: cleanup
    image: alpine:latest
    commands:
      - rm -rf bin/
    dependsOn: ["test", "build", "package"]  # After all complete
```

**Execution Order**:
- Steps without `dependsOn` execute first (in order)
- Steps with `dependsOn` wait for dependencies
- Parallel execution possible with independent steps

#### `timeout` (Optional)

Step-level timeout.

**Format**: Duration string (s, m, h)
**Default**: `"30m"` (30 minutes)
**Validation**: Pattern `^[0-9]+(s|m|h)$`

```yaml
steps:
  - name: quick-test
    timeout: "5m"           # 5 minute timeout

  - name: long-build
    timeout: "1h30m"        # Error: must use single unit
    timeout: "90m"          # Correct: 90 minutes
```

#### `artifacts` (Optional)

File patterns to upload to artifact storage after step completes.

**Format**: Array of glob patterns
**Examples**: `["bin/*", "*.log", "build/output/**"]`

```yaml
steps:
  - name: build
    image: golang:1.25
    commands:
      - go build -o bin/app .
      - go build -o bin/app-windows.exe .
    artifacts:
      - "bin/*"              # All files in bin directory
      - "build/**"           # Recursive glob
      - "*.tar.gz"           # Tar archives in root
      - "logs/*.log"         # Specific directory pattern
```

**Artifact Behavior**:
- Patterns evaluated from repository root
- Artifacts captured after successful step execution
- If step fails, no artifacts are captured
- Artifacts stored in S3-compatible storage
- Artifacts available for download in dashboard

**Common Patterns**:
```yaml
artifacts:
  - "bin/*"                  # Build artifacts
  - "dist/**"                # Distribution files
  - "coverage.html"          # Coverage reports
  - "build/logs/**"          # Build logs
  - "*.jar"                  # JAR files
  - "target/site/**"         # Maven site
```

#### `secrets` (Optional)

Kubernetes Secrets to inject as environment variables.

**Format**: Array of `SecretReference` objects

See [Secrets and Environment Variables](#secrets-and-environment-variables) section.

#### `resources` (Optional)

CPU and memory resource requests and limits.

**Format**: `ResourceRequirements` object

See [Resource Configuration](#resource-configuration) section.

#### `conditional` (Optional)

Conditions for step execution.

**Format**: `ConditionalExecution` object

See [Conditional Execution](#conditional-execution) section.

---

## Conditional Execution

Conditionally execute steps based on branch or previous step results.

### Conditional Fields

#### `branch` (Optional)

Execute step only on matching branch.

**Format**: Glob pattern
**Default**: None (execute on all branches)

```yaml
steps:
  - name: deploy-prod
    image: kubernetes:latest
    commands:
      - kubectl apply -f deploy/production.yaml
    conditional:
      branch: "main"         # Execute only on main branch
```

**Pattern Matching**:
```yaml
conditional:
  branch: "main"             # Exact match
  branch: "release/*"        # Glob pattern
  branch: "hotfix-*"         # Wildcard pattern
```

#### `onSuccess` (Optional)

Execute step only if all previous steps succeeded.

**Format**: Boolean
**Default**: `true` (execute on success)

```yaml
steps:
  - name: test
    image: golang:1.25
    commands:
      - go test ./...

  - name: notify-success
    image: curlimages/curl:latest
    commands:
      - curl https://hooks.slack.com/... -d "Build succeeded"
    conditional:
      onSuccess: true        # Only if test passed

  - name: notify-failure
    image: curlimages/curl:latest
    commands:
      - curl https://hooks.slack.com/... -d "Build failed"
    conditional:
      onSuccess: false       # Only if test failed
```

### Complete Example

```yaml
steps:
  - name: test
    image: golang:1.25
    commands:
      - go test ./...

  - name: build
    image: golang:1.25
    commands:
      - go build -o bin/app .
    dependsOn: ["test"]

  - name: deploy-staging
    image: kubernetes:latest
    commands:
      - kubectl set image deployment/app app=myapp:latest -n staging
    dependsOn: ["build"]
    conditional:
      branch: "develop"      # Only deploy develop to staging

  - name: deploy-production
    image: kubernetes:latest
    commands:
      - kubectl set image deployment/app app=myapp:latest -n prod
    dependsOn: ["build"]
    conditional:
      branch: "main"         # Only deploy main to production

  - name: notify
    image: curlimages/curl:latest
    commands:
      - curl https://hooks.slack.com/... -d "Pipeline complete"
    conditional:
      onSuccess: true        # Notify only on success
```

---

## Matrix Strategy

Build across multiple configurations simultaneously.

### Matrix Fields

#### `dimensions` (Required)

Variables and their values for matrix expansion.

**Format**: Map of variable name to array of values
**Example**: `{"go_version": ["1.24", "1.25"], "os": ["ubuntu", "alpine"]}`

```yaml
matrix:
  dimensions:
    go_version: ["1.24", "1.25"]
    os: ["ubuntu", "alpine"]
```

**Generated Combinations**:
```
go_version=1.24, os=ubuntu
go_version=1.24, os=alpine
go_version=1.25, os=ubuntu
go_version=1.25, os=alpine
```

#### `exclude` (Optional)

Exclude specific dimension combinations.

**Format**: Array of maps
**Use Case**: Skip invalid or unsupported combinations

```yaml
matrix:
  dimensions:
    go_version: ["1.24", "1.25"]
    os: ["ubuntu", "alpine", "windows"]
  exclude:
    - go_version: "1.24"
      os: "windows"          # Go 1.24 doesn't support Windows yet
    - go_version: "1.25"
      os: "alpine"
      arch: "arm64"          # Not supported
```

### Environment Variables

Matrix variables are exposed as environment variables in each job:

```yaml
matrix:
  dimensions:
    go_version: ["1.24", "1.25"]
    os: ["ubuntu", "alpine"]

steps:
  - name: test
    image: golang:latest
    commands:
      - echo "Testing with Go $GO_VERSION on $OS"
      - echo "GO_VERSION=$GO_VERSION"
      - echo "OS=$OS"
```

**Variable Naming**:
- Dimension names converted to UPPERCASE
- Hyphens replaced with underscores
- Available in all steps of that matrix job

### Complete Example

```yaml
spec:
  steps:
    - name: test
      image: golang:latest
      commands:
        - go version
        - go test ./...

    - name: build
      image: golang:latest
      commands:
        - go build -o bin/app-$OS .
      artifacts:
        - "bin/*"

  matrix:
    dimensions:
      go_version: ["1.24", "1.25"]
      os: ["linux", "darwin"]
    exclude:
      - os: "darwin"         # Skip macOS for now
```

This creates 2 parallel jobs (1.24/linux and 1.25/linux).

---

## Resource Configuration

Define CPU and memory requirements for steps.

### ResourceRequirements Fields

#### `cpu` (Optional)

CPU request and limit.

**Format**: Kubernetes CPU notation (m = millicores)
**Default**: `"500m"`
**Validation**: Pattern `^[0-9]+m?$`

```yaml
resources:
  cpu: "100m"                # 100 millicores
  cpu: "1"                   # 1 core
  cpu: "2500m"               # 2.5 cores
```

**Guidelines**:
- **Light tasks** (linting, tests): `50m` - `200m`
- **Build tasks**: `500m` - `1000m`
- **Heavy tasks** (compilation): `1000m` - `2000m`

#### `memory` (Optional)

Memory request and limit.

**Format**: Kubernetes memory notation (Mi = mebibytes, Gi = gibibytes)
**Default**: `"1Gi"`
**Validation**: Pattern `^[0-9]+(Mi|Gi)$`

```yaml
resources:
  memory: "256Mi"             # 256 mebibytes
  memory: "1Gi"               # 1 gibibyte
  memory: "2Gi"               # 2 gibibytes
```

**Guidelines**:
- **Light tasks**: `128Mi` - `256Mi`
- **Build tasks**: `512Mi` - `1Gi`
- **Heavy tasks** (Java, large compilations): `1Gi` - `4Gi`

### Complete Example

```yaml
steps:
  - name: lint
    image: golangci/golangci-lint:latest
    commands:
      - golangci-lint run
    resources:
      cpu: "100m"
      memory: "256Mi"

  - name: test
    image: golang:1.25
    commands:
      - go test -v -race ./...
    resources:
      cpu: "500m"
      memory: "512Mi"

  - name: build
    image: golang:1.25
    commands:
      - go build -o bin/app .
    resources:
      cpu: "1000m"            # 1 core
      memory: "1Gi"

  - name: package
    image: alpine:latest
    commands:
      - tar -czf app.tar.gz bin/
    resources:
      cpu: "200m"
      memory: "256Mi"
```

---

## Secrets and Environment Variables

Inject Kubernetes Secrets as environment variables.

### SecretReference Fields

#### `secretRef` (Required)

Kubernetes Secret name.

**Format**: Secret resource name in same namespace
**Requirements**: Secret must exist and be accessible

```yaml
secrets:
  - secretRef: deploy-credentials    # References a Kubernetes Secret
    key: kubeconfig
    envVar: KUBECONFIG
```

#### `key` (Required)

Key within the Secret containing the value.

**Format**: Key name in Secret data
**Requirements**: Key must exist in the Secret

```bash
# Create example secret
kubectl create secret generic deploy-credentials \
  --from-file=kubeconfig=/path/to/kubeconfig \
  --from-literal=deploy-token=sk_live_xxx
```

#### `envVar` (Optional)

Environment variable name to inject the value into.

**Format**: Valid environment variable name
**Default**: Uses the key name if not specified

```yaml
secrets:
  - secretRef: my-secret
    key: username
    envVar: DB_USER            # Inject as DB_USER

  - secretRef: my-secret
    key: password
    # Defaults to 'password' if envVar not specified
```

### Complete Example

```yaml
steps:
  - name: deploy
    image: kubernetes:latest
    commands:
      - kubectl config use-context $KUBE_CONTEXT
      - kubectl apply -f deploy/ --kubeconfig=$KUBECONFIG
    secrets:
      - secretRef: deploy-credentials
        key: kubeconfig
        envVar: KUBECONFIG

      - secretRef: deploy-credentials
        key: kube-context
        envVar: KUBE_CONTEXT

      - secretRef: docker-registry
        key: auth
        envVar: DOCKER_AUTH

  - name: notify
    image: curlimages/curl:latest
    commands:
      - curl -H "Authorization: Bearer $SLACK_TOKEN" https://slack.com/api/...
    secrets:
      - secretRef: notifications
        key: slack-token
        envVar: SLACK_TOKEN
```

### Creating Secrets

```bash
# From file
kubectl create secret generic deploy-credentials \
  --from-file=kubeconfig=/home/user/.kube/config \
  -n default

# From literal values
kubectl create secret generic notifications \
  --from-literal=slack-token=xoxb_xxx \
  -n default

# From multiple sources
kubectl create secret generic db-creds \
  --from-literal=username=admin \
  --from-literal=password=secret123 \
  -n default

# Using YAML
kubectl apply -f - <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: default
type: Opaque
stringData:
  username: admin
  password: secret123
  api-key: sk_live_xxx
EOF
```

---

## Timeouts and Retries

### Timeout Hierarchy

Timeouts are evaluated in order (first matching timeout applies):

1. **Step timeout** (if specified)
2. **Pipeline timeout** (if specified)
3. **Default**: 30m per step, 1h per pipeline

```yaml
spec:
  timeout: "2h"              # Pipeline-level timeout

  steps:
    - name: quick-test
      timeout: "5m"          # This step: 5 minutes max
      image: golang:1.25
      commands:
        - go test ./...

    - name: slow-build
      timeout: "1h"          # This step: 1 hour max
      image: golang:1.25
      commands:
        - go build -o bin/app .

    - name: package
      # Uses pipeline timeout of 2h
      image: alpine:latest
      commands:
        - tar -czf app.tar.gz bin/
```

### RetryPolicy Fields

#### `maxRetries` (Optional)

Maximum number of retry attempts.

**Format**: Integer
**Range**: 0-5
**Default**: `0` (no retries)

```yaml
retryPolicy:
  maxRetries: 3              # Retry up to 3 times
```

#### `backoffSeconds` (Optional)

Delay between retry attempts.

**Format**: Integer (seconds)
**Minimum**: 0
**Default**: `60` (1 minute)

```yaml
retryPolicy:
  maxRetries: 2
  backoffSeconds: 30         # Wait 30 seconds between retries
```

### Retry Example

```yaml
spec:
  retryPolicy:
    maxRetries: 2
    backoffSeconds: 30

  steps:
    - name: download-dependencies
      image: alpine:latest
      commands:
        - apk add curl
        - curl https://api.github.com/repos/... | jq .
      # If this fails, will retry up to 2 times with 30s delay
```

**Retry Flow**:
```
Attempt 1: Command runs
  ↓ (if fails)
Wait 30 seconds
Attempt 2: Command runs
  ↓ (if fails)
Wait 30 seconds
Attempt 3: Command runs
  ↓ (if fails)
Pipeline fails (max retries exceeded)
```

---

## Examples

### Simple Build and Test

```yaml
apiVersion: c8s.dev/v1alpha1
kind: PipelineConfig
metadata:
  name: simple-go-app
spec:
  repository: https://github.com/myorg/myapp.git
  branches: ["main", "develop"]

  steps:
    - name: test
      image: golang:1.25
      commands:
        - go test -v ./...

    - name: build
      image: golang:1.25
      commands:
        - go build -o app .
      dependsOn: ["test"]
      artifacts:
        - "app"
```

### Multi-Language Matrix Build

```yaml
apiVersion: c8s.dev/v1alpha1
kind: PipelineConfig
metadata:
  name: cross-platform-build
spec:
  repository: https://github.com/myorg/sdk.git
  branches: ["main"]

  steps:
    - name: test
      image: golang:1.25
      commands:
        - go test ./...

    - name: build
      image: golang:1.25
      commands:
        - go build -o build/app-$GOOS-$GOARCH .
      dependsOn: ["test"]
      artifacts:
        - "build/*"

  matrix:
    dimensions:
      goos: ["linux", "darwin", "windows"]
      goarch: ["amd64", "arm64"]
    exclude:
      - goos: "windows"
        goarch: "arm64"  # Windows ARM64 not yet supported
```

### Deployment Pipeline

```yaml
apiVersion: c8s.dev/v1alpha1
kind: PipelineConfig
metadata:
  name: deploy-app
spec:
  repository: https://github.com/myorg/app.git
  branches: ["main", "develop"]
  timeout: "30m"

  steps:
    - name: test
      image: golang:1.25
      commands:
        - go test -v ./...
      timeout: "10m"

    - name: build
      image: golang:1.25
      commands:
        - go build -o bin/app .
      dependsOn: ["test"]
      artifacts:
        - "bin/*"

    - name: docker-build
      image: docker:latest
      commands:
        - docker build -t myapp:$CI_COMMIT_SHA .
        - docker push myapp:$CI_COMMIT_SHA
      dependsOn: ["build"]
      secrets:
        - secretRef: docker-creds
          key: auth
          envVar: DOCKER_AUTH

    - name: deploy-staging
      image: kubernetes:latest
      commands:
        - kubectl set image deployment/app app=myapp:$CI_COMMIT_SHA -n staging
      dependsOn: ["docker-build"]
      conditional:
        branch: "develop"
      secrets:
        - secretRef: kubeconfig
          key: staging
          envVar: KUBECONFIG

    - name: deploy-production
      image: kubernetes:latest
      commands:
        - kubectl set image deployment/app app=myapp:$CI_COMMIT_SHA -n production
      dependsOn: ["docker-build"]
      conditional:
        branch: "main"
      secrets:
        - secretRef: kubeconfig
          key: production
          envVar: KUBECONFIG

    - name: notify
      image: curlimages/curl:latest
      commands:
        - curl -X POST https://hooks.slack.com/... -d "{\"text\":\"Deployment complete\"}"
      conditional:
        onSuccess: true
```

### Complex Build with Conditional Steps

```yaml
apiVersion: c8s.dev/v1alpha1
kind: PipelineConfig
metadata:
  name: conditional-pipeline
spec:
  repository: https://github.com/myorg/complex-app.git
  branches: ["main", "develop", "release/*"]

  steps:
    - name: lint
      image: python:3.11
      commands:
        - pip install flake8
        - flake8 src/
      resources:
        cpu: "100m"
        memory: "256Mi"

    - name: test
      image: python:3.11
      commands:
        - pip install -r requirements-test.txt
        - pytest -v
      dependsOn: ["lint"]
      resources:
        cpu: "500m"
        memory: "1Gi"

    - name: build
      image: python:3.11
      commands:
        - pip install build
        - python -m build
      dependsOn: ["test"]
      artifacts:
        - "dist/*"

    - name: publish-staging
      image: python:3.11
      commands:
        - pip install twine
        - twine upload --repository testpypi dist/*
      dependsOn: ["build"]
      conditional:
        branch: "develop"
      secrets:
        - secretRef: pypi-creds
          key: testpypi-token
          envVar: TWINE_PASSWORD

    - name: publish-production
      image: python:3.11
      commands:
        - pip install twine
        - twine upload dist/*
      dependsOn: ["build"]
      conditional:
        branch: "main"
      secrets:
        - secretRef: pypi-creds
          key: pypi-token
          envVar: TWINE_PASSWORD

    - name: notify-success
      image: curlimages/curl:latest
      commands:
        - curl https://hooks.slack.com/... -d "Pipeline succeeded"
      conditional:
        onSuccess: true

    - name: notify-failure
      image: curlimages/curl:latest
      commands:
        - curl https://hooks.slack.com/... -d "Pipeline failed"
      conditional:
        onSuccess: false
```

---

## Validation

### YAML Validation

Validate your pipeline YAML before applying:

```bash
# Validate with kubectl
kubectl apply -f pipeline.yaml --dry-run=client

# Validate against schema
kubectl explain pipelineconfig.spec
kubectl explain pipelineconfig.spec.steps
```

### Common Validation Errors

| Error | Cause | Solution |
|-------|-------|----------|
| `repository is required` | Missing repository field | Add `spec.repository: https://...` |
| `invalid pattern` | Branch pattern uses invalid syntax | Use standard glob patterns: `*`, `?`, `[abc]` |
| `timeout must match` | Invalid timeout format | Use format `[0-9]+(s\|m\|h)` e.g., `30m` |
| `name must be unique` | Duplicate step names | Rename steps with unique names |
| `dependsOn references missing step` | Invalid step dependency | Verify step names exist and are spelled correctly |
| `maxRetries must be 0-5` | Retry count out of range | Use 0-5 for maxRetries |

### Validation Examples

```bash
# Good: Valid pipeline
apiVersion: c8s.dev/v1alpha1
kind: PipelineConfig
metadata:
  name: valid-pipeline
spec:
  repository: https://github.com/org/repo.git
  branches: ["main"]
  steps:
    - name: test
      image: golang:1.25
      commands:
        - go test ./...

# Bad: Missing repository
spec:
  branches: ["main"]     # Missing required repository
  steps: [...]

# Bad: Invalid step name
steps:
  - name: My-Test        # Uppercase not allowed
    image: golang:1.25

# Bad: Timeout format
timeout: "30 minutes"    # Invalid format
timeout: "30m"           # Correct format

# Bad: Circular dependency
steps:
  - name: a
    dependsOn: ["b"]
  - name: b
    dependsOn: ["a"]     # Circular: a → b → a
```

---

## Related Documentation

- [Getting Started](./GETTING_STARTED.md) - Quick start guide
- [Configuration](./CONFIGURATION.md) - Configuration reference
- [Operator Guide](./OPERATOR_GUIDE.md) - Deployment guide
- [Troubleshooting](./TROUBLESHOOTING.md) - Common issues and solutions
