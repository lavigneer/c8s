# C8S Documentation

Complete documentation for C8S users, developers, and operators.

## Getting Started

Start here if you're new to C8S:

- **[Getting Started](./guides/getting-started.md)** - Installation and your first pipeline
- **[Quick Start](../QUICK_START.md)** - 60-second setup guide (root directory)
- **[Troubleshooting](./guides/troubleshooting.md)** - Common issues and solutions

## User Guides

Essential reference materials for using C8S:

- **[Pipeline Syntax](./guides/pipeline-syntax.md)** - Complete YAML configuration reference
- **[Configuration](./guides/configuration.md)** - System configuration and setup
- **[Dashboard Features](./guides/dashboard-features.md)** - Web UI guide and keyboard shortcuts
- **[Helm Values Reference](./guides/helm-values-reference.md)** - Helm chart configuration options
- **[Troubleshooting](./guides/troubleshooting.md)** - Common issues and solutions

## Development

Resources for building and contributing to C8S:

- **[Development Guide](./development/development.md)** - Building and testing locally
- **[Tilt Setup](./development/tilt-setup.md)** - Local Kubernetes development with Tilt (primary workflow)
- **[Devbox Setup](./development/devbox-setup.md)** - Using Devbox for reproducible development environment
- **[Local Testing](./development/local-testing.md)** - Running tests locally
- **[Image Building](./development/image-building.md)** - Docker image building and registry setup
- **[Testing Matrix](./development/testing-matrix.md)** - Test coverage matrix and E2E testing
- **[Tilt Resource Tracking](./development/tilt-resource-tracking.md)** - Monitoring Tilt resources
- **[Dog-fooding Setup](./development/c8s-dogfooding.md)** - Using C8S to test itself

## Operations & Security

Guides for deploying and managing C8S in production:

- **[Operator Guide](./operations/operator-guide.md)** - Deployment and cluster management
- **[Authentication](./operations/authentication.md)** - JWT and API key setup and configuration
- **[HTTPS Setup](./operations/https-setup.md)** - TLS/HTTPS configuration and certificates
- **[Webhook TLS Guide](./operations/webhook-tls-guide.md)** - Webhook TLS certificate management
- **[Cert-Manager Setup](./operations/cert-manager-setup.md)** - Automatic TLS certificate provisioning
- **[GHCR Setup](./operations/ghcr-setup.md)** - GitHub Container Registry configuration
- **[Autoscaling](./operations/autoscaling.md)** - Pod autoscaling configuration and tuning

## Technical Reference

Deep technical specifications and contracts:

- [Feature Specification](../specs/001-build-a-continuous/spec.md) - Core CI/CD feature spec
- [Data Model](../specs/001-build-a-continuous/data-model.md) - CRD and data structure definitions
- [API Contracts](../specs/001-build-a-continuous/contracts/openapi.yaml) - REST API specification
- [CLI Specification](../specs/002-i-want-to/spec.md) - Command-line interface spec

## Directory Structure

```
docs/
├── README.md                          # This file
├── CONTRIBUTING.md                    # Contribution guidelines
├── guides/                            # User guides and references
│   ├── getting-started.md
│   ├── quick-start.md
│   ├── pipeline-syntax.md
│   ├── configuration.md
│   ├── dashboard-features.md
│   ├── helm-values-reference.md
│   ├── quick-reference.md
│   └── troubleshooting.md
├── development/                       # Development and contribution
│   ├── development.md
│   ├── tilt-setup.md              # Primary local dev workflow
│   ├── devbox-setup.md
│   ├── local-testing.md
│   ├── image-building.md
│   ├── testing-matrix.md
│   ├── tilt-resource-tracking.md
│   ├── screenshot-documentation.md
│   ├── c8s-dogfooding.md
│   └── c8s-dogfooding-quickstart.md
└── operations/                        # Operations and production
    ├── operator-guide.md
    ├── authentication.md
    ├── https-setup.md
    ├── webhook-tls-guide.md
    ├── cert-manager-setup.md
    ├── ghcr-setup.md
    └── autoscaling.md
```

## Finding What You Need

| I want to... | See... |
|---|---|
| Install C8S | [Getting Started](./guides/getting-started.md) |
| Write a pipeline | [Pipeline Syntax](./guides/pipeline-syntax.md) |
| Set up authentication | [Authentication](./operations/authentication.md) |
| Deploy to production | [Operator Guide](./operations/operator-guide.md) |
| Configure HTTPS | [HTTPS Setup](./operations/https-setup.md) |
| Develop C8S locally | [Development Guide](./development/development.md) |
| Use Tilt for local dev | [Tilt Setup](./development/tilt-setup.md) |
| Run tests | [Local Testing](./development/local-testing.md) |
| Troubleshoot issues | [Troubleshooting](./guides/troubleshooting.md) |
| See all settings | [Configuration](./guides/configuration.md) |

## Other Resources

- **[Contributing Guidelines](../CONTRIBUTING.md)** - How to contribute to C8S
- **[Development Guidelines](../CLAUDE.md)** - Project structure and standards
- **[Main README](../README.md)** - Project overview and features
