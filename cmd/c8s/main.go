// Package main provides a deprecation notice for the c8s CLI tool.
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintf(os.Stderr, `c8s - Kubernetes-native CI system

The c8s binary is no longer needed for typical operations.
Use kubectl and standard Kubernetes tools instead:

  # Apply a PipelineConfig
  kubectl apply -f pipeline.yaml

  # Trigger a pipeline run
  kubectl apply -f - <<EOF
apiVersion: c8s.io/v1alpha1
kind: PipelineRun
metadata:
  name: my-pipeline-run-$(date +%%s)
spec:
  pipelineConfigName: my-pipeline
  branch: main
  commit: <sha>
EOF

  # View pipeline runs
  kubectl get pipelineruns

  # View logs
  kubectl logs -n <namespace> <pod-name> -c <step-name> --follow

For local development, use Tilt:
  tilt up

`)
	os.Exit(1)
}
