# C8S Dog-Fooding: Using C8S to Test C8S

This guide explains how to configure C8S to use itself for continuous integration ("dog-fooding"). This allows GitHub Actions to trigger your local C8S instance to run the full CI pipeline.

## Overview

"Dog-fooding" means eating your own dog food - using your product to test itself. For C8S, this means:

1. **Local Development**: You run C8S locally with `tilt up`
2. **Public Tunnel**: ngrok exposes your local services to the internet
3. **GitHub Hook**: GitHub Actions triggers webhook on your local C8S
4. **Pipeline Execution**: Your local C8S runs the full CI pipeline
5. **Results**: Pipeline results appear in your C8S dashboard

```
┌─────────────────┐
│ GitHub Actions  │
└────────┬────────┘
         │ (push event)
         ↓
    ┌────────────┐
    │ ngrok URL  │ (tunnel to local machine)
    └────────┬───┘
             ↓
    ┌───────────────────┐
    │ C8S Webhook (9443)│
    └────────┬──────────┘
             ↓
    ┌────────────────────────┐
    │ Create PipelineRun CRD │
    └────────┬───────────────┘
             ↓
    ┌────────────────────────────┐
    │ C8S Controller             │
    │ - Schedules Jobs           │
    │ - Runs Pipeline Steps      │
    │ - Updates Status           │
    └────────┬───────────────────┘
             ↓
    ┌────────────────────────┐
    │ C8S API Dashboard      │
    │ View Results           │
    └────────────────────────┘
```

## Prerequisites

### Local Setup
- C8S running locally: `tilt up` (see [tilt-setup.md](./tilt-setup.md))
- ngrok installed and authenticated
- GitHub repository with push/PR access

### GitHub Setup
- Administrator access to repository settings
- Ability to create/modify secrets

## Step 1: Prepare Local C8S Instance

### Start Tilt
```bash
cd ~/workspace/c8s
tilt up
```

Wait for all components to be healthy:
- Controller pod ready
- Webhook pod ready
- API Server pod ready

Check status in Tilt dashboard: http://localhost:10350

### Verify C8S Components

```bash
# Check all pods are running
kubectl get pods -n c8s-system

# Verify services
kubectl get svc -n c8s-system

# Check logs
kubectl logs -n c8s-system -l app=c8s-controller -f
```

## Step 2: Create ngrok Tunnels

### Option A: Via Tilt UI (Recommended)

1. Open Tilt dashboard: http://localhost:10350
2. Find `c8s-webhook` resource
3. Click the "ngrok" button
4. Copy the HTTPS URL (e.g., `https://abc123.ngrok.io`)
5. Find `c8s-api-server` resource
6. Click the "ngrok" button
7. Copy the HTTP URL (e.g., `http://def456.ngrok.io`)

### Option B: Manual ngrok

```bash
# Terminal 1: Webhook tunnel (HTTPS)
ngrok http --proto=https 9443

# Terminal 2: API server tunnel (HTTP)
ngrok http 8080

# Terminal 3: View all tunnels
open http://localhost:4040
```

### Verify Tunnels

```bash
# Check ngrok dashboard
open http://localhost:4040

# Test webhook endpoint
curl -v https://YOUR_WEBHOOK_URL/health

# Test API endpoint
curl http://YOUR_API_URL/api/v1/health
```

## Step 3: Create C8S Pipeline Definition

The `.c8s.yaml` file defines your CI pipeline:

```yaml
version: v1alpha1
name: c8s-main
steps:
  - name: lint
    image: golangci/golangci-lint:latest
    commands:
      - golangci-lint run ./cmd/... ./pkg/...

  - name: test
    image: golang:1.25
    commands:
      - go test -v ./...
    dependsOn: [lint]

  - name: build
    image: golang:1.25
    commands:
      - CGO_ENABLED=0 go build -o bin/controller ./cmd/controller
    dependsOn: [test]
```

This file is already provided in the repository at `.c8s.yaml`.

## Step 4: Configure GitHub Secrets

### Add Repository Secrets

1. Go to GitHub repo: **Settings → Secrets and variables → Actions**
2. Create new secret: **C8S_WEBHOOK_URL**
   - Value: `https://YOUR_WEBHOOK_NGROK_URL` (HTTPS!)
   - Example: `https://abc123.ngrok.io`
3. Create new secret: **C8S_API_URL**
   - Value: `http://YOUR_API_NGROK_URL` (HTTP)
   - Example: `http://def456.ngrok.io`

⚠️ **Important**: Keep these URLs secret - they expose your local machine to the internet!

### Verify Secrets

```bash
# GitHub CLI (if installed)
gh secret list

# Or check in GitHub UI
# Settings → Secrets and variables → Actions
```

## Step 5: Trigger Dog-Fooding Workflow

### Option A: Push to Repository

```bash
# Make a change and push
git add .
git commit -m "Test C8S dog-fooding"
git push origin main
```

This will automatically trigger the `.github/workflows/c8s-dogfood.yml` workflow.

### Option B: Manual Workflow Trigger

1. Go to GitHub repo: **Actions**
2. Select **C8S Dog-Fooding** workflow
3. Click **Run workflow**
4. Select branch and click **Run workflow**

### Option C: Direct Webhook Trigger

```bash
WEBHOOK_URL="https://YOUR_WEBHOOK_URL"

curl -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Event: push" \
  -d '{
    "apiVersion": "c8s.dev/v1alpha1",
    "kind": "PipelineRun",
    "metadata": {
      "name": "c8s-test-run",
      "namespace": "c8s-system"
    },
    "spec": {
      "pipelineRef": { "name": "c8s-main" },
      "commit": "abc123def",
      "branch": "main"
    }
  }'
```

## Step 6: Monitor Pipeline Execution

### View in C8S Dashboard

1. Open API server: http://localhost:8080
2. Navigate to pipelines section
3. Find `c8s-gh-{run-id}-{run-number}` pipeline
4. Click to view:
   - Pipeline status
   - Step-by-step execution
   - Real-time logs
   - Artifacts

### View in Tilt Dashboard

1. Open Tilt: http://localhost:10350
2. Click **c8s-webhook** to see webhook events
3. Click **c8s-controller** to see job scheduling
4. Watch logs in real-time

### View in GitHub Actions

1. Go to GitHub repo: **Actions**
2. Select **C8S Dog-Fooding** workflow
3. View workflow execution status
4. Check logs in each step
5. View PR comments (if triggered from PR)

### View ngrok Traffic

1. Open ngrok dashboard: http://localhost:4040
2. See all HTTP requests to your local services
3. Inspect request/response details
4. Monitor bandwidth and connections

## Understanding the Workflow

### Workflow Steps

1. **prepare-dogfood-test**
   - Reads GitHub secrets (C8S_WEBHOOK_URL, C8S_API_URL)
   - Displays setup instructions if secrets missing
   - Outputs URLs for next steps

2. **trigger-c8s-pipeline**
   - Creates PipelineRun with GitHub metadata
   - POSTs to C8S webhook via ngrok
   - Polls pipeline status
   - Waits up to 40 minutes for completion
   - Fetches logs
   - Comments on PR (if applicable)

### Pipeline Execution

The `.c8s.yaml` pipeline runs these steps:

1. **lint** - Code quality checks (golangci-lint)
2. **test-unit** - Unit tests with coverage
3. **test-integration** - Integration tests
4. **build** - Compile binaries (amd64)
5. **verify-manifests** - CRD manifest validation
6. **build-docker-images** - Docker image builds
7. **test-e2e-preparation** - E2E test setup
8. **test-report** - Summary report

Steps run in parallel where possible, with dependencies enforced.

## Troubleshooting

### Workflow doesn't trigger

**Problem**: GitHub Actions workflow doesn't start
- Check workflow file: `.github/workflows/c8s-dogfood.yml` exists
- Check branch: workflow triggers on `push` to `main`, `develop`, `release-*`
- Check paths: only runs if files in `cmd/`, `pkg/`, `.c8s.yaml` changed

**Solution**:
```bash
# Force workflow trigger by changing a covered path
touch cmd/api-server/.touch
git add .
git commit -m "Trigger workflow"
git push
```

### Workflow runs but skips pipeline trigger

**Problem**: Job `trigger-c8s-pipeline` is skipped
- Secrets not configured
- No ngrok URLs set

**Solution**:
1. Add secrets: Settings → Secrets and variables → Actions
2. Verify secret names match: `C8S_WEBHOOK_URL`, `C8S_API_URL`
3. Re-run workflow

### Webhook receives "Connection refused"

**Problem**: HTTP 500 or connection error when triggering webhook
- ngrok tunnel not active
- Webhook pod not running
- Webhook listening on wrong port

**Solution**:
```bash
# Check ngrok is running
ps aux | grep ngrok

# Check webhook pod
kubectl get pod -n c8s-system -l app=c8s-webhook

# Check webhook logs
kubectl logs -n c8s-system -l app=c8s-webhook -f

# Restart webhook
kubectl rollout restart deployment/c8s-webhook -n c8s-system

# Test connection
curl -v https://YOUR_NGROK_URL/health
```

### Pipeline triggers but doesn't start

**Problem**: PipelineRun created but no pods scheduled
- Controller pod not running
- CRDs not installed
- RBAC permissions missing

**Solution**:
```bash
# Check controller pod
kubectl get pod -n c8s-system -l app=c8s-controller

# Check CRDs installed
kubectl api-resources | grep pipelinerun

# Check RBAC
kubectl describe serviceaccount c8s-controller -n c8s-system

# Check controller logs
kubectl logs -n c8s-system -l app=c8s-controller -f
```

### Pipeline runs slowly

**Problem**: Pipeline steps take longer than expected
- Resource constraints
- Slow network
- ngrok rate limiting

**Solution**:
```bash
# Check pod resource usage
kubectl top pods -n c8s-system

# Monitor ngrok dashboard for rate limits
open http://localhost:4040

# Check node resources
kubectl top nodes

# Increase timeouts in .c8s.yaml
```

### Can't get pipeline logs

**Problem**: "Could not retrieve logs from C8S API"
- API server not responding
- Wrong API URL in secret
- Logs not available yet

**Solution**:
```bash
# Test API endpoint directly
curl http://YOUR_API_URL/api/v1/health

# Check API server logs
kubectl logs -n c8s-system -l app=c8s-api-server -f

# Check ngrok is forwarding
open http://localhost:4040

# Try manual log fetch
PIPELINE_RUN="c8s-gh-{run-id}-{run-number}"
curl http://YOUR_API_URL/api/v1/pipelines/$PIPELINE_RUN/logs
```

### ngrok tunnel keeps disconnecting

**Problem**: "Connection refused" errors intermittently
- Free ngrok tier has connection limits
- Tunnel URL expired (2-hour limit)
- Network instability

**Solution**:
1. **Free tier**: Upgrade to Pro for stable tunnels
2. **Free tier workaround**:
   - Use `ngrok config` to set custom subdomain (if available on plan)
   - Re-create tunnel and update secrets
3. **Keep Tilt running**:
   - ngrok extension will auto-recreate tunnels
   - Check Tilt logs: `tilt logs ngrok:status`

### PR comment doesn't appear

**Problem**: Workflow runs but no comment on PR
- `pull_request` event not triggered
- GitHub token permissions
- Script error

**Solution**:
```bash
# Check workflow file specifies pull_request trigger
grep -A2 "on:" .github/workflows/c8s-dogfood.yml

# Check GitHub token has write permission
# Settings → Actions → General → Workflow permissions

# Check workflow logs for script errors
# Actions → workflow run → trigger-c8s-pipeline → Comment on PR
```

## Advanced Configuration

### Custom Pipeline Parameters

Pass parameters via webhook payload:

```bash
curl -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "spec": {
      "timeout": "60m",
      "retryPolicy": {
        "maxRetries": 3,
        "backoffSeconds": 60
      }
    }
  }'
```

### Environment Variables

Configure environment for pipeline steps:

```yaml
# In .c8s.yaml
steps:
  - name: build
    image: golang:1.25
    commands:
      - go build -o bin/controller ./cmd/controller
    env:
      - name: CGO_ENABLED
        value: "0"
      - name: GOOS
        value: "linux"
```

### Secrets in Pipeline

Reference Kubernetes secrets:

```yaml
steps:
  - name: deploy
    image: ubuntu:latest
    commands:
      - ./deploy.sh --token=$DEPLOY_TOKEN
    secrets:
      - secretRef: deploy-credentials
        key: TOKEN
        envVar: DEPLOY_TOKEN
```

### Artifacts

Store build artifacts:

```yaml
steps:
  - name: build
    image: golang:1.25
    commands:
      - go build -o bin/controller ./cmd/controller
    artifacts:
      - bin/controller
      - bin/webhook
```

## Best Practices

### 1. Keep Tilt Running
- Don't stop Tilt between tests
- Changes auto-rebuild
- Speeds up iteration

### 2. Monitor Resources
```bash
# Watch resource usage
kubectl top pods -n c8s-system --watch
```

### 3. Log Retention
- Logs stored in S3/object storage (if configured)
- Check retention policy
- Archive important logs

### 4. Security
- Keep ngrok URLs in GitHub secrets
- Don't commit URLs to repository
- Use strong webhook secrets if implementing
- Rotate secrets periodically

### 5. Testing
- Test locally first
- Run manual workflow trigger before PR
- Monitor first few production runs
- Gradually increase pipeline complexity

## Next Steps

1. **Set up alerts**: Configure GitHub Actions notifications
2. **Add more steps**: Expand pipeline as project grows
3. **Integrate with dashboard**: Add C8S dashboard to CI reports
4. **Custom steps**: Add project-specific validation
5. **Multi-branch**: Configure different pipelines per branch

## See Also

- [Tilt Setup](./tilt-setup.md) - Local development environment
- [.c8s.yaml](./.c8s.yaml) - Pipeline definition
- [GitHub Actions Workflow](./.github/workflows/c8s-dogfood.yml) - CI workflow
- [C8S Documentation](../README.md) - General documentation
