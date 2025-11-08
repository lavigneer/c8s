# Autoscaling Integration

This document explains how C8S integrates with Kubernetes Cluster Autoscaler to automatically scale cluster capacity based on CI workload demands.

## Overview

C8S leverages Kubernetes' native autoscaling capabilities:

1. **Cluster Autoscaler**: Watches for pending Pods and automatically adds/removes nodes
2. **Horizontal Pod Autoscaler** (optional): Scales the number of controller replicas
3. **Resource Requests**: PipelineRun Jobs specify resource requests that trigger scaling

## How It Works

### Pod Scheduling and Autoscaling

```
PipelineRun created
    ↓
Controller creates Jobs with resource requests
    ↓
Kubernetes Scheduler tries to schedule Pods
    ↓
If insufficient capacity → Pods remain Pending
    ↓
Cluster Autoscaler detects pending Pods
    ↓
New nodes added to cluster (30-120 seconds)
    ↓
Pods schedule on new nodes
    ↓
Pipeline execution begins
```

### Scale Down

```
Jobs complete
    ↓
Pods deleted
    ↓
Nodes become underutilized
    ↓
Cluster Autoscaler waits for scale-down delay (10 minutes default)
    ↓
Unused nodes drained and deleted
    ↓
Cluster size reduced
```

## Setting Up Cluster Autoscaler

### AWS EKS

```bash
# Install Cluster Autoscaler
kubectl apply -f https://raw.githubusercontent.com/kubernetes/autoscaler/master/cluster-autoscaler/cloudprovider/aws/examples/cluster-autoscaler-autodiscover.yaml

# Configure for your cluster
kubectl -n kube-system edit deployment cluster-autoscaler

# Add cluster name to command args:
# - --node-group-auto-discovery=asg:tag=k8s.io/cluster-autoscaler/enabled,k8s.io/cluster-autoscaler/<YOUR-CLUSTER-NAME>
# - --balance-similar-node-groups
# - --skip-nodes-with-system-pods=false
```

### GKE

```bash
# GKE has built-in autoscaling, enable during cluster creation:
gcloud container clusters create c8s-cluster \
  --enable-autoscaling \
  --min-nodes=1 \
  --max-nodes=50 \
  --zone=us-central1-a
```

### Azure AKS

```bash
# Enable autoscaler on node pool
az aks nodepool update \
  --enable-cluster-autoscaler \
  --min-count 1 \
  --max-count 50 \
  --name default \
  --cluster-name c8s-cluster \
  --resource-group c8s-rg
```

## Node Affinity for CI Workloads

### Dedicated Node Pool for CI

Create a dedicated node pool for CI workloads to isolate them from application workloads:

**AWS (eksctl)**:
```yaml
# nodegroup-ci.yaml
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: c8s-cluster
  region: us-west-2
nodeGroups:
  - name: ci-workers
    instanceType: c5.4xlarge
    minSize: 0
    maxSize: 20
    desiredCapacity: 2
    labels:
      workload: ci
    taints:
      - key: ci-workload
        value: "true"
        effect: NoSchedule
```

Apply:
```bash
eksctl create nodegroup -f nodegroup-ci.yaml
```

**GKE**:
```bash
gcloud container node-pools create ci-workers \
  --cluster=c8s-cluster \
  --machine-type=n1-standard-8 \
  --enable-autoscaling \
  --min-nodes=0 \
  --max-nodes=20 \
  --node-labels=workload=ci \
  --node-taints=ci-workload=true:NoSchedule
```

### Configure C8S to Use Dedicated Nodes

Update `pkg/controller/job_manager.go` to add node affinity and tolerations to Jobs:

```go
// Add to Job Pod spec
Affinity: &corev1.Affinity{
    NodeAffinity: &corev1.NodeAffinity{
        RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
            NodeSelectorTerms: []corev1.NodeSelectorTerm{
                {
                    MatchExpressions: []corev1.NodeSelectorRequirement{
                        {
                            Key:      "workload",
                            Operator: corev1.NodeSelectorOpIn,
                            Values:   []string{"ci"},
                        },
                    },
                },
            },
        },
    },
},
Tolerations: []corev1.Toleration{
    {
        Key:      "ci-workload",
        Operator: corev1.TolerationOpEqual,
        Value:    "true",
        Effect:   corev1.TaintEffectNoSchedule,
    },
},
```

## Scaling Behavior

### Scale-Up Timing

| Event | Time |
|-------|------|
| PipelineRun created | T+0s |
| Jobs created | T+1s |
| Pods pending (insufficient capacity) | T+2s |
| Cluster Autoscaler detects pending Pods | T+10s |
| Cloud provider API call to add nodes | T+15s |
| Nodes join cluster and ready | T+60-120s |
| Pods schedule on new nodes | T+65-125s |

**Expected scale-up time**: 1-2 minutes from pending to running

### Scale-Down Timing

| Event | Time |
|-------|------|
| Jobs complete | T+0s |
| Pods deleted | T+1s |
| Nodes underutilized | T+2s |
| Scale-down delay (configurable) | T+10 minutes |
| Nodes drained | T+10m + 30s |
| Nodes deleted | T+10m + 60s |

**Expected scale-down time**: 10-15 minutes after idle

### Configure Scale-Down Delay

```bash
# Edit Cluster Autoscaler deployment
kubectl -n kube-system edit deployment cluster-autoscaler

# Add or modify:
# - --scale-down-delay-after-add=10m
# - --scale-down-unneeded-time=10m
# - --scale-down-unready-time=20m
```

For aggressive scale-down (cost optimization):
```yaml
- --scale-down-delay-after-add=5m
- --scale-down-unneeded-time=5m
```

For conservative scale-down (performance optimization):
```yaml
- --scale-down-delay-after-add=30m
- --scale-down-unneeded-time=30m
```

## Resource Requests Best Practices

### Set Accurate Resource Requests

```yaml
version: v1alpha1
name: my-pipeline
steps:
  - name: test
    image: golang:1.21
    commands:
      - go test ./...
    resources:
      cpu: 2000m        # 2 CPU cores
      memory: 4Gi       # 4 GB RAM
```

**Why it matters**:
- **Too low**: Pods may be OOMKilled or throttled
- **Too high**: Wastes capacity, increases costs, delays scaling
- **Just right**: Efficient scheduling, accurate autoscaling triggers

### Default Resources

If not specified, C8S uses defaults:
- CPU: 1 core (1000m)
- Memory: 2Gi

Override defaults by setting resource requests in pipeline config.

## Priority Classes

Use priority classes to ensure critical pipelines run first when capacity is limited:

```yaml
# high-priority-class.yaml
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: ci-high-priority
value: 1000
globalDefault: false
description: "High priority for production CI pipelines"

---
apiVersion: scheduling.k8s.io/v1
kind: PriorityClass
metadata:
  name: ci-low-priority
value: 100
globalDefault: false
description: "Low priority for development CI pipelines"
```

Apply priority to PipelineRuns via labels (requires controller support):
```yaml
apiVersion: c8s.dev/v1alpha1
kind: PipelineRun
metadata:
  name: prod-pipeline-run
  labels:
    c8s.dev/priority: high
spec:
  pipelineConfigRef: prod-pipeline
  # ...
```

## Monitoring Autoscaling

### Check Cluster Autoscaler Logs

```bash
kubectl logs -n kube-system deployment/cluster-autoscaler -f
```

Look for:
- `Expanding node group` - scale-up triggered
- `Scale-down: node <name> is unneeded` - scale-down detected
- `Scale-down: removing node <name>` - scale-down executed

### Check Pending Pods

```bash
# View all pending Pods
kubectl get pods --all-namespaces --field-selector=status.phase=Pending

# Check why Pods are pending
kubectl describe pod <pod-name>
```

Common reasons:
- `Insufficient cpu` - need more CPU capacity
- `Insufficient memory` - need more memory capacity
- `Unschedulable` - node affinity/taints not satisfied

### Metrics

Monitor these Prometheus metrics:

```promql
# Pending pipeline Jobs
sum(kube_job_status_active{job=~".*c8s.*"})

# Node count
count(kube_node_info)

# Node utilization
avg(1 - rate(node_cpu_seconds_total{mode="idle"}[5m]))
```

## Cost Optimization

### Use Spot/Preemptible Instances

**AWS EKS (Spot Instances)**:
```yaml
nodeGroups:
  - name: ci-workers-spot
    instanceType: c5.4xlarge
    minSize: 0
    maxSize: 20
    spot: true
    labels:
      workload: ci
      lifecycle: spot
```

**GKE (Preemptible VMs)**:
```bash
gcloud container node-pools create ci-workers-preemptible \
  --cluster=c8s-cluster \
  --preemptible \
  --machine-type=n1-standard-8 \
  --enable-autoscaling \
  --min-nodes=0 \
  --max-nodes=20
```

### Retry Strategy for Spot Interruptions

C8S automatically retries failed Jobs, so spot instance interruptions are handled gracefully.

Configure retry policy in PipelineConfig:
```yaml
version: v1alpha1
name: my-pipeline
retryPolicy:
  maxRetries: 3
  backoffSeconds: 30
steps:
  - name: test
    # ...
```

### Scale to Zero

Configure node pool to scale to 0 when idle:
```yaml
minSize: 0
maxSize: 20
```

Benefits:
- Zero cost when no pipelines running
- Ideal for dev/staging environments
- Minimal impact on production (1-2 minute startup delay)

## Troubleshooting

### Pods Stuck Pending

**Symptom**: Jobs created but Pods never run

**Causes**:
1. **ResourceQuota exceeded** - Check quota: `kubectl describe resourcequota -n <namespace>`
2. **Node pool at max size** - Increase max nodes in autoscaler config
3. **Insufficient instance types** - Cloud provider out of capacity for requested instance type
4. **Taints/tolerations mismatch** - Check node taints and Pod tolerations

**Resolution**:
```bash
# Check PipelineRun status
kubectl describe pipelinerun <name>

# Check Job events
kubectl describe job <job-name>

# Check Pod events
kubectl describe pod <pod-name>
```

### Slow Scale-Up

**Symptom**: Takes >5 minutes for new nodes to become available

**Causes**:
1. Cloud provider API latency
2. Large AMI/image size (slow download)
3. Node initialization scripts (user-data)

**Resolution**:
- Use smaller, optimized AMIs
- Pre-pull common container images to AMI
- Minimize node initialization time

### Nodes Not Scaling Down

**Symptom**: Nodes remain after Jobs complete

**Causes**:
1. System pods prevent drain (e.g., DaemonSets)
2. PodDisruptionBudgets blocking eviction
3. Local storage on nodes
4. Scale-down delay not elapsed

**Resolution**:
```bash
# Check Cluster Autoscaler config
kubectl -n kube-system get deployment cluster-autoscaler -o yaml | grep scale-down

# Check node events
kubectl describe node <node-name>

# Manually cordon and drain for testing
kubectl cordon <node-name>
kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data
```

## References

- [Kubernetes Cluster Autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler)
- [AWS EKS Autoscaling](https://docs.aws.amazon.com/eks/latest/userguide/autoscaling.html)
- [GKE Cluster Autoscaler](https://cloud.google.com/kubernetes-engine/docs/concepts/cluster-autoscaler)
- [Azure AKS Autoscaler](https://docs.microsoft.com/en-us/azure/aks/cluster-autoscaler)
