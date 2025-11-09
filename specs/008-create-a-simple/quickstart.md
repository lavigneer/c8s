# Quick Start: Deploy C8S to Kubernetes

**Feature**: Deploy C8S Stack to Kubernetes
**Date**: 2025-11-09

---

## Overview

The C8S deployment tool provides a simple way to deploy the entire C8S stack (API server, controller, webhook, frontend, and dependencies) to any Kubernetes cluster in under 5 minutes.

This quickstart will get you from zero to a working C8S instance using minimal commands.

---

## Prerequisites

Before deploying C8S, ensure you have:

- **Kubernetes Cluster** (1.24 or later)
  - Local: k3s, kind, minikube, or Docker Desktop
  - Cloud: EKS, GKE, AKS, or similar
- **kubectl** installed and configured to access your cluster
- **Container image registry** access (Docker Hub by default, or your organization's registry)
- **Internet connectivity** for pulling container images

**Verify prerequisites**:
```bash
# Check kubectl is installed
kubectl version --client

# Check cluster is accessible
kubectl cluster-info

# Check cluster version
kubectl version
# Server version should be 1.24 or later
```

---

## 5-Minute Deployment

### Step 1: Deploy C8S (1 minute)

```bash
# Deploy using defaults (dev environment, c8s-system namespace)
c8s-deploy deploy --environment dev
```

This command will:
- Validate your Kubernetes cluster (version, RBAC, API access)
- Download and apply C8S manifests
- Wait for all components to be ready
- Display the dashboard URL and access instructions

**Expected output**:
```
Validating Kubernetes cluster prerequisites...
✓ Kubernetes 1.25.3 detected
✓ RBAC enabled
✓ API server accessible

Deploying C8S components...
→ Deploying controller...
✓ Controller deployment created (3/3 replicas ready)
→ Deploying webhook...
✓ Webhook deployment created (1/1 replicas ready)
→ Deploying API server...
✓ API server deployment created (1/1 replicas ready)

✓ All C8S components deployed successfully!

Dashboard available at:
  kubectl port-forward svc/c8s-frontend 8080:80 -n c8s-system
```

### Step 2: Access the Dashboard (2 minutes)

In one terminal, forward the dashboard port:
```bash
kubectl port-forward svc/c8s-frontend 8080:80 -n c8s-system
```

In another terminal, get the default credentials:
```bash
# Get the admin password
kubectl get secret c8s-admin-creds -n c8s-system -o jsonpath='{.data.password}' | base64 -d
echo ""  # Add newline

# Username: admin
# Password: <output from above command>
```

Open your browser to **http://localhost:8080** and log in with:
- **Username**: `admin`
- **Password**: (from command above)

### Step 3: Verify Deployment (2 minutes)

Run a health check to verify all components are healthy:

```bash
c8s-deploy health
```

**Expected output**:
```
Health Check Results
====================

Overall Status: ✓ HEALTHY
Timestamp: 2025-11-09T14:23:45Z

✓ c8s-controller
  Replicas: 3/3 ready
  Status: Running

✓ c8s-webhook
  Replicas: 1/1 ready
  Status: Running

✓ c8s-api-server
  Replicas: 1/1 ready
  Status: Running

Dashboard: http://localhost:8080
```

✅ **Done!** Your C8S instance is deployed and ready to use.

---

## Common Scenarios

### Deploy to a Specific Namespace

```bash
c8s-deploy deploy --namespace my-namespace
```

### Deploy a Specific Version

```bash
c8s-deploy deploy --version v0.2.0
```

### Deploy to Production Environment

```bash
c8s-deploy deploy --environment production
```

This uses production presets:
- 3 replicas for high availability
- Higher resource limits
- Persistent storage configuration required

### Use Custom Configuration File

1. **Initialize a configuration file**:
```bash
c8s-deploy config init --environment production --output prod-config.yaml
```

2. **Edit the configuration** as needed:
```bash
# Edit prod-config.yaml with your settings
vim prod-config.yaml
```

3. **Deploy with the configuration**:
```bash
c8s-deploy deploy --config prod-config.yaml
```

### Validate Configuration Before Deploying

```bash
c8s-deploy config validate --config my-config.yaml
```

### Customize Container Image Registry

```bash
# Use a private registry
c8s-deploy deploy --image-registry myregistry.azurecr.io
```

---

## Accessing C8S

### Dashboard

```bash
# Port-forward to dashboard (in one terminal)
kubectl port-forward svc/c8s-frontend 8080:80 -n c8s-system

# Open browser to http://localhost:8080
```

### API Server

```bash
# Port-forward to API server (in one terminal)
kubectl port-forward svc/c8s-api 8081:8080 -n c8s-system

# Access API at http://localhost:8081
# API documentation at http://localhost:8081/api/docs
```

### Logs

View logs from any component:

```bash
# Controller logs
kubectl logs deployment/c8s-controller -n c8s-system

# Webhook logs
kubectl logs deployment/c8s-webhook -n c8s-system

# API server logs
kubectl logs deployment/c8s-api-server -n c8s-system

# Follow logs in real-time
kubectl logs -f deployment/c8s-api-server -n c8s-system

# View logs from all pods of a component
kubectl logs -l app=c8s-controller -n c8s-system --all-containers=true
```

---

## Customization Examples

### Development Setup (Minimal Resources)

```bash
c8s-deploy deploy \
  --environment dev \
  --namespace c8s-dev
```

Default configuration:
- 1 replica per component
- Minimal CPU/memory requests
- Local storage or MinIO

### Staging Setup (Medium Resources)

```bash
# Create configuration
c8s-deploy config init --environment staging --output staging-config.yaml

# Deploy
c8s-deploy deploy --config staging-config.yaml
```

### Production Setup with S3

```bash
# Create production configuration
c8s-deploy config init --environment production --output prod-config.yaml

# Edit to add S3 configuration
vim prod-config.yaml
# Update storage section with your S3 endpoint, bucket, credentials

# Deploy
c8s-deploy deploy --config prod-config.yaml --namespace c8s-prod
```

Example S3 configuration in YAML:
```yaml
storage:
  type: s3-compatible
  endpoint: s3.amazonaws.com
  bucket: my-org-c8s-logs
  region: us-west-2
  accessKey: AKIAIOSFODNN7EXAMPLE
  secretKey: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
  useSSL: true
```

---

## Troubleshooting

### Deployment Stuck or Timeout

```bash
# Check deployment status
kubectl get deployments -n c8s-system

# Describe a specific deployment to see issues
kubectl describe deployment c8s-controller -n c8s-system

# View events in namespace
kubectl get events -n c8s-system --sort-by='.lastTimestamp'
```

### Pod Pending or CrashLoopBackOff

```bash
# Check pod status
kubectl get pods -n c8s-system

# View logs from the problematic pod
kubectl logs <pod-name> -n c8s-system

# Describe pod for detailed status
kubectl describe pod <pod-name> -n c8s-system
```

### Image Pull Errors

```bash
# Verify image exists and is accessible
docker pull docker.io/org/c8s-api-server:v0.2.0

# Check image pull errors in pod events
kubectl describe pod <pod-name> -n c8s-system

# Fix: Update image version or registry
c8s-deploy deploy --version v0.2.0 --image-registry myregistry.com
```

### Insufficient Resources

```bash
# Check node resources
kubectl top nodes

# Check pod resource requests vs available
kubectl describe nodes

# Fix: Add more nodes or reduce resource requests
c8s-deploy config set components.controller.resources.requests.cpu 250m
c8s-deploy deploy --config modified-config.yaml
```

### Health Check Shows Issues

```bash
# Run detailed health check
c8s-deploy health --verbose

# Check component-specific logs for errors
kubectl logs deployment/c8s-controller -n c8s-system
kubectl logs deployment/c8s-api-server -n c8s-system

# View component status
kubectl get deployment c8s-controller -n c8s-system -o wide
kubectl describe deployment c8s-controller -n c8s-system
```

---

## Next Steps After Deployment

### 1. Create Your First Pipeline

Access the C8S dashboard and create a test pipeline to verify everything is working:

1. Log in to http://localhost:8080
2. Navigate to "Pipelines" → "Create Pipeline"
3. Set up a simple test pipeline
4. Run it to verify all components are functioning

### 2. Configure Storage

For production use, configure persistent storage:

```bash
# Check current storage
kubectl get pvc -n c8s-system

# Update storage configuration in config file
c8s-deploy config set storage.type s3-compatible
c8s-deploy config set storage.endpoint s3.amazonaws.com
```

### 3. Set Up Logging

View logs from your pipelines:

```bash
# Check logs from a pipeline run
kubectl logs -l app=c8s-controller -n c8s-system
```

### 4. Configure Authentication

For production, set up proper authentication:

```bash
# Update authentication in config
c8s-deploy config set authentication.type oauth2
c8s-deploy config set authentication.secretRef c8s-oauth-secret

# Deploy the changes
c8s-deploy deploy --config prod-config.yaml
```

### 5. Set Up Monitoring

Monitor your C8S deployment:

```bash
# Get Prometheus metrics (if enabled)
kubectl port-forward svc/c8s-prometheus 9090:9090 -n c8s-system

# Get Grafana dashboards (if enabled)
kubectl port-forward svc/c8s-grafana 3000:3000 -n c8s-system
```

---

## Upgrading C8S

When a new version is available:

```bash
# Upgrade to a specific version
c8s-deploy upgrade --to-version v0.3.0

# Check status during upgrade
c8s-deploy health --verbose

# Rollback if needed (automatic backup created during upgrade)
c8s-deploy rollback --backup c8s-backup-20251109-142345
```

---

## Uninstalling C8S

To cleanly remove C8S from your cluster:

```bash
# Uninstall (interactive confirmation)
c8s-deploy uninstall --namespace c8s-system

# Uninstall with force (no confirmation)
c8s-deploy uninstall --namespace c8s-system --force

# Keep namespace and data for redeployment
c8s-deploy uninstall --keep-namespace --keep-data
```

---

## Getting Help

### View Command Help

```bash
# General help
c8s-deploy --help

# Help for specific command
c8s-deploy deploy --help
c8s-deploy health --help
c8s-deploy config --help
```

### Enable Debug Logging

```bash
# Show detailed debug output
c8s-deploy deploy --verbose --debug
```

### Check C8S Logs

```bash
# All C8S logs
kubectl logs -l app=c8s -n c8s-system --all-containers=true

# Recent logs only
kubectl logs -l app=c8s -n c8s-system --since=10m
```

---

## Configuration Reference

For detailed configuration options, see:
- `contracts/cli-api.md` - Complete CLI command reference
- `contracts/config-schema.json` - Configuration schema with validation
- `data-model.md` - Data model and entities

---

## What's Next?

Now that C8S is deployed, you can:

1. **Create pipelines** for your CI/CD workflows
2. **Configure integrations** with GitHub, GitLab, or your Git provider
3. **Set up webhooks** for automatic pipeline triggering
4. **Monitor pipeline runs** and view logs
5. **Download artifacts** from successful pipeline runs

For detailed documentation, visit the C8S documentation site or check the `/docs` directory in the C8S repository.
