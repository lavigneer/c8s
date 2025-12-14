# Devbox Development Environment

This project uses [Jetify Devbox](https://www.jetify.com/devbox) to provide a reproducible development environment with all necessary tools pre-installed.

## What is Devbox?

Devbox creates an isolated, reproducible development environment using Nix. It ensures all developers have the exact same tool versions, preventing "works on my machine" issues.

## Installed Packages

### Core Development Tools

#### `go_1_25@latest`
- **Purpose**: Go programming language compiler and toolchain
- **Why**: C8S backend is written in Go 1.25
- **Usage**: Compile binaries, run tests, format code

#### `gnumake@latest`
- **Purpose**: Build automation tool
- **Why**: Makefile orchestrates all build, test, and deployment operations
- **Usage**: `make build`, `make test`, `make deploy`

#### `git@latest`
- **Purpose**: Version control system
- **Why**: Required for git operations in pipelines and development
- **Usage**: Standard git commands

### Kubernetes Development

#### `kubectl@latest`
- **Purpose**: Kubernetes command-line tool
- **Why**: Interact with Kubernetes clusters, deploy C8S, manage resources
- **Usage**: `kubectl apply`, `kubectl get`, `kubectl logs`

#### `kind@latest`
- **Purpose**: Kubernetes IN Docker - local Kubernetes clusters
- **Why**: Create local test clusters for CI and manual testing
- **Usage**: `kind create cluster`, `kind delete cluster`

#### `tilt@latest`
- **Purpose**: Local Kubernetes development with live reload
- **Why**: Primary development workflow - builds, deploys, and live-reloads C8S
- **Usage**: `tilt up`, `tilt down`

#### `ctlptl@latest`
- **Purpose**: Control Panel tool for managing Tilt and local clusters
- **Why**: Simplifies cluster creation and Tilt state management
- **Usage**: `ctlptl create cluster kind`, `ctlptl get clusters`

### Code Quality

#### `golangci-lint@latest`
- **Purpose**: Go linter aggregator (runs multiple linters)
- **Why**: Enforce code quality, catch bugs, ensure consistency
- **Usage**: `make lint` or `golangci-lint run`

### Container Tools

#### `docker@latest`
- **Purpose**: Container runtime and image builder
- **Why**: Build Docker images, run containers, local development
- **Usage**: `docker build`, `docker run`, used by Tilt

### Frontend Development

#### `nodejs_24@latest`
- **Purpose**: Node.js JavaScript runtime
- **Why**: Required for Tailwind CSS builds and Playwright E2E tests
- **Usage**: `npm install`, `npm run build:css`, `npm run test:e2e`

### Optional Tools (Dog-fooding Workflow)

#### `ngrok@latest`
- **Purpose**: Secure tunnel to localhost for external access
- **Why**: Expose local C8S instance to GitHub webhooks for dog-fooding
- **Usage**: `ngrok start --all --config=./tilt/config/ngrok-config.yml`
- **Note**: Optional - only needed for GitHub Actions dog-fooding workflow

## Getting Started

### First Time Setup

1. **Install Devbox**:
   ```bash
   curl -fsSL https://get.jetify.com/devbox | bash
   ```

2. **Enter Development Environment**:
   ```bash
   cd c8s
   devbox shell
   ```

   All tools will be automatically available in your PATH.

3. **Start Development**:
   ```bash
   tilt up  # Primary development workflow
   ```

### Using Devbox

#### Enter Shell
```bash
devbox shell
```
Opens a new shell with all tools available.

#### Run Single Command
```bash
devbox run -- make test
```
Runs a command in the devbox environment without entering the shell.

#### Using Script Aliases
Devbox provides convenient aliases for common commands:

```bash
devbox run -- test              # make test
devbox run -- build             # make build
devbox run -- tilt:up           # make tilt-up
devbox run -- test:e2e          # make test-e2e
```

See all available aliases: `devbox run -- help`

## Environment Variables

Devbox automatically sets these environment variables:

- `KUBEBUILDER_ASSETS`: Path to Kubernetes envtest binaries for integration tests
- `GO111MODULE=on`: Enable Go modules (default in Go 1.17+)
- `CGO_ENABLED=0`: Disable CGO for static binary builds

## Updating Tools

To update all tools to their latest versions:

```bash
devbox update
```

To update a specific tool:

```bash
devbox add go_1_25@latest --force
```

## Troubleshooting

### "Command not found" errors
Make sure you're in a devbox shell:
```bash
devbox shell
```

### Tools not updating
Clear devbox cache and re-enter:
```bash
exit  # Exit devbox shell
devbox cache clear
devbox shell
```

### Nix store disk space
Devbox uses Nix under the hood. To clean up:
```bash
nix-collect-garbage
```

## Alternatives to Devbox

If you prefer not to use Devbox, you can install tools manually:

### macOS (Homebrew)
```bash
brew install go golangci-lint kubectl kind docker git make tilt nodejs ngrok
```

### Linux (Ubuntu/Debian)
```bash
# Install Go 1.25 from official site
# Install other tools via package manager or from releases
apt-get install kubectl docker git make nodejs
```

**Note**: Manual installation doesn't guarantee version consistency across developers.

## Learn More

- [Devbox Documentation](https://www.jetify.com/docs/devbox)
- [Devbox GitHub](https://github.com/jetify-com/devbox)
- [Nix Package Search](https://search.nixos.org/packages)

## Files

- `devbox.json`: Package list and configuration
- `devbox.lock`: Lock file with exact package versions (similar to package-lock.json)
- `.envrc`: direnv integration for automatic environment activation
