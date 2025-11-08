# Getting Started with C8S

Welcome to C8S! This guide will help you get up and running with Kubernetes-native continuous integration in less than 5 minutes.

## Table of Contents

1. [Installation (5 minutes)](#installation)
2. [First Pipeline](#first-pipeline)
3. [Dashboard Navigation](#dashboard-navigation)
4. [Next Steps](#next-steps)
5. [Quick Troubleshooting](#quick-troubleshooting)

---

## Installation

### Prerequisites

Before you begin, make sure you have:
- **Kubernetes cluster** (1.24+) - Can be local (k3d) or cloud (GKE, EKS, AKS)
- **kubectl** configured to access your cluster
- **git** and **docker** for building containers (optional)

### Quick Install (2 minutes)

**Step 1: Install CRDs**

```bash
kubectl apply -f https://raw.githubusercontent.com/org/c8s/main/deploy/crds.yaml
```

**Step 2: Install C8S Components**

```bash
kubectl apply -f https://raw.githubusercontent.com/org/c8s/main/deploy/install.yaml
```

**Step 3: Wait for readiness**

```bash
kubectl rollout status deployment/c8s-controller -n c8s-system
kubectl rollout status deployment/c8s-api-server -n c8s-system
kubectl rollout status deployment/c8s-webhook -n c8s-system
```

That's it! C8S is now installed and ready to use.

### Verify Installation

```bash
# Check all components are running
kubectl get pods -n c8s-system

# Expected output:
# NAME                                 READY   STATUS    RESTARTS   AGE
# c8s-controller-xxxxx                1/1     Running   0          1m
# c8s-api-server-xxxxx                1/1     Running   0          1m
# c8s-webhook-xxxxx                   1/1     Running   0          1m
```

### Local Development Setup (Alternative)

For developing or testing C8S locally:

```bash
# Install prerequisites
brew install k3d tilt kubectl

# Start development environment
cd c8s
tilt up

# Tilt automatically:
# - Creates local k3d cluster
# - Deploys C8S components
# - Opens dashboard at http://localhost:10350
```

See [docs/tilt-setup.md](./tilt-setup.md) for detailed local setup guide.

---

## First Pipeline

### Create Your First Pipeline Config

**Step 1: Define a simple pipeline**

Create a file called `pipeline.yaml`:

```yaml
apiVersion: c8s.io/v1alpha1
kind: PipelineConfig
metadata:
  name: hello-world
  namespace: default
spec:
  repository: https://github.com/your-org/your-repo.git
  branches: ["main"]
  steps:
    - name: greet
      image: alpine:latest
      commands:
        - echo "Hello from C8S!"
        - echo "Current working directory:"
        - pwd
    - name: list-files
      image: alpine:latest
      commands:
        - ls -la
      dependsOn:
        - greet  # This step runs after 'greet' completes
```

**Step 2: Create the pipeline**

```bash
kubectl apply -f pipeline.yaml
```

**Step 3: Trigger a run**

Manually trigger the pipeline:

```bash
# Create a PipelineRun
kubectl create -f - <<EOF
apiVersion: c8s.io/v1alpha1
kind: PipelineRun
metadata:
  name: hello-world-run-1
  namespace: default
spec:
  pipelineConfigRef: hello-world
  branch: main
  commit: abc123def456
  author: "Your Name"
EOF
```

Or set up a webhook for automatic triggers (see [docs/WEBHOOK_INTEGRATION.md](./WEBHOOK_INTEGRATION.md)).

**Step 4: View the run**

```bash
# Watch the pipeline execution
kubectl get pipelinerun -w

# View logs from a specific step
kubectl logs pipelinerun/hello-world-run-1 -c greet

# View pipeline status
kubectl describe pipelinerun hello-world-run-1
```

### Dashboard View (Optional)

If you have the dashboard installed:

```bash
# Port-forward to the dashboard
kubectl port-forward svc/c8s-api-server 8080:8080 -n c8s-system

# Open browser to http://localhost:8080/dashboard
```

---

## Dashboard Navigation

### Accessing the Dashboard

If C8S dashboard is installed:

```bash
# Port-forward the API server
kubectl port-forward -n c8s-system svc/c8s-api-server 8080:8080

# Open in browser
# http://localhost:8080/dashboard
```

### Main Pages

#### **Dashboard Home**
- **View**: Quick stats panel showing:
  - Total pipeline runs
  - Success rate (%)
  - Failed runs count
  - Currently running pipelines
- **Action**: Click on any stat to filter and view details

#### **Projects Page** (`/dashboard/projects`)
- **View**: All pipeline configurations (projects)
- **Columns**:
  - Project name
  - Repository URL
  - Last run status
  - Last run timestamp
- **Actions**:
  - Click project name to view recent runs
  - Click "Run Now" to trigger a new run
  - Click settings icon for configuration

#### **Pipeline Runs** (`/dashboard/runs`)
- **View**: Historical list of all pipeline runs
- **Columns**:
  - Run ID
  - Project name
  - Status (success/failed/running)
  - Commit SHA
  - Branch
  - Duration
  - Trigger time
- **Actions**:
  - Click run to view detailed logs
  - Use filters (status, branch, date range)
  - Export run data

#### **Run Details** (`/dashboard/runs/{runId}`)
- **View**: Complete pipeline run execution details
- **Sections**:
  - Run metadata (ID, commit, author, triggered at)
  - Steps execution timeline
  - Real-time logs for each step
  - Artifacts (if any)
  - Resource usage
- **Logs Tab**:
  - Stream logs in real-time
  - Search logs by keyword
  - Copy log content
  - Download full logs as text file

### Keyboard Shortcuts

- `g` + `d` - Go to Dashboard
- `g` + `p` - Go to Projects
- `g` + `r` - Go to Runs
- `?` - Show help menu
- `f` - Focus search filter
- `/` - Quick search across all runs

### Filtering & Search

**By Status**:
```
Status: success
Status: failed
Status: running
```

**By Branch**:
```
Branch: main
Branch: develop
```

**By Date Range**:
```
After: 2025-11-01
Before: 2025-11-02
```

**Combine Filters**:
```
Status: success Branch: main
```

---

## Next Steps

### 1. Set Up Webhook Integration

Connect your Git repository to automatically trigger pipelines:

- [GitHub Integration](./WEBHOOK_INTEGRATION.md#github-setup)
- [GitLab Integration](./WEBHOOK_INTEGRATION.md#gitlab-setup)
- [Bitbucket Integration](./WEBHOOK_INTEGRATION.md#bitbucket-setup)

### 2. Configure Advanced Features

- **Secrets Management**: [SECURITY.md#secret-management](./SECURITY.md#secret-management)
- **Resource Limits**: [CONFIGURATION.md#resource-limits](./CONFIGURATION.md#resource-limits)
- **Matrix Builds**: [PIPELINE_SYNTAX.md#matrix-builds](./PIPELINE_SYNTAX.md#matrix-builds)
- **Conditional Execution**: [PIPELINE_SYNTAX.md#conditional-execution](./PIPELINE_SYNTAX.md#conditional-execution)

### 3. Integrate with Your Workflow

- **CLI Usage**: See [CLI_REFERENCE.md](./CLI_REFERENCE.md)
- **API Integration**: See [API specifications](./API_REFERENCE.md)
- **GitOps**: Manage pipelines via version control

### 4. Production Deployment

- **High Availability**: [OPERATOR_GUIDE.md#high-availability](./OPERATOR_GUIDE.md#high-availability)
- **Persistence**: [CONFIGURATION.md#persistence](./CONFIGURATION.md#persistence)
- **Monitoring**: [OPERATOR_GUIDE.md#monitoring](./OPERATOR_GUIDE.md#monitoring)
- **Backup & Recovery**: [OPERATOR_GUIDE.md#backup-recovery](./OPERATOR_GUIDE.md#backup-recovery)

---

## Quick Troubleshooting

### Components Not Starting

**Symptom**: Pods in Pending or CrashLoopBackOff state

**Solution**:
```bash
# Check logs
kubectl logs deployment/c8s-controller -n c8s-system
kubectl logs deployment/c8s-api-server -n c8s-system

# Check events
kubectl describe pod <pod-name> -n c8s-system

# Ensure cluster resources
kubectl top nodes
kubectl top pods -n c8s-system
```

See full [TROUBLESHOOTING.md](./TROUBLESHOOTING.md) guide.

### Pipeline Run Stuck

**Symptom**: PipelineRun shows "Running" but Job is not executing

**Solution**:
```bash
# Check PipelineRun status
kubectl describe pipelinerun <run-name>

# Check created Jobs
kubectl get jobs -n c8s-system | grep <run-name>

# Check for resource quotas
kubectl describe resourcequota -n default
```

### Connection Issues

**Symptom**: Cannot connect to dashboard or API

**Solution**:
```bash
# Verify service exists
kubectl get svc c8s-api-server -n c8s-system

# Check port-forwarding
kubectl port-forward -n c8s-system svc/c8s-api-server 8080:8080

# Verify network connectivity
kubectl exec -it <pod-name> -n c8s-system -- /bin/sh
curl http://localhost:8080/health
```

### Webhook Not Triggering

**Symptom**: Push to repo doesn't create PipelineRun

**Solution**:
```bash
# Verify webhook is configured in Git platform

# Check webhook service logs
kubectl logs deployment/c8s-webhook -n c8s-system

# Verify webhook secret is created
kubectl get secret -n c8s-system | grep webhook

# Test webhook manually
curl -X POST http://webhook-service:8080/webhooks/github \
  -H "X-Hub-Signature-256: sha256=..." \
  -d '{"ref":"refs/heads/main",...}'
```

For more help, see [TROUBLESHOOTING.md](./TROUBLESHOOTING.md).

---

## Getting Help

- **Documentation**: Browse [docs/](./docs/)
- **Issues**: GitHub Issues for bug reports
- **Discussions**: GitHub Discussions for questions
- **Slack**: Join our community Slack (if available)

---

## What's Next?

🎉 You've successfully set up C8S! Now you can:

1. ✅ Create and run pipelines
2. ✅ Monitor progress via dashboard
3. ✅ Integrate with your Git platform
4. ✅ Scale to production

For comprehensive documentation, see the [docs](./docs/) directory.

Happy CI/CD! 🚀

