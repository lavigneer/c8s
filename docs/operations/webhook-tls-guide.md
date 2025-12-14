# Webhook TLS Configuration Guide

Quick reference for configuring TLS certificates for the C8S webhook.

## Quick Start

### Option 1: Cert-Manager (Recommended)

```bash
# 1. Install cert-manager
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install cert-manager jetstack/cert-manager --namespace cert-manager --create-namespace --set installCRDs=true

# 2. Deploy C8S with automatic certificate management
helm install c8s ./chart/c8s \
  -n c8s-system \
  --create-namespace \
  -f chart/c8s/values-certmanager.yaml

# Done! Certificates are automatically generated and renewed
```

### Option 2: Manual Configuration

If you prefer to manage certificates manually:

```bash
# 1. Generate self-signed certificate
openssl req -x509 -newkey rsa:4096 -keyout webhook.key -out webhook.crt \
  -days 365 -nodes \
  -subj "/CN=c8s-webhook.c8s-system.svc"

# 2. Base64 encode the certificate
CA_BUNDLE=$(base64 -w 0 < webhook.crt)

# 3. Deploy C8S with the certificate
helm install c8s ./chart/c8s \
  -n c8s-system \
  --create-namespace \
  --set components.webhook.tls.caBundle="$CA_BUNDLE"
```

## Comparison

| Feature | Cert-Manager | Manual |
|---------|--------------|--------|
| Automatic Generation | ✓ | ✗ |
| Automatic Renewal | ✓ | ✗ |
| Manual CA Bundle Configuration | ✗ | ✓ |
| Supports Let's Encrypt | ✓ | ✗ |
| Development Use | ✓ | ✓ |
| Production Use | ✓ | ✓ |
| Complexity | Medium | Low |
| Maintenance | Low | High |

## Verification

Check if certificates are properly configured:

```bash
# Verify caBundle is populated
kubectl get validatingwebhookconfigurations c8s-validating-webhook -o yaml | grep caBundle

# For cert-manager: should show cert-manager.io/inject-ca-from annotation
kubectl get validatingwebhookconfigurations c8s-validating-webhook -o yaml | grep cert-manager

# For manual: should show base64-encoded certificate in caBundle field
```

## Troubleshooting

### Deployment Fails: "caBundle is empty"

**Cause**: Neither cert-manager is enabled nor a manual caBundle is provided.

**Fix**:
```bash
# Either use cert-manager:
helm upgrade c8s ./chart/c8s -n c8s-system -f chart/c8s/values-certmanager.yaml

# Or provide manual caBundle:
helm upgrade c8s ./chart/c8s -n c8s-system \
  --set components.webhook.tls.caBundle="$(base64 -w 0 < webhook.crt)"
```

### Webhooks Not Validating

**Cause**: caBundle mismatch between webhook certificate and ValidatingWebhookConfiguration.

**Verify**:
```bash
# Extract caBundle from ValidatingWebhookConfiguration
kubectl get validatingwebhookconfigurations c8s-validating-webhook -o jsonpath='{.webhooks[0].clientConfig.caBundle}' | base64 -d > vwc-ca.crt

# Extract certificate from webhook pod
kubectl get secret c8s-webhook-tls -n c8s-system -o jsonpath='{.data.tls\.crt}' | base64 -d > webhook-ca.crt

# Compare
diff vwc-ca.crt webhook-ca.crt  # Should show no differences
```

### Certificate Not Renewing (cert-manager)

```bash
# Check certificate status
kubectl describe certificate c8s-webhook-tls -n c8s-system

# Check cert-manager logs
kubectl logs -n cert-manager -l app=cert-manager | grep c8s-webhook-tls

# Force renewal by deleting the secret (cert-manager will recreate it)
kubectl delete secret c8s-webhook-tls -n c8s-system
```

## Production Deployment

### Recommended Setup

```bash
# 1. Install cert-manager
helm repo add jetstack https://charts.jetstack.io
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set installCRDs=true \
  --version v1.13.0

# 2. Configure Let's Encrypt (optional, for external-facing webhooks)
# Update values to use letsencrypt-prod issuer

# 3. Deploy C8S with cert-manager
helm install c8s ./chart/c8s \
  -n c8s-system \
  --create-namespace \
  -f chart/c8s/values-certmanager.yaml

# 4. Monitor certificate renewal
kubectl get certificates -n c8s-system --watch
```

### High Availability

For HA deployments:

```bash
# Deploy with multiple webhook replicas
helm install c8s ./chart/c8s \
  -n c8s-system \
  --create-namespace \
  -f chart/c8s/values-certmanager.yaml \
  --set components.webhook.replicas=3
```

All replicas will share the same certificate (stored in Secret), managed by cert-manager.

## Related Documentation

- [Cert-Manager Setup](./cert-manager-setup.md) - Detailed cert-manager configuration
- [Helm Values Reference](./helm-values-reference.md) - All available configuration options
- [Troubleshooting Guide](./troubleshooting.md) - General troubleshooting steps
