# C8S Helm Chart Changelog

All notable changes to this C8S Helm chart will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2024-11-09

### Added (MVP Release)

#### Phase 1: Setup & Project Initialization
- Initial Helm chart structure with `/chart/c8s/` directory
- Chart metadata in Chart.yaml (name: c8s, version: 0.1.0)
- Root values file with all configuration parameters documented
- Environment-specific values files: values-dev.yaml, values-staging.yaml, values-prod.yaml
- Template helper functions for labels, names, and selectors

#### Phase 2: Foundational Components
- Namespace creation template
- Post-install health check hook
- Common ConfigMap and Secret templates
- RBAC setup with ClusterRole and ClusterRoleBinding
- Custom Resource Definition (PipelineRun) management

#### Phase 3: Component Deployments (User Story 1)
- API Server deployment with rolling update strategy
- Controller deployment with full RBAC permissions
- Webhook deployment with ValidatingWebhookConfiguration
- Frontend deployment with LoadBalancer/NodePort service
- All deployments include health checks and resource limits
- Zero-downtime rolling updates enabled

#### Phase 4: Configuration Customization (User Story 2)
- Per-component replicas configuration
- Per-component resource requests and limits
- Per-component image registry and tag overrides
- Environment-type selection (dev/staging/prod)
- Log level configuration
- Storage backend selection (local/PVC/S3-compatible)
- S3 credentials secret management

#### Phase 5: Health Verification (User Story 3)
- Enhanced post-install hook with color-coded output
- Component status reporting with replica counts
- Dashboard URL generation for LoadBalancer and NodePort services
- Remediation suggestions with kubectl commands
- Timeout handling with failure messages
- Overall health summary reporting

#### Phase 6: Lifecycle Management (User Story 4)
- RollingUpdate strategy with zero-downtime deployments
- Progress deadline configuration for slow clusters
- Release history tracking via `helm history`
- Rollback support via `helm rollback`
- Clean uninstall with PVC preservation
- Full cleanup option with data deletion
- Keep-history option for restore scenarios

#### Documentation
- Comprehensive README with quick start guide
- Configuration reference documentation
- Deployment examples for dev/staging/prod
- Lifecycle management guide (upgrade/downgrade/uninstall)
- Troubleshooting section with common issues
- Health verification output explanation

#### Testing
- Helm lint validation tests
- Template rendering tests
- Health verification E2E tests
- Lifecycle management E2E tests
- Multi-scenario deployment tests

#### Integration
- Tiltfile integration for local development
- Docker image building from source
- Automatic image rebuilding on source changes
- File watching for hot reload development
- Local Kubernetes deployment via Tilt

### Features

**Single-Command Deployment**:
```bash
helm install c8s ./chart/c8s -f values-dev.yaml
```

**Environment Presets**:
- Development: Single replicas, minimal resources, local storage
- Staging: Moderate resources, PVC storage
- Production: HA with 3 replicas, high resources, S3 storage

**Zero-Downtime Updates**:
- RollingUpdate strategy with maxSurge: 1, maxUnavailable: 0
- Automatic health check on new replicas
- Graceful pod termination

**Customization**:
- Values files for different environments
- CLI flag overrides: `--set key=value`
- All parameters documented in values.yaml

**Health Verification**:
- Automatic post-install health checks
- Component status reporting
- Dashboard access instructions
- Troubleshooting suggestions on failure

**Storage Flexibility**:
- Local (ephemeral) - for development
- PersistentVolumeClaim - for staging
- S3-compatible - for production (AWS S3, MinIO, etc.)

**RBAC Security**:
- Proper service accounts for each component
- ClusterRole with minimal required permissions
- No cluster-admin required

**Cross-Distribution Support**:
- Kubernetes 1.24+
- Works on k3s, kind, EKS, GKE, AKS, etc.
- No distribution-specific features

### Known Issues

- Health check hooks are currently non-blocking (deployment succeeds even if hooks fail)
- `/readyz` endpoints not yet implemented in components (health probes disabled in dev)
- Dashboard access via LoadBalancer requires external IP assignment (may take 1-5 minutes on cloud)

### Upgrade Instructions

This is the initial release. For future upgrades, see [Lifecycle Management](./README.md#lifecycle-management) section.

### Contributors

- Claude AI (Code generation)
- Team

---

## Future Releases

### [0.1.1] - Planned

**Focus**: Quality of Life Improvements
- Implement `/readyz` health check endpoints in components
- Enable health check probes in all environments
- Additional logging improvements
- Performance optimizations

### [0.2.0] - Planned

**Focus**: Advanced Features
- Horizontal Pod Autoscaling (HPA) support
- Pod Disruption Budgets (PDB) for high availability
- Network policies for security
- Backup and restore procedures
- Multi-cluster deployment support

### [0.3.0] - Planned

**Focus**: Production Hardening
- Resource quotas and limits
- Cost optimization guidelines
- Monitoring and metrics integration
- Audit logging
- Security scanning integration

---

## Migration Guides

### Upgrading from 0.1.0 to 0.1.1

```bash
# Standard upgrade - retains all configuration
helm upgrade c8s ./chart/c8s \
  -f values-prod.yaml \
  -n c8s-system

# Verify deployment
kubectl rollout status deployment/c8s-api-server -n c8s-system
```

### Downgrading

```bash
# View release history
helm history c8s -n c8s-system

# Rollback to previous version
helm rollback c8s 1 -n c8s-system
```

---

## Support

For issues, questions, or contributions:
- GitHub Issues: https://github.com/anthropics/c8s/issues
- Documentation: See [README.md](./README.md)
- Troubleshooting: See [README.md#troubleshooting](./README.md#troubleshooting)

---

## License

This Helm chart is part of C8S and is licensed under the same license as the main project.
