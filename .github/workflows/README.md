# GitHub Actions Workflows

CI/CD pipeline documentation for C8S.

## Overview

C8S uses 5 GitHub Actions workflows for continuous integration, testing, and deployment:

| Workflow | Purpose | Trigger | Runtime |
|----------|---------|---------|---------|
| **ci.yaml** | Lint, test, build | Every PR/push to main | ~5 min |
| **e2e-tests.yml** | E2E tests (Playwright) | Frontend changes | ~15 min |
| **tilt-ci.yml** | Full integration testing | PR/push to main/develop | ~10 min |
| **build-and-push.yml** | Docker images to GHCR | Push to main, tags | ~8 min |
| **c8s-dogfood.yml** | C8S tests itself | PR/push (optional) | ~15 min |

**Total CI/CD Time**: ~20 minutes (workflows run in parallel)

## Workflow Details

### 1. ci.yaml - Main CI Pipeline

**Purpose**: Primary continuous integration - validates code quality and functionality

**Triggers**:
- Push to `main` or `release-*` branches
- Pull requests to `main` or `release-*`

**Jobs** (sequential):

1. **lint** (2 min)
   - Runs golangci-lint on all Go code
   - Uses devbox for consistent environment
   - Checks: formatting, vet, staticcheck, unused code, etc.

2. **test** (3 min)
   - Unit tests: `make test-unit`
   - Integration tests: `make test-integration`
   - Uploads coverage to Codecov
   - Uses devbox for Go 1.25 and envtest

3. **build** (2 min)
   - Compiles all binaries (controller, webhook, CLI)
   - Verifies CRD manifests are up to date
   - Ensures no uncommitted generated code

4. **docker** (depends on: lint, test, build)
   - Builds Docker images for controller and webhook
   - Validates Dockerfile correctness
   - Does NOT push images (see build-and-push.yml)

**Status**: ✅ Must pass for PR merge

**Badge**: ![CI](https://github.com/lavigneer/c8s/workflows/CI/badge.svg)

### 2. e2e-tests.yml - End-to-End Testing

**Purpose**: Full browser automation testing with Playwright

**Triggers**:
- Changes to `tests/e2e/**`, `cmd/api-server/**`, `playwright.config.ts`, `package.json`

**Strategy**: Matrix testing
- **Browsers**: Chromium, Firefox
- **Viewports**: Desktop (1920x1080), Tablet (1024x1366), Mobile (390x844)
- **Total combinations**: 2 browsers × 3 viewports = 6 test runs

**Jobs**:

1. **e2e** (matrix: browser × viewport)
   - Installs devbox and npm dependencies
   - Installs Playwright browsers
   - Builds Tailwind CSS
   - Runs E2E test suite (120+ tests)
   - Uploads test reports and videos on failure

**Artifacts** (retained):
- Playwright HTML reports: 30 days
- Test results JSON: 15 days
- Failure videos: 7 days

**Status**: ✅ Must pass for PR merge (frontend changes)

### 3. tilt-ci.yml - Tilt Integration Testing

**Purpose**: Validates full deployment with Tilt in kind cluster

**Triggers**:
- Push to `main` or `develop` branches
- Pull requests

**Jobs**:

1. **tilt-ci** (15 min)
   - Creates temporary kind cluster with port mappings (80→8080, 443→8443)
   - Installs CRDs
   - Runs `tilt ci` (non-interactive Tilt run)
   - Validates all resources healthy
   - Captures logs and cluster info on failure

**Outputs on Failure**:
- Tilt logs
- kubectl cluster info
- Pod descriptions
- Event logs

**Status**: ✅ Must pass for PR merge

**Use Case**: Validates that Tiltfile and Helm chart work correctly

### 4. build-and-push.yml - Docker Image Publishing

**Purpose**: Builds and publishes Docker images to GitHub Container Registry (GHCR)

**Triggers**:
- Push to `main` or `develop` branches
- Git tags matching `v*` (e.g., v0.1.0)
- Pull requests (build only, no push)

**Strategy**: Matrix build
- Components: api-server, controller, webhook

**Jobs**:

1. **build-and-push** (matrix: component)
   - Builds Docker image for each component
   - Tags images:
     - `latest` (main branch)
     - `develop` (develop branch)
     - `v1.2.3` (git tags)
     - `sha-abc123` (commit SHA)
   - Pushes to `ghcr.io/lavigneer/c8s-<component>`

2. **publish-helm-chart** (on git tags only)
   - Packages Helm chart
   - Pushes to GHCR OCI registry
   - Creates GitHub Release with artifacts

**Registry**: ghcr.io/lavigneer

**Status**: Informational (doesn't block PRs)

### 5. c8s-dogfood.yml - Self-Hosting / Dogfooding

**Purpose**: C8S tests itself - runs C8S pipeline using C8S

**Triggers**:
- Push to `main`, `develop`, or `release-*` branches
- Pull requests

**Prerequisites**:
- Requires `C8S_BASE_URL` secret configured
- Requires ngrok tunnel to local C8S instance
- Optional workflow (skips if C8S_BASE_URL not set)

**Jobs**:

1. **dogfood** (15 min)
   - Creates GitHub webhook payload
   - Sends webhook to C8S instance
   - Polls pipeline status (max 40 minutes)
   - Fetches and displays pipeline logs
   - Comments on PR with pipeline results

**Workflow**:
```
GitHub Event → Webhook → C8S → Creates PipelineRun
                                     ↓
                              Run Steps as K8s Jobs
                                     ↓
                              Update Status → API
                                     ↓
                              GitHub Action polls status
                                     ↓
                              Comment on PR with results
```

**Status**: Optional (doesn't block PRs if disabled)

**Setup**: See [docs/development/c8s-dogfooding.md](../../docs/development/c8s-dogfooding.md)

## Workflow Dependencies

```
On PR Creation:
├─ ci.yaml (lint → test → build → docker) [REQUIRED]
├─ e2e-tests.yml (if frontend changed) [REQUIRED]
├─ tilt-ci.yml [REQUIRED]
└─ c8s-dogfood.yml (if C8S_BASE_URL set) [OPTIONAL]

On Push to main:
├─ ci.yaml [REQUIRED]
├─ tilt-ci.yml [REQUIRED]
├─ build-and-push.yml (builds + pushes images) [INFORMATIONAL]
└─ c8s-dogfood.yml [OPTIONAL]

On Git Tag (v*):
└─ build-and-push.yml (publishes release) [RELEASE]
```

## Status Badges

Add to README.md:

```markdown
![CI](https://github.com/lavigneer/c8s/workflows/CI/badge.svg)
![E2E Tests](https://github.com/lavigneer/c8s/workflows/E2E%20Tests/badge.svg)
![Tilt CI](https://github.com/lavigneer/c8s/workflows/Tilt%20CI/badge.svg)
![Build and Push](https://github.com/lavigneer/c8s/workflows/Build%20and%20Push/badge.svg)
```

## Secrets Required

| Secret | Used By | Purpose |
|--------|---------|---------|
| `GITHUB_TOKEN` | All workflows | Automatic (GitHub provides) |
| `CODECOV_TOKEN` | ci.yaml | Upload coverage reports |
| `C8S_BASE_URL` | c8s-dogfood.yml | URL of C8S instance for dog-fooding |
| `C8S_WEBHOOK_SECRET` | c8s-dogfood.yml | Webhook signature secret |

## Debugging Workflows

### View Workflow Runs

```bash
# List recent runs
gh run list

# View specific run
gh run view <run-id>

# View logs
gh run view <run-id> --log

# Re-run failed jobs
gh run rerun <run-id> --failed
```

### Common Failures

#### ci.yaml Failures

**Lint failures**:
```bash
# Run locally first
make lint
golangci-lint run
```

**Test failures**:
```bash
# Run locally
make test
make test-integration
```

**Build failures**:
```bash
# Check generated files are committed
make manifests
git status  # Should show no changes
```

#### e2e-tests.yml Failures

**Playwright failures**:
```bash
# Run locally
npm run test:e2e

# Debug specific test
npm run test:e2e:debug tests/e2e/specs/authentication.spec.ts

# View report
npm run test:e2e:report
```

**Check artifacts**:
- Download Playwright report from GitHub Actions
- Watch failure videos

#### tilt-ci.yml Failures

**Tilt CI failures**:
```bash
# Run Tilt CI locally
make tilt-ci-local

# Or manually
kind create cluster --name c8s-ci
tilt ci
```

**Check logs**:
- Tilt logs artifact in GitHub Actions
- kubectl cluster info artifact

#### build-and-push.yml Failures

**Image build failures**:
```bash
# Build locally
make docker-build

# Check Dockerfile
docker build --target controller -t test .
```

## Performance Optimization

### Caching Strategy

All workflows use caching:
- **Go modules**: `~/.cache/go-build`, `~/go/pkg/mod`
- **npm packages**: `~/.npm`
- **Devbox**: `~/.cache/devbox`
- **Docker layers**: GitHub Actions cache

### Parallel Execution

Workflows run in parallel:
- ci.yaml and tilt-ci.yml run simultaneously
- e2e-tests.yml runs if triggered
- build-and-push.yml runs independently

### Runtime Targets

| Workflow | Target | Current | Status |
|----------|--------|---------|--------|
| ci.yaml | <5 min | ~5 min | ✅ |
| e2e-tests.yml | <15 min | ~15 min | ✅ |
| tilt-ci.yml | <10 min | ~10 min | ✅ |
| build-and-push.yml | <8 min | ~8 min | ✅ |

## Adding New Workflows

When adding a new workflow:

1. **Create workflow file**: `.github/workflows/new-workflow.yml`
2. **Test locally first**: Use `act` for local testing
3. **Add to this README**: Document purpose, triggers, jobs
4. **Add status badge**: If user-facing workflow
5. **Update required checks**: In GitHub repo settings

## Workflow Best Practices

1. **Use devbox for consistency**: All workflows use devbox shell
2. **Cache aggressively**: Cache go modules, npm, Docker layers
3. **Fail fast**: Run quick checks (lint) before slow ones (E2E)
4. **Parallel when possible**: Use matrix strategies
5. **Upload artifacts**: Save logs, reports, videos for debugging
6. **Clear job names**: Descriptive names aid debugging

## Local Testing

Test workflows locally before pushing:

```bash
# Install act (GitHub Actions local runner)
brew install act

# Run specific workflow
act -W .github/workflows/ci.yaml

# Run specific job
act -j lint

# Run with secrets
act --secret-file .secrets
```

## Maintenance

### Regular Tasks

- **Weekly**: Review failed workflows, optimize slow jobs
- **Monthly**: Update action versions (@v3 → @v4)
- **Quarterly**: Review workflow dependencies, remove unused

### Updating Actions

```yaml
# Before
uses: actions/checkout@v3

# After
uses: actions/checkout@v4
```

Check for updates: [GitHub Actions Updates](https://github.com/actions)

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Workflow Syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [Action Marketplace](https://github.com/marketplace?type=actions)

---

**Questions?** See [ci.yaml](.github/workflows/ci.yaml) for examples.
