# HTTPS/TLS Setup Guide for C8S Dashboard

This guide explains how to configure HTTPS/TLS for the C8S Dashboard API server.

## Quick Start

### Development (Self-Signed Certificates)

For local development, generate self-signed certificates:

```bash
# Generate private key
openssl genrsa -out tls.key 2048

# Generate self-signed certificate (valid for 365 days)
openssl req -new -x509 -key tls.key -out tls.crt -days 365 \
  -subj "/C=US/ST=State/L=City/O=C8S/CN=localhost"

# Create a directory for certificates
mkdir -p config/certs
mv tls.key config/certs/
mv tls.crt config/certs/
```

Then run the API server with TLS enabled:

```bash
./api-server -enable-tls \
  -tls-cert config/certs/tls.crt \
  -tls-key config/certs/tls.key \
  -tls-port :8443
```

### Production (Let's Encrypt)

For production environments, use Let's Encrypt to obtain free, trusted certificates:

```bash
# Install certbot
# On macOS: brew install certbot
# On Ubuntu: sudo apt-get install certbot

# Obtain certificate (example for domain example.com)
certbot certonly --standalone -d example.com -d www.example.com

# Certificates will be in: /etc/letsencrypt/live/example.com/

# Run API server with Let's Encrypt certificates
./api-server -enable-tls \
  -tls-cert /etc/letsencrypt/live/example.com/fullchain.pem \
  -tls-key /etc/letsencrypt/live/example.com/privkey.pem
```

### Kubernetes (Cert-Manager)

For Kubernetes deployments, use cert-manager:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: c8s-dashboard-cert
spec:
  secretName: c8s-dashboard-tls
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
    - dashboard.example.com
```

Then reference the secret in your deployment:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-server
spec:
  template:
    spec:
      containers:
      - name: api-server
        args:
          - "-enable-tls"
          - "-tls-cert=/etc/tls/certs/tls.crt"
          - "-tls-key=/etc/tls/private/tls.key"
        volumeMounts:
        - name: tls-certs
          mountPath: /etc/tls/certs
          readOnly: true
        - name: tls-key
          mountPath: /etc/tls/private
          readOnly: true
      volumes:
      - name: tls-certs
        secret:
          secretName: c8s-dashboard-tls
          items:
          - key: tls.crt
            path: tls.crt
      - name: tls-key
        secret:
          secretName: c8s-dashboard-tls
          items:
          - key: tls.key
            path: tls.key
```

## Environment Variables

You can also set certificate paths via environment variables instead of command-line flags:

```bash
export TLS_CERT_PATH=/path/to/cert.pem
export TLS_KEY_PATH=/path/to/key.pem
./api-server -enable-tls
```

## HTTP to HTTPS Redirection

The API server will serve both HTTP (port 8080) and HTTPS (port 8443) when TLS is enabled.

For production, configure your reverse proxy (nginx, Envoy, etc.) to redirect HTTP to HTTPS:

```nginx
server {
    listen 80;
    server_name example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name example.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    location / {
        proxy_pass http://api-server:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Security Best Practices

1. **Use strong cipher suites**: The API server configures TLS 1.2+ by default
2. **Enable HSTS**: Security headers include `Strict-Transport-Security`
3. **Renew certificates**: Set up automatic renewal with certbot or cert-manager
4. **Monitor certificate expiration**: Keep track of certificate expiration dates
5. **Use strong key sizes**: 2048-bit RSA minimum, 4096-bit recommended for production

## Testing

Verify HTTPS is working:

```bash
# Test with curl (ignore self-signed certificate for dev)
curl -k https://localhost:8443/health

# Test with proper certificate validation
curl https://localhost:8443/health

# Check TLS version
openssl s_client -connect localhost:8443 -tls1_2

# Verify certificate details
openssl x509 -in tls.crt -text -noout
```

## Troubleshooting

### "Permission denied" error
If running on Linux, port 443 (HTTPS) is privileged and requires root or CAP_NET_BIND_SERVICE capability.

Solution: Use a different port (8443) or run through a reverse proxy.

### Certificate validation errors
Ensure the certificate and key file paths are correct and readable by the process.

### "Broken pipe" errors
Check that both certificate and key are valid and properly formatted (PEM format).

## References

- [Let's Encrypt](https://letsencrypt.org/)
- [Certbot Documentation](https://certbot.eff.org/)
- [Cert-Manager](https://cert-manager.io/)
- [OWASP HTTPS](https://cheatsheetseries.owasp.org/cheatsheets/HTTPS_Cheat_Sheet.html)
