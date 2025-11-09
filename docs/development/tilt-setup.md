# Tilt Local Development Setup for C8S

Welcome to the C8S development environment! This guide helps you get started with Tilt for rapid local Kubernetes development.

## 🚀 Quick Start (5 minutes)

### Prerequisites
- Docker running (`docker ps` works)
- Go 1.25+ installed
- kubectl available
- k3d 5.8.3+ installed
- Tilt 0.33.0+ installed
- ngrok installed (for GitHub Actions testing) - `brew install ngrok` on macOS or download from https://ngrok.com/download

### One-Command Setup

```bash
# Start development environment
tilt up

# That's it! Tilt will:
# 1. Create a local k3d cluster
# 2. Install CRDs and configuration
# 3. Build Docker images for all components
# 4. Deploy controller, api-server, and webhook
# 5. Open dashboard at http://localhost:10350
```

### First Iteration

```bash
# Edit any Go file
vim cmd/controller/main.go

# Save the file - Tilt automatically:
# 1. Detects the change
# 2. Rebuilds the component
# 3. Updates the container image
# 4. Restarts the pod

# Watch progress in Tilt dashboard (http://localhost:10350)
```

## 📚 Documentation

- **[Quick Start Guide](docs/tilt-setup.md)**: Comprehensive setup and workflow guide
- **[Data Model](specs/003-implement-tilt-or/data-model.md)**: State management and entities
- **[API Contract](specs/003-implement-tilt-or/contracts/tiltfile-spec.md)**: Tiltfile configuration options

## 🔧 Component Endpoints

| Component | Local Port | Purpose | Public (ngrok) |
|-----------|-----------|---------|-----------------|
| API Server | 8080 | REST API / Web Dashboard | HTTP tunnel |
| Webhook | 9443 | Git webhook endpoint | HTTPS tunnel |
| Controller (Debug) | 6060 | pprof profiling | Local only |
| Tilt Dashboard | 10350 | Web UI for Tilt | Local only |
| ngrok UI | 4040 | View all active tunnels | Local only |

## 🌐 GitHub Actions Testing with ngrok

The ngrok Tilt extension enables testing your local C8S instance with GitHub Actions. This allows CI/CD pipelines running on GitHub to trigger webhooks and test your local development environment.

### Setup (one-time)

1. **Create free ngrok account**:
   - Go to https://ngrok.com/
   - Sign up for a free account
   - Copy your authtoken from the dashboard

2. **Configure ngrok authentication**:
   ```bash
   ngrok config add-authtoken YOUR_AUTHTOKEN_HERE
   ```

3. **Install ngrok** (if not already installed):
   ```bash
   # macOS
   brew install ngrok

   # Linux/WSL
   curl -s https://ngrok-agent.s3.amazonaws.com/ngrok.asc | sudo apt-key add - && echo "deb https://ngrok-agent.s3.amazonaws.com buster main" | sudo tee /etc/apt/sources.list.d/ngrok.list && sudo apt update && sudo apt install ngrok

   # Or download from https://ngrok.com/download
   ```

### Using ngrok with Tilt

1. **Start Tilt normally**:
   ```bash
   tilt up
   ```

2. **Create ngrok tunnels from Tilt UI**:
   - Open Tilt dashboard at http://localhost:10350
   - Find `c8s-api-server` and `c8s-webhook` resources
   - Click the "ngrok" button next to each service
   - This creates public URLs for your local services

3. **View tunnel details**:
   - Open ngrok dashboard at http://localhost:4040
   - See all active tunnels and public URLs
   - Monitor tunnel traffic and requests

### Connecting GitHub to Your Local Environment

1. **Get your webhook URL**:
   - From ngrok dashboard (http://localhost:4040), find the HTTPS tunnel for `c8s-webhook-tunnel`
   - Example: `https://abc123.ngrok.io`

2. **Configure GitHub repository webhook**:
   - Go to your GitHub repo → Settings → Webhooks
   - Click "Add webhook"
   - **Payload URL**: Use the ngrok webhook URL (e.g., `https://abc123.ngrok.io/webhook`)
   - **Content type**: `application/json`
   - **Events**: Select "Push events" or as needed
   - **Secret**: Configure if your webhook expects one
   - Click "Add webhook"

3. **Test the webhook**:
   - In GitHub webhook settings, click the webhook
   - Click "Redeliver" to test
   - Check Tilt logs to verify webhook received the request
   - Monitor ngrok dashboard to see HTTP requests

### GitHub Actions Integration

For testing that GitHub Actions can reach your local services:

1. **Create a test workflow** (`.github/workflows/test-local.yml`):
   ```yaml
   name: Test with Local Services

   on: push

   jobs:
     test:
       runs-on: ubuntu-latest
       steps:
         - name: Test API Server
           run: |
             # Use your ngrok API server URL
             curl -v https://api-abc123.ngrok.io/health

         - name: Trigger Pipeline
           run: |
             # Use your ngrok webhook URL to trigger webhook
             curl -X POST https://webhook-abc123.ngrok.io/webhook \
               -H "Content-Type: application/json" \
               -d '{...webhook payload...}'
   ```

2. **Push to trigger workflow** and verify requests appear in:
   - Tilt logs
   - ngrok dashboard

### Important Notes

- **Free tier limits**: ngrok free tier has bandwidth and connection limits
- **Public URLs**: Your ngrok URLs are public—configure GitHub secrets accordingly
- **Tunnels expire**: Free ngrok connections restart every 2 hours
- **Authentication**: For production, configure ngrok auth in Tiltfile args
- **Rate limits**: Monitor ngrok dashboard for rate limit warnings

### Troubleshooting ngrok

**ngrok not starting**:
```bash
# Check if ngrok is installed
ngrok --version

# Verify authtoken is configured
ngrok config check
```

**Can't see tunnel URLs**:
- Check ngrok dashboard at http://localhost:4040
- Click the "Online" button in Tilt UI next to the service
- Check Tilt logs for ngrok startup errors

**GitHub webhook not received**:
- Verify webhook URL in GitHub matches ngrok tunnel URL
- Check ngrok dashboard shows HTTP requests
- Review webhook delivery in GitHub (Settings → Webhooks → Recent Deliveries)
- Check C8S webhook logs: `kubectl logs -f deployment/c8s-webhook -n c8s-system`

**Connection refused**:
- Ensure Tilt is running: `tilt status`
- Ensure c8s-api-server and c8s-webhook pods are ready: `kubectl get pods -n c8s-system`
- Check Tilt UI shows green status for components

## 🎯 Common Workflows

### Testing Pipeline Definitions

```bash
# Create a test pipeline
cat > test.yaml <<EOF
version: v1alpha1
name: test-echo
steps:
  - name: hello
    image: alpine:latest
    commands:
      - echo "Hello from C8S!"
EOF

# Deploy it
kubectl apply -f test.yaml -n c8s-system

# Monitor execution
kubectl logs -n c8s-system -l app=test-echo --all-containers=true -f
```

### Debugging Multi-Component Flows

1. Open Tilt dashboard: http://localhost:10350
2. Create a PipelineRun: `kubectl apply -f config/samples/simple-build.yaml`
3. View unified logs for:
   - Webhook (receives request)
   - API Server (processes request)
   - Controller (executes job)

### Switching Between Branches

```bash
# Switch to different branch
git checkout feature/new-feature

# CRD definitions may have changed - just keep Tilt running
# It automatically detects and re-applies manifests
```

## 🛠️ Configuration

### Command-Line Options

```bash
# Disable sample pipelines
tilt up -- --with_samples=false

# Enable verbose logging
tilt up -- --verbose_logs=true

# Use different namespace
tilt up -- --k8s_namespace=my-c8s

# Custom image registry
tilt up -- --image_registry=localhost:5000
```

### ngrok Configuration

ngrok is automatically configured when you run `tilt up`. To customize ngrok settings:

```bash
# Configure ngrok authentication (one-time setup)
ngrok config add-authtoken YOUR_AUTHTOKEN

# Check your ngrok config
ngrok config check

# View all active tunnels
open http://localhost:4040

# Manual ngrok tunnel creation (alternative to Tilt UI)
ngrok http 8080          # API server tunnel
ngrok https 9443         # Webhook tunnel
```

For advanced ngrok configuration, edit `~/.config/ngrok/ngrok.yml`:
```yaml
version: "2"
authtoken: YOUR_AUTHTOKEN
tunnels:
  c8s-api:
    proto: http
    addr: 8080
  c8s-webhook:
    proto: https
    addr: 9443
```

### Accessing the Cluster

```bash
# Verify cluster context
kubectl config current-context

# If not set, configure it
kubectl config use-context k3d-c8s-dev

# List all resources
kubectl get all -n c8s-system
```

## 📊 Monitoring

### View Logs

**Via Tilt Dashboard**:
- Go to http://localhost:10350
- Click component name to view live logs
- Use search/filter for specific messages

**Via kubectl**:
```bash
# Stream logs from specific component
kubectl logs -f deployment/c8s-controller -n c8s-system

# View all logs at once
kubectl logs -f -n c8s-system --all-containers=true
```

### Check Component Status

```bash
# Tilt dashboard shows status in real-time

# Or use kubectl
kubectl get pods -n c8s-system
kubectl describe pod <pod-name> -n c8s-system
```

## 🧪 Testing

### Unit Tests

```bash
make test
```

### Integration Tests

```bash
make test-integration
```

### API Contract Tests

```bash
# Full contract test suite
make test-contract

# Quick contract tests
make test-contract-short
```

## 🐛 Troubleshooting

### Port Already in Use

```bash
# Find process using port
lsof -i :8080

# Kill it
kill -9 <PID>

# Or change port in Tiltfile
```

### Cluster Won't Start

```bash
# Check if cluster exists
k3d cluster list

# Delete and recreate
k3d cluster delete c8s-dev
tilt up
```

### Pod Stuck in CrashLoopBackOff

```bash
# Check logs for error
kubectl logs <pod-name> -n c8s-system --previous

# Or in Tilt dashboard - click resource to see logs
```

### Out of Memory

```bash
# Increase Docker memory allocation
# Mac: Docker Desktop → Settings → Resources → Memory

# Or reduce component resource requests in deploy/install.yaml
```

## 🎓 Learning More

- **Tilt Docs**: https://docs.tilt.dev
- **k3d Docs**: https://k3d.io
- **Kubernetes Docs**: https://kubernetes.io/docs
- **C8S Repository**: https://github.com/org/c8s

## ✨ Tips for Faster Development

1. **Keep Tilt running** - Don't stop/start between iterations
2. **Edit and save** - Tilt watches files automatically
3. **Check dashboard** - Verify builds succeed before testing
4. **Use filter** - Search logs to find specific issues
5. **Small changes** - Test one change at a time
6. **Monitor metrics** - Watch CPU/memory in Tilt UI

## 🤝 Contributing

When submitting PRs:

1. ✅ Test locally with Tilt
2. ✅ Verify hot reload works for your changes
3. ✅ Check logs for errors
4. ✅ Run tests: `make test`
5. ✅ Document any new configuration needed

See [CONTRIBUTING.md](CONTRIBUTING.md) for full guidelines.

## 📞 Getting Help

- **Documentation**: See docs/ and specs/ directories
- **Issues**: https://github.com/org/c8s/issues
- **Slack**: https://c8s.slack.com (if applicable)

---

**Ready to develop?** Run `tilt up` and start editing! 🚀
