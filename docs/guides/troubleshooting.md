# C8S Troubleshooting Guide

This guide helps you diagnose and resolve common C8S issues.

## Table of Contents

1. [Common Error Messages](#common-error-messages)
2. [Diagnostic Steps](#diagnostic-steps)
3. [Debug Commands](#debug-commands)
4. [Logs Interpretation](#logs-interpretation)
5. [Performance Issues](#performance-issues)
6. [Network Issues](#network-issues)
7. [Storage Issues](#storage-issues)
8. [Security Issues](#security-issues)

---

## Common Error Messages

### "Webhook Signature Verification Failed"

**Error Message**:
```
Webhook signature verification failed
```

**Causes**:
- Git platform webhook secret doesn't match K8s Secret
- Secret was updated but webhook config wasn't
- Webhook payload was modified in transit

**Solutions**:

1. **Verify secret matches**:
```bash
# Check secret stored in K8s
kubectl get secret webhook-secret -o jsonpath='{.data.webhook-secret}' | base64 -d

# Verify it matches your Git platform webhook settings
# Go to Settings → Webhooks → Edit → Secret
```

2. **Recreate the secret**:
```bash
# Delete old secret
kubectl delete secret webhook-secret

# Create new secret with correct value
kubectl create secret generic webhook-secret \
  --from-literal=webhook-secret='your-secret-key'
```

3. **Check webhook logs**:
```bash
kubectl logs deployment/c8s-webhook -n c8s-system -f
grep "signature" | head -20
```

### "PipelineConfig Not Found"

**Error Message**:
```
No RepositoryConnection found for repository
```

**Causes**:
- PipelineConfig doesn't exist in K8s
- Wrong namespace specified
- Repository URL doesn't match exactly

**Solutions**:

1. **List available configs**:
```bash
kubectl get pipelineconfig -A
kubectl get pipelineconfig -n <namespace>
```

2. **Check config details**:
```bash
kubectl describe pipelineconfig <name> -n <namespace>
```

3. **Verify repository URL**:
```bash
# Config should have exact URL match
kubectl get pipelineconfig -o yaml | grep -A2 "repository:"
```

### "Resource Quota Exceeded"

**Error Message**:
```
Pod failed to create: Forbidden: quotas exceeded
```

**Causes**:
- Namespace has resource quotas
- Pipeline steps request more resources than available
- Multiple concurrent pipelines exhausting quota

**Solutions**:

1. **Check current quota**:
```bash
kubectl describe resourcequota -n <namespace>
kubectl top pods -n <namespace>
```

2. **Adjust step resources**:
```yaml
steps:
  - name: my-step
    image: my-image
    resources:
      requests:
        cpu: 100m      # Reduce from default
        memory: 256Mi   # Reduce from default
      limits:
        cpu: 500m
        memory: 512Mi
```

3. **Increase quota** (if you're the admin):
```bash
kubectl patch resourcequota <quota-name> -p '{"spec":{"hard":{"requests.cpu":"100"}}}'
```

### "PipelineRun Stuck in Running"

**Error Message**:
```
Status: Running (for > 1 hour)
```

**Causes**:
- Job pod crashed but status not updated
- Controller restarted/redeployed
- Network issue preventing status updates
- Resource constraints preventing pod creation

**Solutions**:

1. **Check created Jobs**:
```bash
kubectl get jobs -n c8s-system | grep <run-name>
kubectl describe job <job-name> -n c8s-system
```

2. **Check pod status**:
```bash
kubectl get pods -n c8s-system | grep <run-name>
kubectl logs <pod-name> -n c8s-system
```

3. **Manually update status** (debugging only):
```bash
kubectl patch pipelinerun <run-name> \
  -p '{"status":{"phase":"Failed","reason":"Timeout"}}'
```

### "Log Retrieval Failed"

**Error Message**:
```
Failed to stream logs: Connection refused
```

**Causes**:
- Object storage not accessible
- S3 credentials wrong
- Network connectivity issue
- Service not responding

**Solutions**:

1. **Check S3 connection**:
```bash
# From API server pod
kubectl exec -it deployment/c8s-api-server -c api-server -- sh
aws s3 ls s3://c8s-logs/
```

2. **Verify S3 credentials**:
```bash
kubectl get secret s3-credentials -o yaml
# Verify: access-key, secret-key, endpoint, bucket
```

3. **Check service health**:
```bash
kubectl port-forward svc/c8s-api-server 8080:8080
curl http://localhost:8080/health
curl http://localhost:8080/api/health
```

---

## Diagnostic Steps

### Step 1: Check Installation

```bash
# Verify CRDs are installed
kubectl api-resources | grep c8s

# Should show:
# pipelineconfigs        c8s.io    true        PipelineConfig
# pipelineruns           c8s.io    true        PipelineRun
```

### Step 2: Verify Components

```bash
# Check all C8S pods
kubectl get pods -n c8s-system

# Expected:
# c8s-controller-xxxxx    1/1     Running
# c8s-api-server-xxxxx    1/1     Running
# c8s-webhook-xxxxx       1/1     Running
```

### Step 3: Check Resource Status

```bash
# List all pipelines
kubectl get pipelineconfig -A

# List all runs
kubectl get pipelinerun -A

# Check specific run
kubectl describe pipelinerun <name> -n <namespace>
```

### Step 4: Review Logs

```bash
# Controller logs
kubectl logs deployment/c8s-controller -n c8s-system --tail=100

# API server logs
kubectl logs deployment/c8s-api-server -n c8s-system --tail=100

# Webhook logs
kubectl logs deployment/c8s-webhook -n c8s-system --tail=100
```

### Step 5: Gather Debug Info

```bash
# Create debug bundle
mkdir debug-bundle
kubectl get all -n c8s-system > debug-bundle/resources.yaml
kubectl logs deployment/c8s-controller -n c8s-system > debug-bundle/controller.log
kubectl logs deployment/c8s-api-server -n c8s-system > debug-bundle/api-server.log
kubectl logs deployment/c8s-webhook -n c8s-system > debug-bundle/webhook.log
kubectl describe nodes > debug-bundle/nodes.yaml

# Share debug-bundle/ with support
```

---

## Debug Commands

### View Component Logs

```bash
# Real-time logs
kubectl logs -f deployment/c8s-controller -n c8s-system

# Last 100 lines
kubectl logs deployment/c8s-controller -n c8s-system --tail=100

# Last 10 minutes
kubectl logs deployment/c8s-controller -n c8s-system --since=10m

# Previous pod (if crashed)
kubectl logs deployment/c8s-controller -n c8s-system --previous
```

### Inspect Resources

```bash
# Full YAML of resource
kubectl get pipelinerun <name> -o yaml

# Watch real-time updates
kubectl get pipelinerun -w

# Describe detailed status
kubectl describe pipelinerun <name>

# Show only specific fields
kubectl get pipelinerun <name> -o jsonpath='{.status.phase}'
```

### Execute Commands in Pods

```bash
# Interactive shell
kubectl exec -it deployment/c8s-api-server -n c8s-system -- /bin/sh

# Run single command
kubectl exec deployment/c8s-api-server -n c8s-system -- curl http://localhost:8080/health

# Run command in specific pod
kubectl exec -it <pod-name> -n c8s-system -- bash
```

### Port Forwarding

```bash
# Forward API server
kubectl port-forward svc/c8s-api-server 8080:8080 -n c8s-system

# Forward specific pod
kubectl port-forward pod/<pod-name> 8080:8080 -n c8s-system

# Access: http://localhost:8080
```

### Network Testing

```bash
# Test DNS resolution
kubectl exec <pod> -n c8s-system -- nslookup c8s-api-server

# Test connectivity
kubectl exec <pod> -n c8s-system -- curl http://c8s-api-server:8080/health

# Check network policy
kubectl get networkpolicy -n c8s-system
```

---

## Logs Interpretation

### Controller Logs

**Normal startup**:
```
INFO: Starting C8S controller
INFO: Registering webhooks
INFO: Controller ready
```

**Error to investigate**:
```
ERROR: Failed to update PipelineRun status: context deadline exceeded
ERROR: Failed to create Job: Forbidden
ERROR: Webhook failed: connection refused
```

### API Server Logs

**Normal operation**:
```
INFO: API server listening on :8080
INFO: GET /api/health - 200 OK
INFO: GET /api/projects - 200 OK (user: john)
```

**Issues to fix**:
```
ERROR: Failed to list PipelineRuns: permission denied
ERROR: S3 access failed: InvalidAccessKeyId
ERROR: JWT validation failed: token expired
```

### Webhook Logs

**Successful webhook**:
```
INFO: Received GitHub webhook for repo:owner/repo
INFO: Signature verified successfully
INFO: Created PipelineRun: repo-push-abc123
```

**Failed webhook**:
```
ERROR: Signature verification failed
ERROR: Repository not found
ERROR: Failed to create PipelineRun: [error details]
```

---

## Performance Issues

### Slow Dashboard

**Symptoms**: Dashboard takes >5 seconds to load

**Diagnosis**:
```bash
# Check API response time
time curl http://localhost:8080/api/projects

# Check API server CPU/memory
kubectl top pods -n c8s-system | grep api-server

# Check database/storage latency
kubectl exec -it <api-pod> -- curl -w "@curl-format.txt" http://localhost:8080/api/health
```

**Solutions**:
- Increase API server resources
- Add caching layer
- Optimize database queries
- Scale replicas

### Slow Pipeline Execution

**Symptoms**: Pipelines take longer than expected

**Diagnosis**:
```bash
# Check Job creation delay
kubectl describe pipelinerun <name> | grep "Conditions:"

# Check Pod scheduling delay
kubectl describe job <job-name> -n c8s-system

# Check worker node resources
kubectl top nodes
```

**Solutions**:
- Increase node resources
- Use node affinity for scheduling
- Optimize step dependencies
- Reduce resource requests if realistic

### High Resource Usage

**Symptoms**: Pods using excessive CPU/memory

**Diagnosis**:
```bash
# Top CPU users
kubectl top pods -n c8s-system --sort-by=cpu

# Top memory users
kubectl top pods -n c8s-system --sort-by=memory

# Sustained metrics
watch kubectl top pods -n c8s-system
```

**Solutions**:
- Add resource limits
- Reduce concurrent pipeline runs
- Optimize workload
- Scale to more nodes

---

## Network Issues

### Cannot Connect to API

**Symptom**: `Connection refused`

**Diagnosis**:
```bash
# Verify service exists
kubectl get svc c8s-api-server -n c8s-system

# Test connectivity from within cluster
kubectl run -it --rm debug --image=alpine --restart=Never -- \
  wget -O- http://c8s-api-server:8080/health

# Check firewall/network policies
kubectl get networkpolicy -n c8s-system
```

**Solutions**:
- Verify service is running: `kubectl get pods -n c8s-system`
- Check port-forward: `kubectl port-forward svc/c8s-api-server 8080:8080`
- Verify no network policies blocking traffic

### Webhook Not Accessible

**Symptom**: Git platform can't reach webhook

**Diagnosis**:
```bash
# Check webhook service
kubectl get svc c8s-webhook -n c8s-system

# Verify it's exposed (LoadBalancer/NodePort/Ingress)
kubectl get ingress -n c8s-system

# Test from outside cluster
curl -v http://<webhook-url>/webhooks/github
```

**Solutions**:
- Create Ingress for webhook
- Use NodePort instead of ClusterIP
- Check firewall rules
- Verify DNS resolution

---

## Storage Issues

### Missing Logs

**Symptom**: Logs not available after 5 minutes

**Diagnosis**:
```bash
# Check S3 bucket
aws s3 ls s3://c8s-logs/

# Check API server can reach S3
kubectl exec -it <api-pod> -- aws s3 ls

# Check PipelineRun log reference
kubectl get pipelinerun <name> -o yaml | grep -i "log"
```

**Solutions**:
- Verify S3 credentials in secret
- Check S3 bucket permissions
- Verify S3 endpoint configuration
- Check disk space on nodes

### Artifact Not Found

**Symptom**: `404 Not Found` when downloading artifacts

**Diagnosis**:
```bash
# Check if artifact was created
kubectl get pipelinerun <name> -o yaml | grep -i "artifact"

# List all artifacts
aws s3 ls s3://c8s-artifacts/ --recursive

# Check artifact path
kubectl exec -it <step-pod> -- ls -la /tmp/artifacts
```

**Solutions**:
- Ensure artifact path is correctly mounted
- Check step actually creates artifacts
- Verify S3 upload succeeds (check logs)
- Check artifact retention policy

---

## Security Issues

### Unauthorized Access

**Symptom**: `401 Unauthorized`

**Diagnosis**:
```bash
# Check JWT validation
kubectl logs deployment/c8s-api-server -n c8s-system | grep -i "jwt\|auth"

# Verify secret key
kubectl get secret jwt-secret -o yaml

# Check token expiration
jwt decode <token>  # Use jwt-cli tool
```

**Solutions**:
- Verify JWT secret is consistent
- Check token hasn't expired
- Ensure Bearer token in Authorization header
- Check token signing method matches config

### Permission Denied

**Symptom**: `403 Forbidden`

**Diagnosis**:
```bash
# Check RBAC rules
kubectl get rolebindings -n c8s-system
kubectl get clusterrolebindings | grep c8s

# Check user's roles
kubectl get rolebinding -n <namespace> -o yaml | grep <user>

# Check project access
kubectl get <resource> -n <namespace> -o yaml | grep "accessControl"
```

**Solutions**:
- Verify user has proper role binding
- Check project access permissions
- Verify K8s RBAC allows operation
- Check field-level access control rules

### Secret Not Accessible

**Symptom**: "webhook secret key not found"

**Diagnosis**:
```bash
# List all secrets
kubectl get secrets -n c8s-system

# Check secret structure
kubectl get secret webhook-secret -o yaml

# Verify key exists
kubectl get secret webhook-secret -o jsonpath='{.data.webhook-secret}'
```

**Solutions**:
- Recreate secret with correct key name
- Verify key is base64 encoded
- Check secret isn't corrupted
- Verify mounting in pod

---

## Getting Help

If you can't resolve the issue:

1. **Gather debug info** (see "Diagnostic Steps")
2. **Check logs** (see "Logs Interpretation")
3. **Search documentation** (see specific guides)
4. **Create GitHub issue** with:
   - Error message
   - Debug bundle
   - Steps to reproduce
   - Expected vs actual behavior

---

## Quick Reference

| Issue | Diagnosis | Fix |
|-------|-----------|-----|
| Pod not starting | `kubectl logs pod` | Check image/resources |
| Webhook failing | `kubectl logs c8s-webhook` | Verify secret/URL |
| Slow pipelines | `kubectl top pods` | Increase resources |
| High memory | `watch kubectl top` | Add memory limit |
| No logs | Check S3 | Verify credentials |
| Access denied | Check RBAC | Add role binding |

