# Code Generation and Manifests

Generate Go code and Kubernetes manifests using controller-gen.

## Running in Devbox (Recommended)
All commands should be run in devbox to ensure consistent environments:
```bash
devbox run make generate        # Generate Go code (DeepCopy, clients, etc.)
devbox run make manifests       # Generate CRD and RBAC manifests
devbox run make all             # Full build + test + generate
```

Or enter devbox shell once:
```bash
devbox shell
make generate
make manifests
# etc...
```

## Quick Commands
```bash
make generate        # Generate Go code (DeepCopy, clients, etc.)
make manifests       # Generate CRD and RBAC manifests
make all             # Full build + test + generate
```

## When to Run

### After modifying CRD types:
Always run code generation to keep generated files in sync:
```bash
make generate        # Update generated Go code
make manifests       # Update CRD definitions and RBAC
```

### Example: Adding a field to a CRD type
1. Edit the type definition in `pkg/apis/v1alpha1/*.go`
2. Add kubebuilder markers if needed for validation/defaults
3. Run `make generate && make manifests`
4. Commit the generated files

## What Gets Generated

### Code Generation
- DeepCopy methods for all CRD types
- Type-safe client libraries
- Boilerplate headers on generated files

### Manifest Generation
- CRD YAML definitions from Go struct tags
- RBAC role and bindings for controller
- Webhook configurations

**Output**: `config/crd/bases/` directory

## Best Practices
- Always commit generated files
- Never manually edit generated files (marked with _generated.go suffix)
- Run as part of pre-commit workflow when CRDs change
- Part of CI/CD verification - manifests checked in

## Key Configuration
- **Tool**: controller-gen (via go tool)
- **Boilerplate header**: hack/boilerplate.go.txt
- **CRD options**: allowDangerousTypes=true (allows complex field types)
