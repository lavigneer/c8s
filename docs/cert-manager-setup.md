# Cert-Manager Setup for C8S Webhook TLS Certificates

This guide explains how to set up cert-manager to automatically provision and manage TLS certificates for the C8S webhook, eliminating the need for manual CA bundle configuration.

## Why Cert-Manager?

- **Automatic Certificate Generation**: Generates self-signed or ACME certificates automatically
- **Automatic Renewal**: Renews certificates before expiration
- **CA Bundle Injection**: Automatically injects the CA certificate into ValidatingWebhookConfiguration
- **No Manual Configuration**: No need to generate and encode certificates manually
- **Production Ready**: Supports Let's Encrypt and other ACME providers

## Installation

### 1. Install cert-manager

```bash
# Add cert-manager Helm repository
helm repo add jetstack https://charts.jetstack.io
helm repo update

# Install cert-manager with CRDs
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set installCRDs=true \
  --version v1.13.0  # Use latest stable version
```

Verify installation:
```bash
kubectl get pods -n cert-manager
# Should see cert-manager, cert-manager-webhook, and cert-manager-cainjector pods
```

### 2. Deploy C8S with Cert-Manager Integration

```bash
# Option 1: Using the cert-manager values file
helm install c8s ./chart/c8s \
  -n c8s-system \
  --create-namespace \
  -f chart/c8s/values-certmanager.yaml

# Option 2: Manual cert-manager enablement
helm install c8s ./chart/c8s \
  -n c8s-system \
  --create-namespace \
  --set certManager.enabled=true \
  --set certManager.issuerType=self-signed
```

## Configuration

### Self-Signed Certificates (Development)

Best for development environments where you trust your own infrastructure.

```yaml
certManager:
  enabled: true
  issuerType: self-signed

components:
  webhook:
    tls:
      certManager:
        enabled: true
        issuerName: c8s-webhook-selfsigned
        issuerKind: Issuer
        duration: 2160h  # 90 days
        renewBefore: 720h  # Renew 30 days before expiry
```

### Let's Encrypt (Production)

For production deployments that need publicly trusted certificates.

```bash
helm install c8s ./chart/c8s \
  -n c8s-system \
  --set certManager.enabled=true \
  --set certManager.issuerType=letsencrypt-prod \
  --set certManager.acmeEmail=admin@example.com
```

**Note**: Let's Encrypt for internal services requires DNS challenges or HTTP-01 challenges. This setup is more complex and typically only needed for external-facing webhooks.

## Verification

### 1. Check Certificate Status

```bash
# List all certificates
kubectl get certificates -n c8s-system

# Check certificate details
kubectl describe certificate c8s-webhook-tls -n c8s-system

# Expected output:
# Status:
#   Conditions:
#   - Issuing
#   - Ready (once cert is issued)
```

### 2. Verify CA Bundle Injection

```bash
# Check if ValidatingWebhookConfiguration has caBundle populated
kubectl get validatingwebhookconfigurations c8s-validating-webhook -o yaml | grep -A 5 caBundle

# Should show a base64-encoded certificate, not an empty string
```

### 3. Check Webhook Pod Mounts

```bash
# Verify the webhook pod has mounted the certificate
kubectl describe pod -n c8s-system -l app.kubernetes.io/name=c8s,app.kubernetes.io/component=webhook

# Look for:
# Mounts:
#   /etc/webhook/certs from webhook-certs (ro)
```

## Certificate Renewal

Cert-manager automatically renews certificates before expiration:

```yaml
components:
  webhook:
    tls:
      certManager:
        duration: 2160h     # Certificate valid for 90 days
        renewBefore: 720h   # Renew 30 days before expiry (60 days remaining)
```

To monitor renewal:
```bash
# Watch certificate status
kubectl get certificates -n c8s-system --watch

# Check cert-manager logs for renewal activity
kubectl logs -n cert-manager -l app=cert-manager --tail=50 -f
```

## Troubleshooting

### Certificate Not Issuing

```bash
# Check certificate status
kubectl describe certificate c8s-webhook-tls -n c8s-system

# Look for:
# - Status: False/Issuing
# - Message: details about what's blocking issuance

# Check issuer status
kubectl describe issuer c8s-webhook-selfsigned -n c8s-system

# Check cert-manager logs
kubectl logs -n cert-manager -l app=cert-manager
```

### CABundle Not Injected

```bash
# Verify cert-manager CA injector is running
kubectl get pod -n cert-manager | grep cainjector

# Check the ValidatingWebhookConfiguration annotation
kubectl get validatingwebhookconfigurations c8s-validating-webhook -o yaml | grep cert-manager

# Should show: cert-manager.io/inject-ca-from: c8s-system/c8s-webhook-tls

# Restart the webhook to trigger injection
kubectl rollout restart deployment c8s-webhook -n c8s-system
```

### Webhook Pod Can't Read Certificate

```bash
# Verify the secret exists
kubectl get secret c8s-webhook-tls -n c8s-system -o yaml

# Check webhook pod mount permissions
kubectl exec -it <webhook-pod> -n c8s-system -- ls -la /etc/webhook/certs/

# Should show tls.crt and tls.key with read permissions
```

## Migration from Manual Configuration

If you already have C8S deployed without cert-manager:

### 1. Install cert-manager
Follow the installation steps above.

### 2. Update C8S Deployment
```bash
helm upgrade c8s ./chart/c8s \
  -n c8s-system \
  -f chart/c8s/values-certmanager.yaml
```

This will:
- Create the Certificate and Issuer resources
- Update the ValidatingWebhookConfiguration with cert-manager annotations
- Mount the generated certificate in the webhook pod
- Automatically inject the caBundle

### 3. Verify
```bash
# Wait for certificate to be ready
kubectl get certificate -n c8s-system --watch

# Verify caBundle is populated
kubectl get validatingwebhookconfigurations c8s-validating-webhook -o yaml | grep caBundle
```

## Performance Considerations

- **Cert-Manager Overhead**: Minimal - runs as separate pods in cert-manager namespace
- **Certificate Refresh**: Transparent - no downtime during renewal
- **CA Injection**: Automatic via cainjector webhook - no manual restart needed

## Security Notes

- Self-signed certificates should only be used in development/internal environments
- For production, consider Let's Encrypt or your organizational CA
- Cert-manager securely manages private keys as Kubernetes secrets
- RBAC rules prevent unauthorized certificate creation/modification
- Certificates are automatically rotated before expiration

## Next Steps

1. Install cert-manager using the commands above
2. Deploy C8S with cert-manager integration
3. Verify certificate and caBundle injection
4. Monitor cert-manager logs for any issues
5. Configure certificate renewal settings as needed

## References

- [Cert-Manager Official Documentation](https://cert-manager.io)
- [Cert-Manager Webhook Configuration](https://cert-manager.io/docs/usage/webhook/)
- [Kubernetes Admission Webhooks](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
