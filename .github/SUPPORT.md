# Getting Support for C8S

Welcome to C8S! We're here to help you get the most out of our Kubernetes-native CI system.

## 📚 Documentation

Before asking for help, please check our comprehensive documentation:

- **[Getting Started Guide](../docs/guides/getting-started.md)** - Installation and your first pipeline
- **[Architecture Guide](../docs/guides/architecture.md)** - How C8S works under the hood
- **[Pipeline Syntax](../docs/guides/pipeline-syntax.md)** - Complete YAML configuration reference
- **[Troubleshooting Guide](../docs/guides/troubleshooting.md)** - Common issues and solutions
- **[FAQ](../docs/guides/troubleshooting.md#frequently-asked-questions)** - Frequently asked questions

### For Developers

- **[5-Minute Quick Start](../docs/development/QUICKSTART.md)** - Get developing fast
- **[Tilt Workflow](../docs/development/TILT-WORKFLOW.md)** - Local development guide
- **[Contributing Guide](../CONTRIBUTING.md)** - How to contribute
- **[Testing Guide](../tests/README.md)** - Running and writing tests

### Quick References

- **[Examples](../examples/README.md)** - 4 complete pipeline configurations
- **[API Reference](../specs/001-build-a-continuous/contracts/openapi.yaml)** - REST API specification
- **[CLI Reference](../cmd/README.md#c8s)** - Command-line usage

## 💬 Community Support

### GitHub Discussions

For questions, ideas, and general discussion:

👉 **[GitHub Discussions](https://github.com/lavigneer/c8s/discussions)**

- **Q&A**: Ask questions and get help from the community
- **Ideas**: Share feature ideas and suggestions
- **Show & Tell**: Share your C8S pipelines and use cases
- **General**: Everything else!

### GitHub Issues

For bug reports and feature requests, use GitHub Issues:

👉 **[Create an Issue](https://github.com/lavigneer/c8s/issues/new/choose)**

- **Bug Report**: Report a bug or problem
- **Feature Request**: Suggest a new feature or enhancement

**Before creating an issue**:
1. Search existing issues to avoid duplicates
2. Check the [Troubleshooting Guide](../docs/guides/troubleshooting.md)
3. Gather relevant information (logs, configuration, steps to reproduce)

## 🐛 Reporting Bugs

When reporting a bug, please include:

1. **C8S Version**: `c8s version` or image tag
2. **Environment**:
   - Kubernetes version: `kubectl version`
   - Installation method: Helm, Tilt, kubectl
   - Operating system
3. **Steps to Reproduce**:
   - Minimal pipeline configuration
   - Commands executed
   - Expected vs. actual behavior
4. **Logs**:
   ```bash
   # Controller logs
   kubectl logs -n c8s-system deployment/c8s-controller

   # Webhook logs
   kubectl logs -n c8s-system deployment/c8s-webhook

   # PipelineRun status
   kubectl get pipelineruns <run-name> -o yaml
   ```
5. **Configuration**: Relevant YAML files (sanitize sensitive data!)

## 🔒 Security Issues

**DO NOT** report security vulnerabilities through public GitHub issues.

Instead, please follow our [Security Policy](../SECURITY.md):

📧 Email: security@c8s.dev (replace with actual email)

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

We will respond within 48 hours.

## ❓ Frequently Asked Questions

### Installation & Setup

**Q: How do I install C8S?**

A: See the [Getting Started Guide](../docs/guides/getting-started.md). Quick start:
```bash
helm install c8s ./chart/c8s -n c8s-system --create-namespace
```

**Q: Can I use C8S with my local Kubernetes cluster?**

A: Yes! Use [Tilt](../docs/development/TILT-WORKFLOW.md) for local development:
```bash
tilt up
```

### Pipeline Configuration

**Q: Where should I put my `.c8s.yaml` file?**

A: In the root of your Git repository.

**Q: How do I trigger a pipeline manually?**

A: Use the CLI:
```bash
c8s run my-pipeline --commit=$(git rev-parse HEAD) --branch=$(git branch --show-current)
```

**Q: Can I use private Docker images?**

A: Yes! Create a Kubernetes Secret with your registry credentials and reference it in your pipeline.

### Troubleshooting

**Q: My pipeline is stuck "Pending"**

A: Check:
1. Pod status: `kubectl get pods -l c8s.dev/pipeline=<name>`
2. Events: `kubectl get events -n c8s-system --sort-by='.lastTimestamp'`
3. Resource availability: `kubectl top nodes`

**Q: Where are my pipeline logs?**

A: Logs are stored in S3-compatible storage. View via:
- Dashboard: `http://c8s-api-server/runs/<run-id>`
- CLI: `c8s logs <run-id> --follow`
- API: `GET /api/v1/runs/<run-id>/logs`

**Q: How do I debug webhook issues?**

A: Check webhook logs:
```bash
kubectl logs -f deployment/c8s-webhook -n c8s-system
```

Verify webhook configuration in GitHub:
- Payload URL is correct
- Secret matches
- Content type is `application/json`

## 🤝 Contributing

We welcome contributions! Please read:

- **[Contributing Guide](../CONTRIBUTING.md)** - Guidelines and workflow
- **[Development Guide](../docs/development/QUICKSTART.md)** - Setting up your dev environment
- **[Code of Conduct](../CODE_OF_CONDUCT.md)** - Community standards (if applicable)

### Quick Contribution Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `make test-all`
5. Commit: `git commit -m "[Feature] Add amazing feature"`
6. Push: `git push origin feature/amazing-feature`
7. Open a Pull Request

## 📖 Additional Resources

### External Resources

- **[Kubernetes Documentation](https://kubernetes.io/docs/)**
- **[Kubebuilder Book](https://book.kubebuilder.io/)** - For understanding CRDs
- **[Tilt Documentation](https://docs.tilt.dev/)** - Local development

### Related Projects

- **[Tekton](https://tekton.dev/)** - Kubernetes-native CI/CD
- **[Argo Workflows](https://argoproj.github.io/workflows/)** - Container-native workflows
- **[Jenkins X](https://jenkins-x.io/)** - CI/CD for Kubernetes

## 📬 Contact

- **GitHub**: [@lavigneer](https://github.com/lavigneer)
- **Project**: [C8S](https://github.com/lavigneer/c8s)
- **Documentation**: [docs/](../docs/)

---

**Need help right away?** Check the [Troubleshooting Guide](../docs/guides/troubleshooting.md) for quick solutions to common problems!
