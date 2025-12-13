#!/bin/bash
# Wait for a pod to be ready, then start port-forward
# Usage: ./wait-and-forward.sh <namespace> <pod-selector> <local-port> <remote-port>

set -e

NAMESPACE=$1
POD_SELECTOR=$2
LOCAL_PORT=$3
REMOTE_PORT=$4

if [ -z "$NAMESPACE" ] || [ -z "$POD_SELECTOR" ] || [ -z "$LOCAL_PORT" ] || [ -z "$REMOTE_PORT" ]; then
  echo "Usage: $0 <namespace> <pod-selector> <local-port> <remote-port>"
  exit 1
fi

echo "Waiting for pod with selector '$POD_SELECTOR' in namespace '$NAMESPACE' to be ready..."

# Wait for pod to be ready with a timeout of 5 minutes
timeout=300
elapsed=0
interval=2

while [ $elapsed -lt $timeout ]; do
  # Check if pod exists and is ready
  READY_PODS=$(kubectl get pods -n "$NAMESPACE" -l "$POD_SELECTOR" \
    -o jsonpath='{.items[?(@.status.conditions[?(@.type=="Ready")].status=="True")].metadata.name}' 2>/dev/null || echo "")

  if [ -n "$READY_PODS" ]; then
    echo "Pod is ready! Starting port-forward..."
    # Use the service instead of pod for more reliable port-forwarding
    exec kubectl port-forward -n "$NAMESPACE" "svc/$POD_SELECTOR" "$LOCAL_PORT:$REMOTE_PORT"
  fi

  sleep $interval
  elapsed=$((elapsed + interval))
done

echo "Timeout waiting for pod to be ready after ${timeout}s"
exit 1
