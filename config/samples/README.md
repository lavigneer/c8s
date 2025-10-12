# C8S Sample Custom Resources

This directory contains example Custom Resource definitions for C8S.

## Sample Files

### PipelineConfig
`pipelineconfig_example.yaml` - Complete pipeline configuration with:
- Multi-step pipeline (test, lint, build, security-scan)
- Dependencies between steps
- Resource requests and timeouts
- Artifact collection
- Retry policy

Apply with:
```bash
kubectl apply -f config/samples/pipelineconfig_example.yaml
```

### PipelineRun
`pipelinerun_example.yaml` - Manual pipeline run trigger with:
- Reference to PipelineConfig
- Commit SHA and branch
- Trigger metadata (who, when, message)

Apply with:
```bash
kubectl apply -f config/samples/pipelinerun_example.yaml
```

### RepositoryConnection
`repositoryconnection_example.yaml` - Git webhook configuration with:
- GitHub webhook integration
- Event filtering (push, pull_request)
- Branch and tag patterns
- Webhook secret reference

Apply with:
```bash
# First create webhook secret
kubectl create secret generic example-webhook-secret \
  --from-literal=webhook-secret=$(openssl rand -hex 32)

# Then apply repository connection
kubectl apply -f config/samples/repositoryconnection_example.yaml
```

## Complete Example Workflow

1. Create webhook secret:
```bash
kubectl create secret generic example-webhook-secret \
  --from-literal=webhook-secret=$(openssl rand -hex 32)
```

2. Create PipelineConfig:
```bash
kubectl apply -f config/samples/pipelineconfig_example.yaml
```

3. Create RepositoryConnection:
```bash
kubectl apply -f config/samples/repositoryconnection_example.yaml
```

4. Manually trigger a run:
```bash
kubectl apply -f config/samples/pipelinerun_example.yaml
```

5. Watch pipeline execution:
```bash
kubectl get pipelinerun -w
kubectl logs -f job/example-go-pipeline-abc123-test
```

## Notes

- Replace `example-org/example-repo` with your actual repository URL
- Update commit SHAs, branches, and resource names as needed
- Ensure webhook secret matches what's configured in GitHub/GitLab
- PipelineRun names are typically auto-generated with `generateName`
