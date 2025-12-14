# Security Policy

## Supported Versions

We release security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |
| < 0.1   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

### How to Report

If you discover a security vulnerability in C8S, please report it by emailing:

**security@example.com** (Replace with actual security contact)

Please include the following information:

- **Description**: A clear description of the vulnerability
- **Impact**: What an attacker could achieve by exploiting this vulnerability
- **Reproduction**: Step-by-step instructions to reproduce the issue
- **Version**: The C8S version affected
- **Environment**: Kubernetes version and deployment environment
- **Suggested Fix**: If you have a proposed solution

### What to Expect

1. **Acknowledgment**: We will acknowledge receipt of your report within 48 hours
2. **Initial Assessment**: We will provide an initial assessment within 5 business days
3. **Updates**: We will keep you informed of our progress
4. **Resolution**: We aim to resolve critical vulnerabilities within 30 days
5. **Disclosure**: We will coordinate public disclosure with you

### Security Update Process

1. **Triage**: Security team assesses severity and impact
2. **Fix Development**: Patch is developed and tested
3. **Review**: Security fix undergoes additional review
4. **Release**: Security update is released
5. **Announcement**: Security advisory is published
6. **CVE Assignment**: CVE is requested if applicable

## Security Best Practices

### Deployment

- **Use TLS**: Always enable TLS for webhook and API server endpoints
- **Network Policies**: Restrict pod-to-pod communication
- **RBAC**: Follow principle of least privilege
- **Secrets Management**: Use Kubernetes Secrets, never hardcode secrets
- **Image Scanning**: Scan Docker images for vulnerabilities
- **Resource Limits**: Set CPU/memory limits on all pods

### Pipeline Security

- **Trusted Images**: Only use images from trusted registries
- **Secret Injection**: Use C8S secret references, not environment variables in YAML
- **Log Masking**: Sensitive data is automatically masked in logs
- **Network Isolation**: Pipeline jobs run in isolated namespaces
- **Artifact Signing**: Consider signing build artifacts (future feature)

### Webhook Security

- **Signature Validation**: Always configure webhook secrets
- **HTTPS Only**: Use HTTPS URLs for webhooks
- **IP Filtering**: Restrict webhook access to known IPs if possible
- **Rate Limiting**: Configure rate limits to prevent abuse

### API Security

- **Authentication**: Use JWT tokens for API access
- **Authorization**: Implement fine-grained access control
- **TLS**: Use TLS 1.2+ for all API traffic
- **Input Validation**: API validates all inputs
- **Rate Limiting**: Protect against DoS attacks

## Known Security Considerations

### Secrets in Logs

C8S automatically masks secrets in logs, but:
- Secrets could appear in command output
- Users should avoid echoing secrets
- Review logs for accidental secret exposure

### Container Escape

Pipeline jobs run as containers:
- Use least-privileged containers
- Avoid running as root unless necessary
- Be cautious with volume mounts
- Consider using security policies (PSP/PSA)

### Supply Chain

- **Dependencies**: We regularly update dependencies
- **Image Scanning**: Official images are scanned for vulnerabilities
- **SBOMs**: Software Bill of Materials available (future)

## Security Features

### Current

- **Secret Masking**: Automatic redaction of secrets in logs
- **RBAC Integration**: Kubernetes-native access control
- **TLS Support**: HTTPS for webhooks and API
- **Signature Validation**: Webhook HMAC verification
- **Resource Isolation**: Jobs run in isolated pods
- **Audit Logging**: All API calls are logged

### Planned

- [ ] Mutual TLS (mTLS) for inter-component communication
- [ ] Image signature verification
- [ ] Artifact signing and verification
- [ ] Advanced RBAC policies
- [ ] Security scanning in pipelines
- [ ] Compliance reporting (SLSA, SOC2)

## Security Advisories

Security advisories will be published at:

- **GitHub Security Advisories**: https://github.com/lavigneer/c8s/security/advisories
- **Release Notes**: https://github.com/lavigneer/c8s/releases

Subscribe to GitHub notifications to receive security updates.

## Compliance

### Certifications

C8S is designed to support compliance with:
- **SOC 2** (in progress)
- **GDPR** (data handling practices)
- **SLSA Level 3** (supply chain security, planned)

### Audit Logs

C8S maintains audit logs for:
- API requests (authentication, authorization, actions)
- Webhook events (source, signature validation)
- Pipeline executions (user, commit, status)
- Secret access (which jobs accessed which secrets)

## Security Tools

### Scanning

We use the following tools to maintain security:

- **Dependabot**: Automatic dependency updates
- **Trivy**: Container image vulnerability scanning
- **gosec**: Go security checker
- **golangci-lint**: Go code linting with security checks
- **OWASP Dependency-Check**: Dependency vulnerability scanning

### CI/CD Security

- All commits are signed (enforcement planned)
- PRs require review before merge
- CI runs security scans automatically
- Dependencies are pinned and regularly updated

## Responsible Disclosure

We follow responsible disclosure practices:

1. **Private Reporting**: Vulnerabilities reported privately
2. **Coordinated Disclosure**: We work with reporters to coordinate disclosure
3. **Credit**: We acknowledge reporters (with permission)
4. **CVE Assignment**: We request CVEs for significant vulnerabilities
5. **Public Advisory**: We publish advisories after fixes are available

## Security Contact

- **Email**: security@example.com
- **PGP Key**: Available at https://example.com/security-pgp-key.asc

## Bug Bounty

We do not currently offer a bug bounty program, but we greatly appreciate security researchers who report vulnerabilities responsibly.

## Additional Resources

- [OWASP Kubernetes Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Kubernetes_Security_Cheat_Sheet.html)
- [Kubernetes Security Best Practices](https://kubernetes.io/docs/concepts/security/)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)

## Questions

For security-related questions that are not vulnerabilities, please:
- Open a [GitHub Discussion](https://github.com/lavigneer/c8s/discussions)
- Email general@example.com

---

**Last Updated**: 2025-01-15
**Next Review**: 2025-04-15
