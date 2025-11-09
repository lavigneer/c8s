# C8S Helm Chart - Troubleshooting Guide

Common issues and solutions for deploying C8S with Helm.

## Helm Installation Issues

### Error: "Chart.yaml: icon is recommended"

This is just a warning (INFO level), not an error. Your chart is valid.

**Fix**: To make it go away, add an icon to Chart.yaml:
```yaml
icon: https://example.com/icon.png
```

### Error: "chart requires kubeVersion..."

Your Kubernetes cluster version doesn't meet the minimum requirements.

**Solution**:
```bash
# Check your cluster version
kubectl version --short

# C8S requires Kubernetes 1.24+
# Upgrade your cluster or adjust Chart.yaml requirements
```

### Error: "failed to download dependency..."

The chart couldn't download a dependency.

**Solution**:
```bash
# Update Helm repositories
helm repo update

# Try installing again
helm install c8s ./chart/c8s
```

---

## Deployment Issues

### Pods stuck in Pending state

The pods can't be scheduled on available nodes.

**Diagnosis**:
```bash
# Check pod status
kubectl get pods -n c8s-system

# Get detailed information
kubectl describe pod <pod-name> -n c8s-system

# Check node resources
kubectl top nodes
kubectl describe nodes
```

**Common causes and solutions**:

1. **Insufficient resources**: Cluster doesn't have enough CPU/memory
   ```bash
   # Reduce resource requests
   helm upgrade c8s ./chart/c8s \
     -f values-dev.yaml \
     --set components.controller.resources.requests.cpu=50m \
     --set components.controller.resources.requests.memory=128Mi
   ```

2. **No nodes available**: Cluster has no worker nodes
   ```bash
   kubectl get nodes
   # Should show at least one worker node in Ready state
   ```

3. **Taints/tolerations**: Nodes have taints but pods don't tolerate them
   ```bash
   kubectl describe nodes | grep Taints
   # If taints exist, add tolerations to your values file
   ```

### ImagePullBackOff errors

Docker image can't be pulled from the registry.

**Diagnosis**:
```bash
# Check the error message
kubectl describe pod <pod-name> -n c8s-system

# Test image availability
docker pull <image-name>:<tag>
```

**Solutions**:

1. **Wrong image tag**:
   ```bash
   helm upgrade c8s ./chart/c8s \
     --set components.controller.image.tag=v0.1.0
   ```

2. **Wrong registry**:
   ```bash
   helm upgrade c8s ./chart/c8s \
     --set components.controller.image.registry=gcr.io
   ```

3. **Private registry authentication**:
   ```bash
   # Create image pull secret
   kubectl create secret docker-registry regcred \
     --docker-server=<registry> \
     --docker-username=<username> \
     --docker-password=<password> \
     -n c8s-system
   
   # Update values to use the secret
   # Add imagePullSecrets to deployment templates
   ```

### CrashLoopBackOff or ExitCode errors

The pod is running but crashing or exiting.

**Diagnosis**:
```bash
# Check pod logs
kubectl logs <pod-name> -n c8s-system --tail=50

# Check for previous crash logs
kubectl logs <pod-name> -n c8s-system --previous

# Get detailed pod events
kubectl describe pod <pod-name> -n c8s-system
```

**Common causes**:
- Configuration errors (LOG_LEVEL invalid, missing environment variables)
- Port conflicts (port already in use)
- Missing dependencies (database, cache, etc.)
- Application bugs

---

## Health Check Issues

### Post-install hook fails

The health check didn't complete successfully.

**Diagnosis**:
```bash
# Check if hook pod exists
kubectl get pods -n c8s-system -l app.kubernetes.io/name=c8s

# Check hook pod logs
kubectl logs -n c8s-system -l job-name=c8s-post-install-hook

# Check if deployments are ready
kubectl get deployment -n c8s-system
```

**Solutions**:

1. **Timeout too short**: Increase the timeout
   ```bash
   helm upgrade c8s ./chart/c8s \
     --set postInstallHook.timeout=600  # 10 minutes instead of 5
   ```

2. **Component not ready**: Check individual component logs
   ```bash
   kubectl logs deployment/c8s-controller -n c8s-system --tail=50
   kubectl logs deployment/c8s-webhook -n c8s-system --tail=50
   ```

### Dashboard not accessible after deployment

The frontend service is created but not accessible.

**Diagnosis**:
```bash
# Check service
kubectl get svc -n c8s-system

# Check if LoadBalancer has external IP (cloud) or NodePort (local)
kubectl get svc c8s-frontend -n c8s-system -o wide

# Test connectivity with port-forward
kubectl port-forward svc/c8s-frontend -n c8s-system 3000:80
# Then visit http://localhost:3000
```

**Solutions**:

1. **LoadBalancer pending (cloud)**: Wait for external IP
   ```bash
   kubectl get svc c8s-frontend -n c8s-system --watch
   # May take 1-5 minutes to get external IP
   ```

2. **Using port-forward locally**:
   ```bash
   kubectl port-forward svc/c8s-frontend -n c8s-system 3000:80
   # Access at http://localhost:3000
   ```

3. **Check service endpoints**:
   ```bash
   kubectl get endpoints c8s-frontend -n c8s-system
   # Should show pod IPs
   ```

---

## Storage Issues

### S3 credentials not working

S3 storage configuration fails due to invalid credentials.

**Diagnosis**:
```bash
# Check if S3 secret was created
kubectl get secrets -n c8s-system | grep s3

# Check secret contents (base64 encoded)
kubectl get secret c8s-s3-secret -n c8s-system -o yaml

# Check if environment variables are set in pods
kubectl exec -it <pod-name> -n c8s-system -- env | grep AWS
```

**Solutions**:

1. **Update S3 credentials**:
   ```bash
   helm upgrade c8s ./chart/c8s \
     --set storage.type=s3-compatible \
     --set storage.s3.enabled=true \
     --set storage.s3.endpoint=s3.amazonaws.com \
     --set storage.s3.accessKey=<new-key> \
     --set storage.s3.secretKey=<new-secret>
   ```

2. **Test S3 connectivity**:
   ```bash
   # From inside a pod
   kubectl exec -it <pod-name> -n c8s-system -- sh
   
   # Test S3 endpoint
   curl -I https://s3.amazonaws.com
   ```

### PVC not binding

PersistentVolumeClaim can't find a PersistentVolume.

**Diagnosis**:
```bash
# Check PVC status
kubectl get pvc -n c8s-system

# Check PV availability
kubectl get pv

# Describe PVC for events
kubectl describe pvc -n c8s-system
```

**Solutions**:

1. **No storage class available**:
   ```bash
   # List available storage classes
   kubectl get storageclass
   
   # Use specific storage class
   helm upgrade c8s ./chart/c8s \
     --set storage.pvc.storageClass=fast
   ```

2. **Insufficient storage capacity**:
   ```bash
   # Reduce PVC size
   helm upgrade c8s ./chart/c8s \
     --set storage.pvc.size=5Gi
   ```

---

## Tilt Issues

### Tiltfile Error: "module has no .analysis field"

Invalid Tilt syntax in Tiltfile.

**Fix**: Update Tiltfile to use correct syntax:
```python
# Load Helm resource extension
load('ext://helm_resource', 'helm_resource')

# Deploy using Helm
helm_resource(
  name='c8s',
  chart_dir='./chart/c8s',
  flags=['-f', './chart/c8s/values-dev.yaml'],
  namespace='c8s-system'
)
```

### Tilt can't connect to Kubernetes cluster

Tilt can't reach your local Kubernetes cluster.

**Diagnosis**:
```bash
# Check kubectl can access cluster
kubectl cluster-info

# Check current context
kubectl config current-context

# List available contexts
kubectl config get-contexts
```

**Solutions**:

1. **Switch to correct Kubernetes context**:
   ```bash
   kubectl config use-context docker-desktop
   # or
   kubectl config use-context minikube
   # or
   kubectl config use-context kind-kind
   ```

2. **Update Tiltfile to allow your context**:
   ```python
   allow_k8s_context('docker-desktop')
   allow_k8s_context('minikube')
   allow_k8s_context('kind-kind')
   ```

---

## Configuration Issues

### Values override not working

CLI --set flag doesn't override values file settings.

**Solution**: Check the order of precedence:
```bash
# Order matters: later values override earlier ones
helm install c8s ./chart/c8s \
  -f values.yaml \                    # First
  -f values-prod.yaml \                # Overrides above
  --set components.controller.replicas=5  # Overrides everything
```

### Invalid parameter name error

Parameter name doesn't exist in values file.

**Diagnosis**:
```bash
# Check available parameters
helm values ./chart/c8s | less

# Validate values file
helm template c8s ./chart/c8s -f custom.yaml --debug
```

**Solution**:
```bash
# Use correct parameter path
helm install c8s ./chart/c8s \
  --set components.controller.replicas=3  # Correct
  # NOT --set controller.replicas=3
```

---

## RBAC / Permissions Issues

### Webhook can't validate CRDs

ValidatingWebhookConfiguration can't reach webhook service.

**Diagnosis**:
```bash
# Check webhook service exists
kubectl get svc c8s-webhook -n c8s-system

# Check webhook pod is running
kubectl get pods -n c8s-system -l app.kubernetes.io/component=webhook

# Check webhook logs
kubectl logs deployment/c8s-webhook -n c8s-system
```

**Solutions**:

1. **Enable webhook in values**:
   ```bash
   helm upgrade c8s ./chart/c8s \
     --set components.webhook.enabled=true
   ```

2. **Check webhook service has endpoints**:
   ```bash
   kubectl get endpoints c8s-webhook -n c8s-system
   ```

### Controller can't watch CRDs

Controller needs ClusterRole permissions.

**Diagnosis**:
```bash
# Check ClusterRole exists
kubectl get clusterrole c8s-controller

# Check ClusterRoleBinding exists
kubectl get clusterrolebinding c8s-controller

# Check service account
kubectl get serviceaccount c8s-controller -n c8s-system
```

**Solutions**:

1. **Recreate RBAC**:
   ```bash
   # Delete and reinstall
   helm uninstall c8s -n c8s-system
   helm install c8s ./chart/c8s \
     --set rbac.create=true \
     -n c8s-system --create-namespace
   ```

2. **Add missing permissions**: Edit rbac.yaml to add CRD rules

---

## General Debugging

### Enable debug logging

Get more detailed output for troubleshooting.

```bash
# Helm debug output
helm install c8s ./chart/c8s --debug --dry-run

# Kubectl verbose output
kubectl get pods -n c8s-system -v=6

# Check API server logs (if self-hosted)
kubectl logs -n kube-system -l component=kube-apiserver
```

### Collect diagnostics

Gather information for support.

```bash
# Export all resources
kubectl get all -n c8s-system -o yaml > c8s-resources.yaml

# Export configuration
kubectl get cm,secret -n c8s-system -o yaml > c8s-config.yaml

# Export pod descriptions
kubectl describe pods -n c8s-system > c8s-pods.txt

# Export events
kubectl get events -n c8s-system > c8s-events.txt
```

---

## Getting Help

If you're stuck:

1. **Check logs**: `kubectl logs <pod> -n c8s-system`
2. **Describe resources**: `kubectl describe pod <pod> -n c8s-system`
3. **Check events**: `kubectl get events -n c8s-system`
4. **Check GitHub issues**: https://github.com/lavigneer/c8s/issues
5. **Consult Kubernetes docs**: https://kubernetes.io/docs/

---

## Related Documentation

- [C8S Helm Chart README](../chart/c8s/README.md)
- [Kubernetes Troubleshooting Guide](https://kubernetes.io/docs/tasks/debug-application-cluster/)
- [Helm Documentation](https://helm.sh/docs/)
