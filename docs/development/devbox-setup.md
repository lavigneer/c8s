# Devbox Setup Guide

C8S uses [Devbox](https://www.jetify.com/devbox) to manage development dependencies in a reproducible way.

## What is Devbox?

Devbox is a portable, isolated development environment manager that uses Nix packages under the hood. It ensures all developers have the same versions of tools without polluting your system.

## Prerequisites

1. **Install Devbox**:
   ```bash
   curl -fsSL https://get.jetify.com/devbox | bash
   ```

2. **Optional: Install direnv** (for automatic environment loading):
   ```bash
   # macOS
   brew install direnv

   # Linux
   curl -sfL https://direnv.net/install.sh | bash
   ```

   Add to your shell RC file (~/.bashrc, ~/.zshrc):
   ```bash
   eval "$(direnv hook bash)"  # or zsh
   ```

## Getting Started

### Option 1: Manual Shell Activation

```bash
# Enter the development environment
devbox shell

# You'll see a welcome message with available commands
# All tools (Go, kubectl, kind, make, etc.) are now available
```

### Option 2: Automatic Activation with direnv

```bash
# Allow direnv for this directory (one-time setup)
direnv allow

# Now the environment activates automatically when you cd into the project
cd /path/to/c8s  # Environment loads automatically
```

## What's Included?

Devbox provides the following tools:

- **go_1_25**: Go 1.25 compiler
- **kubectl**: Kubernetes CLI
- **kind**: Kubernetes in Docker (for local testing)
- **docker**: Docker CLI
- **git**: Git version control
- **gnumake**: GNU Make

Additionally, the init hook automatically installs Go tools:
- **controller-gen**: CRD and RBAC manifest generation
- **setup-envtest**: Kubernetes API server for integration tests
- **golangci-lint**: Go linter

## Quick Commands

Once in the devbox shell, you can use these shortcut commands:

```bash
# Run tests
devbox run test

# Build binaries
devbox run build

# Lint code
devbox run lint

# Generate CRDs and code
devbox run generate

# Install CRDs to cluster
devbox run install-crds

# Run controller locally
devbox run run-controller
```

Or use make directly:

```bash
make help     # Show all available targets
make test     # Run tests
make build    # Build binaries
```

## Environment Variables

Devbox sets the following environment variables:

- `KUBEBUILDER_ASSETS`: Path to kubebuilder test binaries
- `GO111MODULE`: Enabled for Go modules
- `CGO_ENABLED`: Disabled for static binaries
- `GOPATH`: Set to $HOME/go
- `PATH`: Includes $GOPATH/bin

## Updating Dependencies

To add new tools:

1. Edit `devbox.json`:
   ```json
   {
     "packages": [
       "go_1_25@latest",
       "your-new-tool@version"
     ]
   }
   ```

2. Update the environment:
   ```bash
   devbox update
   ```

## Troubleshooting

### "devbox: command not found"

Install devbox:
```bash
curl -fsSL https://get.jetify.com/devbox | bash
```

### Go tools not found

Exit and re-enter the shell:
```bash
exit
devbox shell
```

The init hook will reinstall missing tools.

### Permission denied on .envrc

Allow direnv:
```bash
direnv allow
```

### Kind cluster not accessible

Ensure Docker is running:
```bash
docker ps
```

## VS Code Integration

Add to `.vscode/settings.json`:

```json
{
  "go.goroot": "${workspaceFolder}/.devbox/nix/profile/default/share/go",
  "go.gopath": "${env:HOME}/go",
  "terminal.integrated.env.linux": {
    "DEVBOX_SHELL": "1"
  },
  "terminal.integrated.env.osx": {
    "DEVBOX_SHELL": "1"
  }
}
```

## Benefits

✅ **Reproducible**: Same tool versions across all developers
✅ **Isolated**: Doesn't conflict with system-installed tools
✅ **Fast**: Nix caching makes setup near-instant
✅ **Cross-platform**: Works on macOS, Linux, and WSL
✅ **No containers**: Unlike devcontainers, no Docker overhead
✅ **Declarative**: `devbox.json` is version-controlled

## Learn More

- Devbox docs: https://www.jetify.com/devbox/docs
- Nix packages: https://search.nixos.org/packages
- direnv: https://direnv.net/
