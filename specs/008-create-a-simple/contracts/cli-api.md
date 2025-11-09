# CLI API Contract: c8s-deploy Command

**Feature**: Deploy C8S Stack to Kubernetes
**Date**: 2025-11-09

---

## CLI Command Structure

### Main Command: `c8s-deploy`

Base deployment orchestration tool for C8S.

```bash
c8s-deploy [command] [flags]
```

---

## Commands

### 1. `deploy` - Deploy C8S to Kubernetes

Deploy the entire C8S stack to a Kubernetes cluster.

**Usage**:
```bash
c8s-deploy deploy [flags]
```

**Flags**:
- `--config, -c` (string) - Path to deployment configuration file
- `--namespace` (string) - Kubernetes namespace (default: "c8s-system", overrides config)
- `--version` (string) - C8S version to deploy (default: "latest", overrides config)
- `--environment` (string) - Environment preset: dev, staging, production (default: "dev")
- `--kubeconfig` (string) - Path to kubeconfig file (default: ~/.kube/config)
- `--context` (string) - Kubernetes context to use (default: current context)
- `--timeout` (duration) - Deployment timeout (default: "5m")
- `--wait` (bool) - Wait for all components to be ready (default: true)
- `--skip-validation` (bool) - Skip cluster prerequisite validation (default: false)
- `--dry-run` (bool) - Show what would be deployed without applying (default: false)
- `--verbose, -v` (bool) - Enable verbose logging (default: false)
- `--image-registry` (string) - Container image registry (overrides config)

**Output (Success)**:
```
Validating Kubernetes cluster prerequisites...
✓ Kubernetes 1.24+ detected
✓ RBAC enabled
✓ API server accessible

Loading configuration...
✓ Config file parsed: deployment-config.yaml
✓ 3 components configured

Deploying C8S components...
→ Deploying controller...
✓ Controller deployment created (3/3 replicas ready)
→ Deploying webhook...
✓ Webhook deployment created (1/1 replicas ready)
→ Deploying API server...
✓ API server deployment created (1/1 replicas ready)

Waiting for deployments to be ready... (2 of 3)
⟳ API server: 0/1 ready (pending)
✓ API server: 1/1 ready

✓ All C8S components deployed successfully!

Dashboard available at:
  Port-forward: kubectl port-forward svc/c8s-frontend 8080:80 -n c8s-system
  Then access: http://localhost:8080

Default credentials:
  Username: admin
  Password: [check kubectl secret c8s-admin-creds]

Next steps:
  1. Access dashboard: kubectl port-forward svc/c8s-frontend 8080:80 -n c8s-system
  2. Check logs: kubectl logs -f deployment/c8s-api-server -n c8s-system
  3. Verify health: c8s-deploy health
  4. View configuration: kubectl get configmap c8s-config -o yaml -n c8s-system

Deployment completed in 2m 15s
```

**Output (Failure)**:
```
Validating Kubernetes cluster prerequisites...
✗ Validation failed:
  - Kubernetes version 1.20 detected, minimum 1.24 required
  - RBAC not enabled in cluster

Please ensure:
  1. Upgrade Kubernetes cluster to 1.24+
  2. Enable RBAC on cluster

Deployment failed. 0 components deployed.
```

**Output (Partial)**:
```
Deploying C8S components...
✓ Controller deployment created (3/3 replicas ready)
✓ Webhook deployment created (1/1 replicas ready)
→ Deploying API server...
✗ API server: 0/1 ready (image pull error)

Partial deployment: 2 of 3 components ready

Issues detected:
  - API server: ErrImagePull - Failed to pull image docker.io/org/c8s-api-server:typo-version
    Check image registry and version

Resolution steps:
  1. Fix image tag in configuration: --image-registry or config file
  2. Retry deployment: c8s-deploy deploy

View logs: kubectl logs c8s-api-server-xxx -n c8s-system
```

**Exit Codes**:
- `0` - Deployment successful, all components ready
- `1` - Deployment failed, validation error or critical issue
- `2` - Partial deployment, some components ready
- `3` - Timeout waiting for readiness

---

### 2. `health` - Verify Deployment Health

Check the health and readiness of deployed C8S components.

**Usage**:
```bash
c8s-deploy health [flags]
```

**Flags**:
- `--namespace` (string) - Kubernetes namespace to check (default: "c8s-system")
- `--kubeconfig` (string) - Path to kubeconfig file (default: ~/.kube/config)
- `--context` (string) - Kubernetes context to use (default: current context)
- `--format` (string) - Output format: text, json, yaml (default: "text")
- `--verbose, -v` (bool) - Enable verbose output with timestamps
- `--output, -o` (string) - Write output to file instead of stdout

**Output (Healthy)**:
```
Health Check Results
====================

Overall Status: ✓ HEALTHY
Timestamp: 2025-11-09T14:23:45Z
Namespace: c8s-system
Kubernetes Version: 1.25.3
Nodes: 3

Components
----------
✓ c8s-controller
  Replicas: 3/3 ready
  Status: Running
  Last transition: 2m ago
  Liveness: OK (response: 45ms)
  Readiness: OK (response: 12ms)

✓ c8s-webhook
  Replicas: 1/1 ready
  Status: Running
  Last transition: 2m ago
  Liveness: OK (response: 8ms)
  Readiness: OK (response: 3ms)

✓ c8s-api-server
  Replicas: 1/1 ready
  Status: Running
  Last transition: 1m ago
  Liveness: OK (response: 15ms)
  Readiness: OK (response: 5ms)

Dependencies
-----------
✓ Storage: S3 endpoint accessible
✓ Database: PostgreSQL reachable
✓ RBAC: Service accounts configured

Dashboard
---------
Access via: kubectl port-forward svc/c8s-frontend 8080:80 -n c8s-system
URL: http://localhost:8080
```

**Output (Degraded)**:
```
Health Check Results
====================

Overall Status: ⚠ DEGRADED
Timestamp: 2025-11-09T14:23:45Z

Components
----------
✓ c8s-controller
  Replicas: 3/3 ready
  Status: Running

⚠ c8s-webhook
  Replicas: 1/2 ready
  Status: Pending (waiting for node resources)
  Last transition: 5m ago
  Issue: Insufficient CPU resources on nodes

✓ c8s-api-server
  Replicas: 1/1 ready
  Status: Running

Issues & Remediation
--------------------
1. Webhook pod pending (1/2 replicas)
   Cause: Insufficient CPU resources
   Fix: Add more nodes or increase node capacity

View pod details:
  kubectl describe pod -l app=c8s-webhook -n c8s-system
  kubectl top nodes
  kubectl logs c8s-webhook-xxx -n c8s-system
```

**Output (Unhealthy)**:
```
Health Check Results
====================

Overall Status: ✗ UNHEALTHY
Timestamp: 2025-11-09T14:23:45Z

Components
----------
✗ c8s-controller
  Replicas: 0/3 ready
  Status: CrashLoopBackOff
  Issue: Application crashing on startup

✗ c8s-api-server
  Replicas: 0/1 ready
  Status: ImagePullBackOff
  Issue: Image not found in registry

Critical Issues
---------------
1. Controller pods crashing
   Last log: panic: failed to connect to database

2. API server unable to pull image
   Image: docker.io/org/c8s-api-server:v99.0.0
   Error: manifest not found

Remediation Steps
-----------------
1. Check database connectivity:
   kubectl exec -it deployment/c8s-controller -- mysql -h db -e "SELECT 1"

2. Verify image exists:
   docker pull docker.io/org/c8s-api-server:v99.0.0
   Or use correct version: c8s-deploy deploy --version v0.2.0

3. View full logs:
   kubectl logs deployment/c8s-controller -n c8s-system --all-containers
```

**Exit Codes**:
- `0` - All components healthy
- `1` - One or more components unhealthy
- `2` - Partial health (some components running)

---

### 3. `config` - Manage Deployment Configuration

Manage C8S deployment configuration.

**Usage**:
```bash
c8s-deploy config [subcommand] [flags]
```

**Subcommands**:

#### `config init`
Initialize a new deployment configuration file.

```bash
c8s-deploy config init [flags]
```

**Flags**:
- `--name` (string) - Configuration name (default: "c8s-deployment")
- `--environment` (string) - Environment type: dev, staging, production (default: "dev")
- `--namespace` (string) - Kubernetes namespace (default: "c8s-system")
- `--output, -o` (string) - Output configuration file path (default: "c8s-config.yaml")

**Output**:
```
Creating new C8S deployment configuration...
✓ Configuration template created: c8s-config.yaml

Next steps:
1. Review and edit configuration: c8s-config.yaml
2. Customize components, storage, and resources as needed
3. Deploy: c8s-deploy deploy --config c8s-config.yaml

For help on configuration options: c8s-deploy config help
```

#### `config validate`
Validate a configuration file.

```bash
c8s-deploy config validate --config c8s-config.yaml
```

**Output (Valid)**:
```
Validating configuration file: c8s-config.yaml
✓ Configuration is valid
  - Deployment name: production-c8s
  - Environment: production
  - Namespace: c8s-system
  - Components: 3 (controller, webhook, api-server)
  - Storage: S3-compatible (minio.local:9000)
  - Resource limits: configured
```

**Output (Invalid)**:
```
Validating configuration file: c8s-config.yaml
✗ Configuration validation failed:

Errors:
1. components.controller.image.tag must not be empty
2. environment.type must be one of: dev, staging, production
   Current value: "development"
3. storage.endpoint must be valid hostname:port
   Current value: "minio"

Fix these errors and retry: c8s-deploy config validate --config c8s-config.yaml
```

#### `config set`
Modify configuration values.

```bash
c8s-deploy config set KEY VALUE [flags]
```

**Examples**:
```bash
c8s-deploy config set environment.type production
c8s-deploy config set components.controller.replicas 3
c8s-deploy config set storage.endpoint s3.amazonaws.com
c8s-deploy config set --config staging-config.yaml components.apiServer.image.tag v0.2.0
```

#### `config get`
Display configuration values.

```bash
c8s-deploy config get [KEY] [flags]
```

**Examples**:
```bash
c8s-deploy config get                              # Show entire config
c8s-deploy config get environment.type
c8s-deploy config get components.controller      # Show component config
```

---

### 4. `upgrade` - Upgrade C8S Version

Upgrade deployed C8S to a new version.

**Usage**:
```bash
c8s-deploy upgrade [flags]
```

**Flags**:
- `--to-version` (string) - Target version to upgrade to (required, e.g., "v0.2.0")
- `--namespace` (string) - Kubernetes namespace (default: "c8s-system")
- `--dry-run` (bool) - Show what would be upgraded without making changes
- `--timeout` (duration) - Upgrade timeout (default: "10m")
- `--preserve-config` (bool) - Keep current configuration (default: true)

**Output (Success)**:
```
Preparing C8S upgrade...
Current version: v0.1.0
Target version: v0.2.0
Namespace: c8s-system

Pre-upgrade checks...
✓ Current deployment healthy
✓ All components ready
✓ Backup created: c8s-backup-20251109-142345

Upgrading components...
→ Updating controller (v0.1.0 → v0.2.0)...
✓ Controller rolling update complete (3/3 replicas updated)

→ Updating webhook (v0.1.0 → v0.2.0)...
✓ Webhook rolling update complete (1/1 replicas updated)

→ Updating API server (v0.1.0 → v0.2.0)...
✓ API server rolling update complete (1/1 replicas updated)

Post-upgrade verification...
✓ All components healthy after upgrade
✓ Readiness probes passing

Upgrade completed successfully!
Time taken: 3m 45s
Backup: c8s-backup-20251109-142345 (kept for 7 days)

To rollback if needed:
  c8s-deploy rollback --backup c8s-backup-20251109-142345
```

**Exit Codes**:
- `0` - Upgrade successful
- `1` - Upgrade failed, rollback performed
- `2` - Pre-upgrade checks failed, no upgrade attempted

---

### 5. `uninstall` - Remove C8S Deployment

Cleanly remove C8S from Kubernetes cluster.

**Usage**:
```bash
c8s-deploy uninstall [flags]
```

**Flags**:
- `--namespace` (string) - Kubernetes namespace (default: "c8s-system")
- `--keep-namespace` (bool) - Keep namespace after uninstall (default: false)
- `--keep-data` (bool) - Keep persistent volumes with data (default: false)
- `--force` (bool) - Force removal without confirmation (default: false)
- `--timeout` (duration) - Uninstall timeout (default: "5m")

**Output (Interactive)**:
```
C8S Uninstall Wizard
====================

This will remove:
✓ All C8S deployments (controller, webhook, api-server)
✓ C8S services and configmaps
✓ C8S RBAC resources (service accounts, roles)

This will NOT remove:
✓ Namespace (will keep c8s-system for future use)
✓ Persistent data (volumes and data retained)
✓ Kubernetes cluster itself

Confirm uninstall? (type 'yes' to confirm): yes

Removing C8S resources...
→ Removing deployments...
✓ Deployment c8s-controller deleted
✓ Deployment c8s-webhook deleted
✓ Deployment c8s-api-server deleted

→ Removing services...
✓ Service c8s-frontend deleted
✓ Service c8s-api deleted

→ Removing RBAC resources...
✓ ServiceAccount c8s-controller deleted
✓ ServiceAccount c8s-webhook deleted
✓ ClusterRole c8s-controller deleted
✓ ClusterRoleBinding c8s-controller deleted

→ Removing configmaps and secrets...
✓ ConfigMap c8s-config deleted

✓ C8S uninstalled successfully!

Namespace: c8s-system (kept for future use)
Data: 2 persistent volumes retained (10GB total)

To remove namespace: kubectl delete namespace c8s-system
To remove data: kubectl delete pvc --all -n c8s-system
```

**Exit Codes**:
- `0` - Uninstall successful
- `1` - Uninstall failed, some resources remain
- `2` - User cancelled uninstall

---

## Global Flags

Available for all commands:

- `--help, -h` - Show command help
- `--version` - Show c8s-deploy version
- `--debug` - Enable debug logging
- `--no-color` - Disable colored output
- `--quiet` - Suppress non-error output

---

## Configuration File Format

See `contracts/config-schema.json` for the complete JSON schema.

Example configuration file (YAML format):
```yaml
metadata:
  name: production-c8s
  version: v0.2.0

cluster:
  namespace: c8s-system
  kubeconfig: ~/.kube/config
  context: gke-prod

environment:
  type: production
  logLevel: info

components:
  controller:
    replicas: 3
    image:
      registry: docker.io
      tag: v0.2.0
  webhook:
    replicas: 2
  apiServer:
    replicas: 2

storage:
  type: s3-compatible
  endpoint: s3.amazonaws.com
  bucket: prod-c8s-logs
  region: us-west-2
  useSSL: true

resources:
  requests:
    cpu: 500m
    memory: 512Mi
  limits:
    cpu: 2000m
    memory: 2Gi
```

---

## Error Codes Reference

| Code | Meaning | Resolution |
|------|---------|-----------|
| `VALIDATION_FAILED` | Cluster validation failed | Check Kubernetes version, RBAC, API access |
| `CONFIG_INVALID` | Configuration file invalid | Validate with `config validate` |
| `DEPLOYMENT_TIMEOUT` | Deployment took too long | Increase `--timeout`, check cluster resources |
| `IMAGE_PULL_ERROR` | Failed to pull container image | Check image registry, tag, credentials |
| `INSUFFICIENT_RESOURCES` | Not enough CPU/memory on nodes | Add nodes or reduce resource requests |
| `RBAC_DENIED` | Permission denied on cluster | Verify RBAC permissions |
| `NAMESPACE_EXISTS` | Namespace already exists | Use existing namespace or choose different name |
| `CRD_NOT_FOUND` | Required CRD missing | Ensure C8S CRDs are installed |
| `STORAGE_ERROR` | Storage configuration error | Check S3 endpoint, bucket, credentials |

---

## Examples

Deploy to local cluster:
```bash
c8s-deploy deploy --environment dev --namespace c8s-local
```

Deploy with custom configuration:
```bash
c8s-deploy deploy --config my-config.yaml --timeout 10m
```

Check health after deployment:
```bash
c8s-deploy health --format json > health-report.json
```

Initialize production configuration:
```bash
c8s-deploy config init --environment production --output prod-config.yaml
```

Upgrade to new version:
```bash
c8s-deploy upgrade --to-version v0.2.0 --preserve-config
```

Uninstall cleanly:
```bash
c8s-deploy uninstall --keep-data --keep-namespace
```
