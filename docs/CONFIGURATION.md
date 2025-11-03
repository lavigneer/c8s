# C8S Configuration Reference

**Version**: 1.0
**Last Updated**: 2025-11-02

Complete guide to configuring C8S components for your deployment environment.

## Table of Contents

- [Environment Variables](#environment-variables)
- [Configuration Hierarchy](#configuration-hierarchy)
- [Component-Specific Configuration](#component-specific-configuration)
- [Network Configuration](#network-configuration)
- [Storage Configuration](#storage-configuration)
- [Authentication & Security](#authentication--security)
- [Resource Configuration](#resource-configuration)
- [Logging Configuration](#logging-configuration)
- [Advanced Configuration](#advanced-configuration)

---

## Environment Variables

### Controller Configuration

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `CONTROLLER_NAMESPACE` | `c8s-system` | No | Kubernetes namespace for controller deployment |
| `CONTROLLER_LOG_LEVEL` | `info` | No | Log level: debug, info, warn, error |
| `CONTROLLER_WORKERS` | `3` | No | Number of concurrent workers |
| `CONTROLLER_RESYNC_PERIOD` | `15m` | No | CRD resync interval |
| `CONTROLLER_LEADER_ELECT` | `true` | No | Enable leader election for HA |

### API Server Configuration

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `API_SERVER_PORT` | `8080` | No | HTTP port for API server |
| `API_SERVER_TLS_PORT` | `8443` | No | HTTPS port (requires TLS config) |
| `API_SERVER_LOG_LEVEL` | `info` | No | Log level |
| `API_SERVER_CORS_ORIGINS` | `http://localhost:3000` | No | Comma-separated CORS origins |
| `API_SERVER_REQUEST_TIMEOUT` | `30s` | No | Request timeout duration |
| `API_SERVER_ENABLE_DOCS` | `false` | No | Enable OpenAPI documentation |

### Webhook Service Configuration

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `WEBHOOK_PORT` | `9000` | No | HTTP port for webhooks |
| `WEBHOOK_SECRET_HEADER` | `X-Hub-Signature-256` | No | Header containing webhook signature |
| `WEBHOOK_VERIFY_SSL` | `true` | No | Verify SSL certificates on outbound webhooks |
| `WEBHOOK_TIMEOUT` | `10s` | No | Webhook delivery timeout |

### Authentication Configuration

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `AUTH_MODE` | `jwt` | No | Authentication mode: jwt, oauth2, none |
| `JWT_ALGORITHM` | `HS256` | No | JWT signing algorithm |
| `JWT_SECRET` | (empty) | If AUTH_MODE=jwt | Secret for HS256 signing |
| `JWT_ISSUER` | (empty) | If AUTH_MODE=jwt | Expected JWT issuer |
| `JWT_AUDIENCE` | (empty) | If AUTH_MODE=jwt | Expected JWT audience |
| `JWT_EXPIRY_TOLERANCE` | `0s` | No | Clock skew tolerance |

### Authorization Configuration

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `AUTHZ_MODE` | `k8s-rbac` | No | Authorization mode |
| `AUTHZ_DEFAULT_ROLE` | `viewer` | No | Default role for new users |
| `AUTHZ_CACHE_TTL` | `5m` | No | Role cache time-to-live |

### Storage Configuration

| Variable | Default | Required | Description |
|----------|---------|----------|-------------|
| `STORAGE_TYPE` | `s3` | No | Storage backend: s3, gcs, azblob |
| `S3_ENDPOINT` | (empty) | If STORAGE_TYPE=s3 | S3 endpoint URL |
| `S3_BUCKET` | (empty) | If STORAGE_TYPE=s3 | S3 bucket name |
| `S3_REGION` | `us-east-1` | No | AWS region |
| `S3_ACCESS_KEY` | (empty) | If STORAGE_TYPE=s3 | AWS access key |
| `S3_SECRET_KEY` | (empty) | If STORAGE_TYPE=s3 | AWS secret key |

---

## Configuration Hierarchy

Configuration is applied in this order (later overrides earlier):

1. **Defaults** - Built-in defaults
2. **Environment Variables** - Override defaults
3. **ConfigMap** - Kubernetes ConfigMap in c8s-system namespace
4. **Secret** - Kubernetes Secret for sensitive data
5. **Runtime Flags** - Command-line arguments (highest priority)

### Example ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: c8s-config
  namespace: c8s-system
data:
  controller.log-level: debug
  controller.workers: "5"
  api-server.cors-origins: "http://localhost:3000,https://app.example.com"
```

### Example Secret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: c8s-secrets
  namespace: c8s-system
type: Opaque
stringData:
  jwt-secret: "your-secure-secret-here"
  s3-access-key: "AKIA..."
  s3-secret-key: "..."
```

---

## Component-Specific Configuration

### Controller

**Configuration File**: `/etc/c8s/controller-config.yaml`

```yaml
# Controller-specific settings
controller:
  namespace: c8s-system
  logLevel: info
  workers: 3
  resyncPeriod: 15m
  leaderElect: true

  # PipelineRun behavior
  pipelineRun:
    defaultTimeout: 1h
    maxRetries: 3
    retryBackoff: 5s

  # Job configuration
  job:
    imagePolicy: Always
    imagePullSecrets:
      - name: docker-creds
    nodeSelector:
      workload: ci
```

### API Server

**Configuration File**: `/etc/c8s/api-server-config.yaml`

```yaml
# API Server settings
api:
  port: 8080
  tlsPort: 8443
  tlsCert: /etc/certs/tls.crt
  tlsKey: /etc/certs/tls.key

  # CORS configuration
  cors:
    origins:
      - http://localhost:3000
      - https://app.example.com
    credentials: true
    maxAge: 3600

  # Request handling
  requests:
    timeout: 30s
    maxBodySize: 10MB
    maxHeaderSize: 1MB
```

### Webhook Service

**Configuration File**: `/etc/c8s/webhook-config.yaml`

```yaml
# Webhook settings
webhook:
  port: 9000

  # GitHub webhooks
  github:
    secretHeader: X-Hub-Signature-256
    verifySSL: true

  # GitLab webhooks
  gitlab:
    tokenHeader: X-Gitlab-Token
    verifySSL: true
```

---

## Network Configuration

### Ingress Setup

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: c8s-api
  namespace: c8s-system
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - api.example.com
      secretName: c8s-tls
  rules:
    - host: api.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: c8s-api
                port:
                  number: 8080
```

### Service Configuration

```yaml
apiVersion: v1
kind: Service
metadata:
  name: c8s-api
  namespace: c8s-system
spec:
  type: ClusterIP
  selector:
    app: c8s-api
  ports:
    - name: http
      port: 8080
      targetPort: 8080
    - name: https
      port: 8443
      targetPort: 8443
```

---

## Storage Configuration

### S3-Compatible Storage (Minio, AWS S3, etc.)

```bash
# Environment variables for S3
export STORAGE_TYPE=s3
export S3_ENDPOINT=https://minio.example.com:9000
export S3_BUCKET=c8s-logs
export S3_REGION=us-east-1
export S3_ACCESS_KEY=minioadmin
export S3_SECRET_KEY=minioadmin
```

### Google Cloud Storage

```bash
export STORAGE_TYPE=gcs
export GCS_PROJECT_ID=my-project
export GCS_BUCKET=c8s-logs
export GOOGLE_APPLICATION_CREDENTIALS=/var/secrets/gcp-key.json
```

### Azure Blob Storage

```bash
export STORAGE_TYPE=azblob
export AZURE_STORAGE_ACCOUNT=myaccount
export AZURE_STORAGE_CONTAINER=c8s-logs
export AZURE_STORAGE_KEY=...
```

---

## Authentication & Security

### JWT Configuration

For HS256 (Shared Secret):

```bash
export AUTH_MODE=jwt
export JWT_ALGORITHM=HS256
export JWT_SECRET="your-256-bit-secret-key"
export JWT_ISSUER="https://auth.example.com"
export JWT_AUDIENCE="c8s-api"
```

For RS256 (RSA Public Key):

```bash
export AUTH_MODE=jwt
export JWT_ALGORITHM=RS256
export JWT_PUBLIC_KEY=/etc/certs/jwt-public-key.pem
export JWT_ISSUER="https://auth.example.com"
```

### TLS/HTTPS Configuration

```bash
# Enable HTTPS
export API_SERVER_TLS_PORT=8443
export TLS_CERT_PATH=/etc/certs/tls.crt
export TLS_KEY_PATH=/etc/certs/tls.key
```

### RBAC Configuration

```yaml
# Example RBAC RoleBinding for user
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: alice-editor
  namespace: c8s-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: c8s-editor
subjects:
  - kind: User
    name: alice@example.com
```

---

## Resource Configuration

### CPU and Memory Limits

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: c8s-controller
spec:
  containers:
  - name: controller
    resources:
      requests:
        cpu: 100m
        memory: 256Mi
      limits:
        cpu: 500m
        memory: 512Mi
```

### Pod Disruption Budgets

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: c8s-api-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: c8s-api
```

---

## Logging Configuration

### Log Levels

Available levels (in order of detail):
- `debug` - Very detailed debugging information
- `info` - General informational messages (default)
- `warn` - Warning messages for potentially problematic situations
- `error` - Error messages for failures

### Log Output

```bash
# Log to stdout (default for containers)
export LOG_OUTPUT=stdout

# Log to file
export LOG_OUTPUT=file
export LOG_FILE_PATH=/var/log/c8s/controller.log
export LOG_MAX_SIZE=100 # MB
export LOG_MAX_AGE=7    # days
export LOG_BACKUPS=3    # number of backups
```

### Structured Logging

C8S uses structured JSON logging. Example output:

```json
{
  "timestamp": "2025-11-02T15:30:45Z",
  "level": "info",
  "component": "controller",
  "message": "Processing PipelineRun",
  "pipelinerun": "my-run-123",
  "namespace": "default",
  "duration_ms": 42
}
```

---

## Advanced Configuration

### Custom Certificate Authorities

```bash
# For S3-compatible storage with self-signed certs
export S3_CA_CERT=/etc/certs/ca.crt
export S3_INSECURE_SKIP_VERIFY=false
```

### Proxy Configuration

```bash
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=https://proxy.example.com:8080
export NO_PROXY=localhost,.svc.cluster.local
```

### Performance Tuning

```yaml
# Controller optimization
controller:
  workers: 8          # Increase for high-volume workloads
  resyncPeriod: 30m   # Reduce for real-time responsiveness

  # API Server optimization
  api:
    port: 8080
    workers: 4
    connectionPool: 100
```

### Debugging

Enable verbose logging for troubleshooting:

```bash
export LOG_LEVEL=debug
export CONTROLLER_LOG_LEVEL=debug
export API_SERVER_LOG_LEVEL=debug
export WEBHOOK_LOG_LEVEL=debug
```

---

## Configuration Validation

### Validate Configuration

```bash
# Check environment variables
c8s config validate

# Check ConfigMap
kubectl get configmap c8s-config -n c8s-system -o yaml | c8s config validate

# Health check
curl http://localhost:8080/health
```

### Common Configuration Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| "Invalid JWT" | Wrong secret or algorithm | Verify JWT_SECRET matches issuer, check JWT_ALGORITHM |
| "S3 connection refused" | Wrong endpoint or network | Check S3_ENDPOINT, verify network connectivity |
| "CORS error in dashboard" | Origin not in whitelist | Add your domain to API_SERVER_CORS_ORIGINS |
| "Pod evicted" | Resource limits exceeded | Increase CPU/memory requests or limits |

---

## Next Steps

- [Getting Started](./GETTING_STARTED.md) - Quick start guide
- [Operator Guide](./OPERATOR_GUIDE.md) - Deployment best practices
- [Troubleshooting](./TROUBLESHOOTING.md) - Common issues and fixes
