# Using GHCR with Your GitHub Account

This guide explains how to use GitHub Container Registry (GHCR) with the C8S Helm chart for your personal or organization repository.

## How It Works

The Tiltfile and GitHub Actions automatically detect your GitHub repository and use the correct GHCR path:

**Your Setup**:
- Repository: `git@github.com:lavigneer/c8s.git`
- GHCR Registry: `ghcr.io/lavigneer/c8s`

**How it's detected**:
1. **Local (Tilt)**: Reads from `git config remote.origin.url`
2. **CI/CD (GitHub Actions)**: Uses `${{ github.repository }}` variable
3. **Both**: Always uses the correct owner/repo path from your git remote

## Local Development

### Prerequisites

```bash
# Check your git remote is set correctly
git remote -v
# Should show: git@github.com:lavigneer/c8s.git

# Verify the registry that will be used
git config --get remote.origin.url
# Shows: git@github.com:lavigneer/c8s.git
# Maps to: ghcr.io/lavigneer/c8s
```

### Using Tilt

Simply run Tilt - it will auto-detect your repository:

```bash
tilt up
```

**What happens**:
- Reads your git remote automatically
- Builds images with reference: `c8s-api-server`, `c8s-controller`, `c8s-webhook`
- Deploys to your local cluster with local Docker registry
- When output, shows: `ghcr.io/lavigneer/c8s/c8s-api-server:latest`

## Publishing to GHCR

### Step 1: Create GitHub Personal Access Token (PAT)

You need a token with `write:packages` permission to push to GHCR:

1. Go to https://github.com/settings/tokens/new
2. Select scopes:
   - ✅ `write:packages` (write to GitHub Container Registry)
   - ✅ `read:packages` (read from packages)
   - ✅ `repo` (for private repositories)
3. Copy the token (you'll need it once)

### Step 2: Authenticate with GHCR

```bash
# Save your token to a file (for automation) or enter it when prompted
export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx

# Log in to GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u lavigneer --password-stdin
```

### Step 3: Tag and Push Local Images

After building with Tilt:

```bash
# Tag the locally-built images
docker tag c8s-api-server ghcr.io/lavigneer/c8s/c8s-api-server:latest
docker tag c8s-controller ghcr.io/lavigneer/c8s/c8s-controller:latest
docker tag c8s-webhook ghcr.io/lavigneer/c8s/c8s-webhook:latest

# Push to GHCR
docker push ghcr.io/lavigneer/c8s/c8s-api-server:latest
docker push ghcr.io/lavigneer/c8s/c8s-controller:latest
docker push ghcr.io/lavigneer/c8s/c8s-webhook:latest
```

Or push with a version tag:

```bash
# Tag with version
docker tag c8s-api-server ghcr.io/lavigneer/c8s/c8s-api-server:v0.1.0
docker tag c8s-controller ghcr.io/lavigneer/c8s/c8s-controller:v0.1.0
docker tag c8s-webhook ghcr.io/lavigneer/c8s/c8s-webhook:v0.1.0

# Push
docker push ghcr.io/lavigneer/c8s/c8s-api-server:v0.1.0
docker push ghcr.io/lavigneer/c8s/c8s-controller:v0.1.0
docker push ghcr.io/lavigneer/c8s/c8s-webhook:v0.1.0
```

## Automatic Publishing with GitHub Actions

When you push a git tag, GitHub Actions automatically builds and publishes images:

```bash
# Create a version tag
git tag v0.1.0

# Push the tag
git push origin v0.1.0
```

**GitHub Actions will**:
1. Build all component images
2. Push to `ghcr.io/lavigneer/c8s/c8s-api-server:v0.1.0`
3. Push to `ghcr.io/lavigneer/c8s/c8s-controller:v0.1.0`
4. Push to `ghcr.io/lavigneer/c8s/c8s-webhook:v0.1.0`
5. Create a GitHub Release with the Helm chart

## Using Published Images

Once images are published to GHCR, deploy using Helm:

```bash
# Deploy with latest images
helm install c8s ./chart/c8s \
  --set components.apiServer.image.registry=ghcr.io \
  --set components.apiServer.image.repository=lavigneer/c8s/c8s-api-server \
  --set components.apiServer.image.tag=v0.1.0 \
  --set components.controller.image.registry=ghcr.io \
  --set components.controller.image.repository=lavigneer/c8s/c8s-controller \
  --set components.controller.image.tag=v0.1.0 \
  --set components.webhook.image.registry=ghcr.io \
  --set components.webhook.image.repository=lavigneer/c8s/c8s-webhook \
  --set components.webhook.image.tag=v0.1.0 \
  -n c8s-system --create-namespace
```

Or create a custom values file:

```yaml
# ghcr-values.yaml
components:
  apiServer:
    image:
      registry: ghcr.io
      repository: lavigneer/c8s/c8s-api-server
      tag: v0.1.0
  controller:
    image:
      registry: ghcr.io
      repository: lavigneer/c8s/c8s-controller
      tag: v0.1.0
  webhook:
    image:
      registry: ghcr.io
      repository: lavigneer/c8s/c8s-webhook
      tag: v0.1.0
```

Then:

```bash
helm install c8s ./chart/c8s -f ghcr-values.yaml
```

## Making Images Public

By default, your GHCR images are private. To make them public:

1. Go to: https://github.com/lavigneer/c8s/settings/packages
2. Select each package (c8s-api-server, c8s-controller, c8s-webhook)
3. Change visibility to "Public"

This allows others to pull images without authentication:

```bash
docker pull ghcr.io/lavigneer/c8s/c8s-api-server:v0.1.0
```

## Troubleshooting

### "authentication required" error

**Problem**: `docker push` fails with authentication error

**Solution**:
```bash
# Re-authenticate with GHCR
echo $GITHUB_TOKEN | docker login ghcr.io -u lavigneer --password-stdin

# Or use personal access token interactively
docker login ghcr.io
# Username: lavigneer
# Password: [paste your PAT token]
```

### Images not found after push

**Problem**: Pushed image but `docker pull` fails

**Solution**:
1. Verify image was pushed: Go to https://github.com/lavigneer/c8s/settings/packages
2. Check visibility: Change to "Public" if needed
3. Wait a few seconds for GHCR to index the image
4. Try pulling again: `docker pull ghcr.io/lavigneer/c8s/c8s-api-server:latest`

### GitHub Actions not building images

**Problem**: Pushed code but no images in GHCR

**Solution**:
1. Check GitHub Actions logs: https://github.com/lavigneer/c8s/actions
2. Verify Dockerfile exists and is valid
3. Check workflow file: `.github/workflows/build-and-push.yml`
4. Ensure `go.mod`, `go.sum`, `cmd/`, and `pkg/` directories exist

## Organization Repositories

If using an organization repository (e.g., `github.com/myorg/c8s`):

1. Registry path: `ghcr.io/myorg/c8s`
2. Git remote: `git@github.com:myorg/c8s.git`
3. Everything else works the same - no config changes needed\!

## Multiple Registries

To push to multiple registries (Docker Hub, ECR, private registry):

Modify `.github/workflows/build-and-push.yml` and add additional push steps for each registry.

## Security

### Keep Tokens Safe

- Never commit tokens to git
- Use GitHub repository secrets for CI/CD
- Rotate tokens periodically
- Delete tokens from `.docker/config.json` after use

### Image Signing (Optional)

For production, consider signing images with Cosign:

```bash
# Install cosign
brew install cosign

# Sign image after push
cosign sign ghcr.io/lavigneer/c8s/c8s-api-server:v0.1.0
```

## Related Documentation

- [GitHub Container Registry Docs](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Creating Personal Access Tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)
- [Helm Chart README](../chart/c8s/README.md)
- [Image Building Guide](./image-building.md)
