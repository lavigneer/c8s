# C8S Operator Guide

**Version**: 1.0
**Audience**: Platform operators, DevOps engineers, SREs
**Last Updated**: 2025-11-02

Complete guide to deploying, operating, and maintaining C8S in production environments.

## Table of Contents

- [Installation](#installation)
- [Namespaces and RBAC](#namespaces-and-rbac)
- [Resource Management](#resource-management)
- [High Availability](#high-availability)
- [Backup and Recovery](#backup-and-recovery)
- [Upgrade Procedures](#upgrade-procedures)
- [Monitoring and Observability](#monitoring-and-observability)
- [Troubleshooting](#troubleshooting)

---

## Installation

### Prerequisites

- Kubernetes 1.24+
- kubectl configured with cluster access
- Sufficient cluster resources (see [Resource Requirements](#resource-requirements))
- Optional: Helm 3.0+ for Helm-based installation

### Quick Installation

#### Option 1: Kubectl Apply

```bash
# Create namespace
kubectl create namespace c8s-system

# Apply CRDs
kubectl apply -f https://raw.githubusercontent.com/org/c8s/main/config/crds/pipelineconfig_crd.yaml
kubectl apply -f https://raw.githubusercontent.com/org/c8s/main/config/crds/pipelinerun_crd.yaml

# Apply RBAC
kubectl apply -f https://raw.githubusercontent.com/org/c8s/main/config/rbac/

# Deploy C8S
kubectl apply -f https://raw.githubusercontent.com/org/c8s/main/deploy/c8s-deployment.yaml
```

#### Option 2: Helm Chart

```bash
# Add C8S Helm repository
helm repo add c8s https://charts.c8s.dev
helm repo update

# Install C8S
helm install c8s c8s/c8s \
  --namespace c8s-system \
  --create-namespace \
  --values values.yaml
```

Example `values.yaml`:

```yaml
controller:
  replicas: 2
  resources:
    requests:
      cpu: 100m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 512Mi

apiServer:
  replicas: 2
  port: 8080

storage:
  type: s3
  s3:
    endpoint: https://minio.example.com
    bucket: c8s-logs
    region: us-east-1

auth:
  enabled: true
  mode: jwt
  jwtSecret: your-secret-here
```

### Verify Installation

```bash
# Check deployments
kubectl get deployments -n c8s-system

# Check pods are running
kubectl get pods -n c8s-system

# Check CRDs are installed
kubectl get crd | grep c8s

# Test API server
kubectl port-forward -n c8s-system svc/c8s-api 8080:8080 &
curl http://localhost:8080/health
```

---

## Namespaces and RBAC

### Namespace Structure

Recommended namespace organization:

```
Kubernetes Cluster
├── c8s-system                 # C8S components
│   ├── Deployments (controller, api-server, webhook)
│   ├── Services (c8s-api, c8s-webhook)
│   ├── ConfigMaps & Secrets
│   └── RBAC Resources
│
├── team-a                     # Team A projects
│   └── PipelineConfigs, PipelineRuns
│
├── team-b                     # Team B projects
│   └── PipelineConfigs, PipelineRuns
│
└── c8s-projects              # Shared projects (optional)
    └── PipelineConfigs, PipelineRuns
```

### RBAC Setup

#### Controller RBAC

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: c8s-controller
rules:
  # Read PipelineConfigs and PipelineRuns
  - apiGroups: ["c8s.dev"]
    resources: ["pipelineconfigs", "pipelineruns"]
    verbs: ["get", "list", "watch"]

  # Create/update Jobs
  - apiGroups: ["batch"]
    resources: ["jobs"]
    verbs: ["create", "get", "list", "watch"]

  # Read Pods
  - apiGroups: [""]
    resources: ["pods", "pods/log"]
    verbs: ["get", "list", "watch"]

  # Read/write ConfigMaps for metadata
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "create", "update"]
```

#### User RBAC (Editor Role)

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: c8s-editor
  namespace: team-a
rules:
  # Create and manage pipelines
  - apiGroups: ["c8s.dev"]
    resources: ["pipelineconfigs", "pipelineruns"]
    verbs: ["get", "list", "create", "update", "delete"]
```

#### User RBAC (Viewer Role)

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: c8s-viewer
  namespace: team-a
rules:
  # Read-only access to pipelines
  - apiGroups: ["c8s.dev"]
    resources: ["pipelineconfigs", "pipelineruns"]
    verbs: ["get", "list"]

  # Read logs
  - apiGroups: [""]
    resources: ["pods/log"]
    verbs: ["get"]
```

### Bind Roles to Users

```bash
# Make user an editor in team-a namespace
kubectl create rolebinding alice-editor \
  --clusterrole=c8s-editor \
  --user=alice@example.com \
  -n team-a

# Make user a viewer across all namespaces
kubectl create clusterrolebinding bob-viewer \
  --clusterrole=c8s-viewer \
  --user=bob@example.com
```

---

## Resource Management

### Resource Requirements

#### Minimum Configuration (Development)

```yaml
# Controller
controller:
  requests:
    cpu: 50m
    memory: 128Mi
  limits:
    cpu: 200m
    memory: 256Mi

# API Server
apiServer:
  requests:
    cpu: 50m
    memory: 128Mi
  limits:
    cpu: 200m
    memory: 256Mi

# Webhook
webhook:
  requests:
    cpu: 25m
    memory: 64Mi
  limits:
    cpu: 100m
    memory: 128Mi
```

#### Recommended Configuration (Production)

```yaml
# Controller
controller:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 1000m
    memory: 1Gi

# API Server
apiServer:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 1000m
    memory: 1Gi

# Webhook
webhook:
  requests:
    cpu: 250m
    memory: 256Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

### Resource Quotas

Limit resource consumption per team/namespace:

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-a-quota
  namespace: team-a
spec:
  hard:
    requests.cpu: "10"
    requests.memory: "20Gi"
    limits.cpu: "20"
    limits.memory: "40Gi"
    pods: "100"

  scopeSelector:
    matchExpressions:
      - operator: In
        scopeName: PriorityClass
        values: ["default", "high"]
```

### Network Policies

Restrict traffic between namespaces:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: c8s-api-allow
  namespace: c8s-system
spec:
  podSelector:
    matchLabels:
      app: c8s-api
  policyTypes:
    - Ingress
  ingress:
    # Allow from dashboard
    - from:
        - namespaceSelector:
            matchLabels:
              name: dashboards
      ports:
        - protocol: TCP
          port: 8080

    # Allow from webhooks
    - from:
        - podSelector:
            matchLabels:
              app: c8s-webhook
      ports:
        - protocol: TCP
          port: 8080
```

---

## High Availability

### Multi-Replica Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: c8s-controller
  namespace: c8s-system
spec:
  replicas: 2  # HA configuration

  selector:
    matchLabels:
      app: c8s-controller

  template:
    metadata:
      labels:
        app: c8s-controller
    spec:
      # Anti-affinity to spread across nodes
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: app
                      operator: In
                      values:
                        - c8s-controller
                topologyKey: kubernetes.io/hostname

      containers:
      - name: controller
        image: c8s:latest
        env:
          - name: CONTROLLER_LEADER_ELECT
            value: "true"
```

### Pod Disruption Budgets

Ensure minimum availability during maintenance:

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: c8s-controller-pdb
  namespace: c8s-system
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: c8s-controller

---
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: c8s-api-pdb
  namespace: c8s-system
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: c8s-api
```

---

## Backup and Recovery

### Backup Strategy

Essential data to backup:

1. **Kubernetes etcd** - Cluster state (includes CRDs)
2. **ConfigMaps/Secrets** - C8S configuration
3. **Object Storage** - Pipeline logs and artifacts
4. **Database** (if used) - Pipeline history, metadata

### Backup Implementation

```bash
#!/bin/bash
# backup-c8s.sh - Backup C8S data

BACKUP_DIR="/backups/c8s/$(date +%Y-%m-%d_%H-%M-%S)"
mkdir -p $BACKUP_DIR

# Backup Kubernetes resources
echo "Backing up Kubernetes resources..."
kubectl get all -A -o yaml > $BACKUP_DIR/k8s-resources.yaml
kubectl get secrets -A -o yaml > $BACKUP_DIR/secrets.yaml
kubectl get configmaps -A -o yaml > $BACKUP_DIR/configmaps.yaml

# Backup etcd (if you have direct access)
# ETCDCTL_API=3 etcdctl snapshot save $BACKUP_DIR/etcd-snapshot.db

# Backup S3 logs
echo "Backing up object storage..."
aws s3 sync s3://c8s-logs $BACKUP_DIR/s3-backup \
  --profile c8s-backup

echo "Backup completed to $BACKUP_DIR"
```

### Recovery Procedures

#### Restore from Backup

```bash
# Restore Kubernetes resources
kubectl apply -f backup-dir/k8s-resources.yaml

# Restore secrets (carefully)
kubectl apply -f backup-dir/secrets.yaml

# Restore etcd (full cluster recovery)
# etcdctl snapshot restore backup-dir/etcd-snapshot.db \
#   --data-dir=/var/lib/etcd
```

---

## Upgrade Procedures

### Pre-Upgrade Checklist

- [ ] Backup etcd and all data
- [ ] Verify no running pipelines (or wait for completion)
- [ ] Check storage capacity
- [ ] Document current version
- [ ] Notify users of maintenance window

### Rolling Upgrade

```bash
# Get current version
kubectl get deployment c8s-controller -n c8s-system \
  -o jsonpath='{.spec.template.spec.containers[0].image}'

# Update controller deployment
kubectl set image deployment/c8s-controller \
  controller=c8s:v1.2.0 \
  -n c8s-system

# Monitor rollout
kubectl rollout status deployment/c8s-controller -n c8s-system

# Update API server
kubectl set image deployment/c8s-api \
  api-server=c8s:v1.2.0 \
  -n c8s-system

# Wait for completion
kubectl rollout status deployment/c8s-api -n c8s-system

# Verify
kubectl get pods -n c8s-system
```

### Rollback

```bash
# Rollback controller
kubectl rollout undo deployment/c8s-controller -n c8s-system

# Rollback API server
kubectl rollout undo deployment/c8s-api -n c8s-system

# Verify
kubectl rollout status deployment/c8s-controller -n c8s-system
kubectl rollout status deployment/c8s-api -n c8s-system
```

---

## Monitoring and Observability

### Prometheus Metrics

Example Prometheus scrape config:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'c8s-controller'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - c8s-system
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: c8s-controller

  - job_name: 'c8s-api'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - c8s-system
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        action: keep
        regex: c8s-api
```

### Key Metrics to Monitor

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `c8s_pipeline_runs_total` | Total pipeline runs | - |
| `c8s_pipeline_runs_success` | Successful runs | - |
| `c8s_pipeline_runs_failed` | Failed runs | > 10% failure rate |
| `c8s_controller_queue_depth` | Controller queue size | > 100 |
| `c8s_api_requests_duration_seconds` | API request latency | p99 > 1s |
| `c8s_api_errors_total` | Total API errors | > 5% error rate |

### Logging Configuration

Forward logs to centralized system:

```yaml
# Fluent Bit configuration
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
  namespace: c8s-system
data:
  fluent-bit.conf: |
    [SERVICE]
        Log_Level info

    [INPUT]
        Name tail
        Tag c8s.*
        Path /var/log/containers/*c8s*.log
        Parser cri
        DB /var/log/flb_c8s.db

    [OUTPUT]
        Name stackdriver
        Match c8s.*
        google_service_credentials /var/secrets/google/key.json
        resource k8s_cluster
        k8s_cluster_name my-cluster
        k8s_cluster_location us-central1
```

---

## Troubleshooting

### Common Issues

#### Issue: Pods Stuck in Pending

```bash
# Check pod events
kubectl describe pod <pod-name> -n c8s-system

# Check node resources
kubectl top nodes
kubectl describe node <node-name>

# Solution: Add more nodes or increase resource requests
```

#### Issue: API Server Not Responding

```bash
# Check API server logs
kubectl logs -f deployment/c8s-api -n c8s-system

# Check service connectivity
kubectl exec -it <pod> -n c8s-system -- curl localhost:8080/health

# Check network policies
kubectl get networkpolicies -n c8s-system
```

#### Issue: Pipeline Runs Timing Out

```bash
# Check controller logs for errors
kubectl logs -f deployment/c8s-controller -n c8s-system

# Check job creation
kubectl get jobs -n c8s-system

# Increase controller workers
kubectl set env deployment/c8s-controller \
  CONTROLLER_WORKERS=5 -n c8s-system
```

### Debug Commands

```bash
# Get all C8S resources
kubectl get pipelineconfigs,pipelineruns -A

# Watch a pipeline run
kubectl get pipelinerun <run-name> -n <ns> -w

# Check controller status
kubectl get deployment c8s-controller -n c8s-system -o wide

# View recent events
kubectl get events -n c8s-system --sort-by='.lastTimestamp' | tail -20
```

---

## Related Documentation

- [Getting Started](./GETTING_STARTED.md) - Quick start guide
- [Configuration](./CONFIGURATION.md) - Configuration reference
- [Troubleshooting](./TROUBLESHOOTING.md) - Troubleshooting guide
