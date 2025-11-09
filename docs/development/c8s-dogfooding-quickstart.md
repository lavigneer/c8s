# C8S Dog-Fooding Quick Start Guide

Use GitHub Actions to trigger your local C8S via ngrok

---

## Step 1: Start Local C8S

```bash
tilt up
```

Wait for all pods to be ready:
- ✓ c8s-controller
- ✓ c8s-webhook
- ✓ c8s-api-server

Access: http://localhost:10350

---

## Step 2: Create ngrok Tunnels

### Option A: Via Tilt UI (Recommended)

1. Open http://localhost:10350
2. Click "ngrok" button on c8s-webhook
3. Copy HTTPS URL (e.g., `https://abc123.ngrok.io`)
4. Click "ngrok" button on c8s-api-server
5. Copy HTTP URL (e.g., `http://def456.ngrok.io`)
6. View all tunnels: http://localhost:4040

### Option B: Manual ngrok

```bash
# Terminal 1: Webhook tunnel
ngrok http --proto=https 9443

# Terminal 2: API server tunnel
ngrok http 8080

# Terminal 3: View tunnels
open http://localhost:4040
```

---

## Step 3: Configure GitHub Secrets

1. Go to GitHub repo → **Settings** → **Secrets and variables** → **Actions**
2. Click **New repository secret**

### Secret 1: C8S_WEBHOOK_URL
- **Name:** `C8S_WEBHOOK_URL`
- **Value:** `https://YOUR_WEBHOOK_NGROK_URL` (HTTPS!)
- **Example:** `https://abc123.ngrok.io`

### Secret 2: C8S_API_URL
- **Name:** `C8S_API_URL`
- **Value:** `http://YOUR_API_NGROK_URL` (HTTP)
- **Example:** `http://def456.ngrok.io`

---

## Step 4: Trigger Workflow

### Option A: Automatic (Push/PR)

```bash
git push origin main
```

Workflow automatically triggers on code changes

### Option B: Manual

1. Go to GitHub repo → **Actions**
2. Select **"C8S Dog-Fooding"** workflow
3. Click **"Run workflow"**

---

## Step 5: Monitor Pipeline

### C8S Dashboard
- **URL:** http://localhost:8080
- **View:** Pipeline status, logs, artifacts

### Tilt Dashboard
- **URL:** http://localhost:10350
- **View:** Controller scheduling jobs, webhook events

### GitHub Actions
- **Location:** GitHub repo → **Actions** → **"C8S Dog-Fooding"**
- **View:** Workflow status, PR comments

### ngrok Dashboard
- **URL:** http://localhost:4040
- **View:** Tunnel traffic, request/response details

---

## Troubleshooting

### ✗ Workflow doesn't run

- Check secrets are configured: Settings → Secrets and variables → Actions
- Verify branch matches trigger: `main`, `develop`, `release-*`
- Ensure changes touch: `cmd/`, `pkg/`, `.c8s.yaml`

### ✗ Webhook timeout

- Verify ngrok tunnel is active: http://localhost:4040
- Check webhook pod: `kubectl get pod -n c8s-system`
- Test connection: `curl -v https://YOUR_WEBHOOK_URL/health`

### ✗ Pipeline doesn't start

- Check controller: `kubectl get pod -n c8s-system`
- Verify CRDs: `kubectl api-resources | grep pipelinerun`
- Check logs: `kubectl logs -n c8s-system -l app=c8s-controller`

### ✗ Can't fetch logs

- Verify API URL in secret is correct
- Check API server: `curl http://YOUR_API_URL/api/v1/health`
- View API logs: `kubectl logs -n c8s-system -l app=c8s-api-server`

---

## Useful Commands

```bash
# Check all C8S pods
kubectl get pods -n c8s-system

# View controller logs
kubectl logs -n c8s-system -l app=c8s-controller -f

# View webhook logs
kubectl logs -n c8s-system -l app=c8s-webhook -f

# View API server logs
kubectl logs -n c8s-system -l app=c8s-api-server -f

# List pipelines
kubectl get pipelinerun -n c8s-system

# Get pipeline details
kubectl describe pipelinerun <name> -n c8s-system

# Watch ngrok traffic
open http://localhost:4040

# Manual webhook trigger
curl -X POST "https://YOUR_WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "apiVersion":"c8s.dev/v1alpha1",
    "kind":"PipelineRun",
    "metadata":{"name":"test-run"}
  }'
```

---

## Architecture

```
GitHub Actions (internet)
        ↓
    ngrok tunnel (HTTPS)
        ↓
    Local Machine
        ↓
    c8s-webhook:9443 (receives push event)
        ↓
    Creates PipelineRun CRD
        ↓
    c8s-controller (watches for new PipelineRuns)
        ↓
    Schedules Kubernetes Jobs
        ↓
    Jobs run in isolated Pods
        ↓
    c8s-api-server (streams logs)
        ↓
    GitHub Actions polls status
        ↓
    Results appear in PR comments
```

---

## Documentation

| Resource | Location |
|----------|----------|
| Full Guide | `docs/development/c8s-dogfooding.md` |
| Tilt Setup | `docs/development/tilt-setup.md` |
| Pipeline Definition | `.c8s.yaml` |
| Workflow Definition | `.github/workflows/c8s-dogfood.yml` |

---

## Next Steps

1. ✓ Set up local C8S (`tilt up`)
2. ✓ Create ngrok tunnels
3. ✓ Add GitHub secrets
4. → Trigger first workflow
5. → Monitor pipeline execution
6. → Iterate on pipeline definition

---

## For More Help

See [`docs/development/c8s-dogfooding.md`](./c8s-dogfooding.md) for:
- Detailed step-by-step guide
- Troubleshooting 12 common issues
- Advanced configuration
- Best practices
- Security considerations
