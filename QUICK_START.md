# C8S Local Development - Quick Start Card

## 60-Second Setup

```bash
# 1. Create cluster (takes ~2 min)
c8s dev cluster create my-dev --wait

# 2. Deploy operator
c8s dev deploy operator --cluster my-dev

# 3. Deploy samples
c8s dev deploy samples --cluster my-dev

# 4. Run tests
c8s dev test run --cluster my-dev

# 5. View results
c8s dev test logs --cluster my-dev
```

## Common Commands

| Task | Command |
|------|---------|
| Create cluster | `c8s dev cluster create NAME` |
| Delete cluster | `c8s dev cluster delete NAME` |
| Show status | `c8s dev cluster status NAME` |
| List clusters | `c8s dev cluster list` |
| Stop cluster | `c8s dev cluster stop NAME` |
| Start cluster | `c8s dev cluster start NAME` |
| Deploy operator | `c8s dev deploy operator --cluster NAME` |
| Deploy samples | `c8s dev deploy samples --cluster NAME` |
| Run tests | `c8s dev test run --cluster NAME` |
| View logs | `c8s dev test logs --cluster NAME` |
| Stream logs | `c8s dev test logs --follow` |

## Development Workflow

```bash
# Initial setup
c8s dev cluster create dev-env
c8s dev deploy operator --cluster dev-env
c8s dev deploy samples --cluster dev-env

# Iterate on changes
make build          # Rebuild CLI
c8s dev test run --cluster dev-env --output json

# Cleanup
c8s dev cluster delete dev-env
```

## Environment Variables

```bash
# Persistent defaults
export C8S_DEV_CLUSTER=my-cluster
export C8S_NAMESPACE=c8s-system
export C8S_VERBOSE=true
export C8S_NO_COLOR=false
```

## Troubleshooting

| Problem | Solution |
|---------|----------|
| Docker not found | `export PATH="$HOME/.rd/bin:$PATH"` |
| Cluster won't create | Check Docker: `docker info` |
| Image load fails | Build locally: `docker build -t my-img --target controller .` |
| Tests fail to connect | Wait longer: `--timeout 600` |
| Kubeconfig issues | `kubectl config use-context k3d-NAME` |

## Direct kubectl Access

```bash
# Get nodes
kubectl --context k3d-my-cluster get nodes

# View PipelineConfigs
kubectl get pipelineconfigs

# Describe resource
kubectl describe pipelineconfig simple-build

# View logs
kubectl logs -n c8s-system -l app=c8s-controller
```

## File Locations

| Item | Path |
|------|------|
| Local testing guide | `docs/local-testing.md` |
| Default config | `.c8s/cluster-defaults.yaml` |
| Sample pipelines | `config/samples/` |
| CLI binary | `./bin/c8s` |

## Key URLs

- Documentation: `docs/local-testing.md`
- Full Summary: `IMPLEMENTATION_SUMMARY.md`
- This Quick Start: `QUICK_START.md`

## Make Targets

```bash
make build                    # Build CLI
make test-contract-short      # Run tests
make dev-help                 # Show dev commands
make clean-clusters           # Delete all clusters
make help                      # All make targets
```

## Output Formats

```bash
# Human-readable (default)
c8s dev test run --cluster my-dev

# JSON output for automation
c8s dev test run --cluster my-dev --output json

# YAML output
c8s dev test run --cluster my-dev --output yaml
```

## Real Examples

```bash
# Test specific pipeline
c8s dev test run --cluster dev --pipeline simple-build

# Stream logs for specific pipeline
c8s dev test logs --cluster dev --pipeline simple-build --follow

# Get last 50 lines
c8s dev test logs --cluster dev --tail 50

# Use custom image
c8s dev deploy operator --cluster dev --image my-controller:v1.0

# Custom namespace
c8s dev deploy samples --cluster dev --namespace my-tests

# Filter samples
c8s dev deploy samples --cluster dev --select multi-step
```

## Performance Tips

1. **Reuse clusters** - Stop/start instead of delete/recreate
2. **Cache images** - Build once, deploy to multiple clusters
3. **Parallel testing** - Create multiple clusters for parallel tests
4. **Allocate resources** - Give Docker more CPU/memory for faster builds

## Getting Help

```bash
c8s dev --help                          # Dev command help
c8s dev cluster --help                  # Cluster subcommands
c8s dev deploy operator --help          # Operator deploy help
c8s dev test run --help                 # Test run help
make dev-help                           # Make targets
```

## What's Working

✅ Create/manage k3d clusters
✅ Deploy CRDs to cluster
✅ Deploy sample pipelines
✅ Run and monitor tests
✅ Stream logs
✅ Full lifecycle management (stop/start/delete)
✅ Automatic kubeconfig management
✅ Multiple output formats

## What's Next

For complete details, see:
1. `docs/local-testing.md` - Full user guide
2. `IMPLEMENTATION_SUMMARY.md` - Complete project overview
3. `README.md` - Project background
