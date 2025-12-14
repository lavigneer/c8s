# hack/

Kubebuilder code generation utilities and templates.

## Overview

The `hack/` directory contains utilities and templates used by Kubebuilder's code generation tools. This follows the Kubernetes project convention of placing build scripts and code generation helpers in a `hack/` directory.

## Contents

### boilerplate.go.txt

Copyright header template automatically prepended to all generated Go files.

**Used by**: Kubebuilder code generators when running `make generate` or `make manifests`

**Template**:
```go
/*
Copyright 2025 C8S Authors.

Licensed under the Apache License, Version 2.0 (the "License");
...
*/
```

## When This Gets Used

The boilerplate template is automatically applied when running:

```bash
# Generate code (DeepCopy, clientset, etc.)
make generate

# Generate CRD manifests
make manifests

# Generate mocks
make mocks
```

## Customization

To update the copyright header for all generated files:

1. Edit `hack/boilerplate.go.txt`
2. Run `make generate manifests` to regenerate code with new header
3. Commit both the template and regenerated files

## Why "hack/"?

The name "hack/" is a Kubernetes project convention dating back to early Kubernetes development. It signals:
- Build scripts and utilities
- Code generation helpers
- Developer tools not part of the main codebase

Similar directories exist in:
- kubernetes/kubernetes
- kubernetes-sigs/controller-runtime
- kubernetes-sigs/kubebuilder

## Related

- [Kubebuilder Documentation](https://book.kubebuilder.io/)
- [Makefile](../Makefile) - Code generation targets
- [PROJECT](../PROJECT) - Kubebuilder project configuration
