# C8S Quick Start Guide

**Feature**: Kubernetes-Native Continuous Integration System
**Audience**: Developers onboarding to C8S
**Last Updated**: 2025-10-12

## Overview

C8S is a Kubernetes-native CI system that runs pipeline steps as isolated container Jobs. This guide will get you from zero to running your first pipeline in 15 minutes.

---

## Prerequisites

- Kubernetes cluster (1.24+) with kubectl access
- Namespace with appropriate RBAC permissions
- Git repository with code to test
- Basic familiarity with YAML and containers

---

## Step 1: Install C8S (5 minutes)

### Install CRDs

```bash
# Install Custom Resource Definitions
kubectl apply -f https://raw.githubusercontent.com/org/c8s/main/deploy/crds.yaml

# Verify CRDs are installed
kubectl get crd | grep c8s.dev
# Expected output:
#   pipelineconfigs.c8s.dev
#   pipelineruns.c8s.dev
#   repositoryconnections.c8s.dev
```

### Install Controller and API Server

```bash
# Install C8S controller, webhook service, and API server
kubectl apply -f https://raw.githubusercontent.com/org/c8s/main/deploy/install.yaml

# Wait for pods to be ready
kubectl wait --for=condition=ready pod -l app=c8s-controller -n c8s-system --timeout=60s
kubectl wait --for=condition=ready pod -l app=c8s-webhook -n c8s-system --timeout=60s
kubectl wait --for=condition=ready pod -l app=c8s-api -n c8s-system --timeout=60s

# Verify installation
kubectl get pods -n c8s-system
# Expected output:
#   NAME                              READY   STATUS
#   c8s-controller-xxxxx-yyyyy        1/1     Running
#   c8s-webhook-xxxxx-yyyyy           1/1     Running
#   c8s-api-xxxxx-yyyyy               1/1     Running
```

### Install CLI (optional but recommended)

```bash
# macOS
brew install c8s/tap/c8s-cli

# Linux
curl -Lo /usr/local/bin/c8s https://github.com/org/c8s/releases/latest/download/c8s-linux-amd64
chmod +x /usr/local/bin/c8s

# Verify CLI installation
c8s version
```

---

## Step 2: Configure Object Storage (3 minutes)

C8S needs S3-compatible storage for logs and artifacts.

### Option A: Use Existing S3 Bucket

```bash
# Create secret with AWS credentials
kubectl create secret generic c8s-storage \
  --from-literal=access-key-id=YOUR_ACCESS_KEY \
  --from-literal=secret-access-key=YOUR_SECRET_KEY \
  --from-literal=bucket=your-bucket-name \
  --from-literal=region=us-west-2 \
  -n c8s-system
```

### Option B: Deploy MinIO for Local Testing

```bash
# Install MinIO for local development
kubectl apply -f https://raw.githubusercontent.com/org/c8s/main/deploy/minio.yaml

# Wait for MinIO to be ready
kubectl wait --for=condition=ready pod -l app=minio -n c8s-system --timeout=60s

# Get MinIO credentials (default: minioadmin / minioadmin)
kubectl get secret minio-credentials -n c8s-system -o yaml
```

---

## Step 3: Create Your First Pipeline (5 minutes)

### Create Pipeline Configuration File

In your repository root, create `.c8s.yaml`:

```yaml
version: v1alpha1
name: hello-world-ci
steps:
  - name: hello
    image: ubuntu:22.04
    commands:
      - echo "Hello from C8S!"
      - echo "Running in commit: $COMMIT_SHA"
      - echo "Branch: $BRANCH"
    resources:
      cpu: 500m
      memory: 512Mi
    timeout: 5m
```

### Create PipelineConfig in Kubernetes

```bash
# Create PipelineConfig from your .c8s.yaml
cat <<EOF | kubectl apply -f -
apiVersion: c8s.dev/v1alpha1
kind: PipelineConfig
metadata:
  name: hello-world-ci
  namespace: default
spec:
  repository: https://github.com/your-org/your-repo
  branches: ["*"]
  steps:
    - name: hello
      image: ubuntu:22.04
      commands:
        - echo "Hello from C8S!"
        - echo "Running in commit: \$COMMIT_SHA"
        - echo "Branch: \$BRANCH"
      resources:
        cpu: 500m
        memory: 512Mi
      timeout: 5m
  timeout: 10m
EOF

# Verify PipelineConfig was created
kubectl get pipelineconfig hello-world-ci
```

---

## Step 4: Trigger Your First Run (2 minutes)

### Manual Trigger via CLI

```bash
# Trigger pipeline run manually
c8s run hello-world-ci --commit=$(git rev-parse HEAD) --branch=$(git branch --show-current)

# Or use kubectl directly
cat <<EOF | kubectl apply -f -
apiVersion: c8s.dev/v1alpha1
kind: PipelineRun
metadata:
  generateName: hello-world-ci-
  namespace: default
spec:
  pipelineConfigRef: hello-world-ci
  commit: $(git rev-parse HEAD)
  branch: $(git branch --show-current)
  triggeredBy: $(whoami)
  triggeredAt: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
EOF
```

### Watch Pipeline Execution

```bash
# List pipeline runs
kubectl get pipelinerun

# Watch specific run (replace NAME with your run name)
c8s logs hello-world-ci-xxxxx --follow

# Or use kubectl
kubectl get pipelinerun hello-world-ci-xxxxx -w

# View step logs
kubectl logs -f job/hello-world-ci-xxxxx-hello
```

### Verify Success

```bash
# Check final status
kubectl get pipelinerun hello-world-ci-xxxxx -o jsonpath='{.status.phase}'
# Expected output: Succeeded

# View complete status
kubectl get pipelinerun hello-world-ci-xxxxx -o yaml
```

---

## Step 5: Set Up Webhook (Optional, 5 minutes)

Connect your repository to trigger pipelines automatically on push.

### Create Webhook Secret

```bash
# Generate webhook secret
WEBHOOK_SECRET=$(openssl rand -hex 32)

# Store in Kubernetes Secret
kubectl create secret generic my-repo-webhook \
  --from-literal=webhook-secret=$WEBHOOK_SECRET \
  -n default

# Save this secret - you'll need it for GitHub/GitLab configuration
echo "Webhook Secret: $WEBHOOK_SECRET"
```

### Create RepositoryConnection

```bash
cat <<EOF | kubectl apply -f -
apiVersion: c8s.dev/v1alpha1
kind: RepositoryConnection
metadata:
  name: my-repo-connection
  namespace: default
spec:
  repository: https://github.com/your-org/your-repo
  provider: github
  webhookSecretRef: my-repo-webhook
  pipelineConfigRef: hello-world-ci
EOF
```

### Configure Webhook in GitHub

1. Go to your repository settings → Webhooks → Add webhook
2. **Payload URL**: `https://c8s.example.com/webhooks/github`
3. **Content type**: `application/json`
4. **Secret**: Paste the `$WEBHOOK_SECRET` from above
5. **Events**: Select "Just the push event"
6. Click "Add webhook"

### Test Webhook

```bash
# Push a commit to trigger pipeline
git commit --allow-empty -m "Test C8S webhook"
git push

# Watch for new pipeline run
kubectl get pipelinerun -w
```

---

## Common Pipeline Patterns

### Multi-Step Pipeline with Dependencies

```yaml
version: v1alpha1
name: test-and-build
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
    dependsOn: [test]  # Only runs if test succeeds
    artifacts:
      - app
```

### Parallel Execution

```yaml
version: v1alpha1
name: parallel-tests
steps:
  - name: unit-tests
    image: golang:1.21
    commands:
      - go test ./internal/...

  - name: integration-tests
    image: golang:1.21
    commands:
      - go test ./integration/...

  - name: lint
    image: golangci/golangci-lint:latest
    commands:
      - golangci-lint run

  # All three steps above run in parallel

  - name: report
    image: ubuntu:22.04
    commands:
      - echo "All tests passed!"
    dependsOn: [unit-tests, integration-tests, lint]
```

### Using Secrets

```yaml
version: v1alpha1
name: deploy-with-secrets
steps:
  - name: test-database
    image: postgres:15
    commands:
      - psql -h $DB_HOST -U $DB_USER -c "SELECT version();"
    secrets:
      - secretRef: database-credentials
        key: DB_HOST
        envVar: DB_HOST
      - secretRef: database-credentials
        key: DB_USER
        envVar: DB_USER
      - secretRef: database-credentials
        key: DB_PASSWORD
        envVar: DB_PASSWORD
```

### Matrix Strategy

```yaml
version: v1alpha1
name: multi-platform-test
matrix:
  dimensions:
    os: ["ubuntu", "alpine"]
    go_version: ["1.21", "1.22"]
steps:
  - name: test
    image: golang:${{ matrix.go_version }}-${{ matrix.os }}
    commands:
      - go test ./...
# This creates 4 parallel pipeline runs:
#   ubuntu + 1.21, ubuntu + 1.22, alpine + 1.21, alpine + 1.22
```

### Conditional Execution

```yaml
version: v1alpha1
name: deploy-on-main
steps:
  - name: test
    image: golang:1.21
    commands:
      - go test ./...

  - name: deploy-staging
    image: ubuntu:22.04
    commands:
      - ./deploy.sh staging
    dependsOn: [test]
    conditional:
      branch: "develop"  # Only runs on develop branch

  - name: deploy-production
    image: ubuntu:22.04
    commands:
      - ./deploy.sh production
    dependsOn: [test]
    conditional:
      branch: "main"  # Only runs on main branch
```

---

## CLI Commands Reference

```bash
# List pipeline configurations
c8s config list
kubectl get pipelineconfig

# Get pipeline configuration details
c8s config get my-pipeline
kubectl get pipelineconfig my-pipeline -o yaml

# Create pipeline run (manual trigger)
c8s run my-pipeline --commit=abc123 --branch=main

# List pipeline runs
c8s runs list
c8s runs list --config=my-pipeline --phase=Running

# Get pipeline run status
c8s runs get my-pipeline-abc123
kubectl get pipelinerun my-pipeline-abc123

# Stream logs from running pipeline
c8s logs my-pipeline-abc123 --follow
c8s logs my-pipeline-abc123 --step=test

# Cancel running pipeline
c8s cancel my-pipeline-abc123
kubectl delete pipelinerun my-pipeline-abc123

# Validate pipeline configuration before committing
c8s validate .c8s.yaml
```

---

## Troubleshooting

### Pipeline Stuck in Pending

```bash
# Check if PipelineRun was created
kubectl get pipelinerun

# Check controller logs
kubectl logs -n c8s-system -l app=c8s-controller

# Check resource quotas
kubectl describe resourcequota -n default
```

### Step Failing Immediately

```bash
# View step Job
kubectl get job

# Check Job events
kubectl describe job my-pipeline-abc123-step-name

# View Pod logs
kubectl logs -l job-name=my-pipeline-abc123-step-name
```

### Webhook Not Triggering Pipelines

```bash
# Check RepositoryConnection status
kubectl get repositoryconnection my-repo-connection -o yaml

# Check webhook service logs
kubectl logs -n c8s-system -l app=c8s-webhook

# Verify webhook secret matches in GitHub and Kubernetes
kubectl get secret my-repo-webhook -o jsonpath='{.data.webhook-secret}' | base64 -d
```

### Logs Not Available

```bash
# Check object storage configuration
kubectl get secret c8s-storage -n c8s-system

# Check controller can access storage
kubectl logs -n c8s-system -l app=c8s-controller | grep storage

# Manually check S3 bucket
aws s3 ls s3://your-bucket/c8s-logs/
```

---

## Next Steps

- **Dashboard**: Install optional web UI for visual pipeline monitoring
- **Resource Quotas**: Configure per-team quotas with `ResourceQuota` objects
- **RBAC**: Set up namespace-scoped access control for teams
- **Autoscaling**: Configure Cluster Autoscaler to handle workload spikes
- **Artifacts**: Access build artifacts from object storage URLs in `status.steps[].artifactURLs`
- **Metrics**: Export pipeline metrics to Prometheus for monitoring

---

## Getting Help

- **Documentation**: https://docs.c8s.dev
- **GitHub Issues**: https://github.com/org/c8s/issues
- **Slack**: https://c8s.slack.com
- **Examples**: https://github.com/org/c8s-examples

---

## Architecture Summary

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

**Key Benefits**:
- ✅ Kubernetes-native (no external databases)
- ✅ Isolated execution (one Job per step)
- ✅ Declarative pipelines (GitOps-friendly)
- ✅ Secure secrets (K8s Secrets with log masking)
- ✅ Scalable (leverages K8s autoscaling)
- ✅ Observable (real-time logs and metrics)
