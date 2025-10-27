#!/bin/bash

# Local Tilt CI Testing Script
# This script runs tilt ci locally using a kind cluster
# Useful for testing Tiltfile changes before pushing to CI

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

CLUSTER_NAME="${CLUSTER_NAME:-c8s-ci}"
TIMEOUT="${TIMEOUT:-600}"

echo -e "${YELLOW}=== Tilt CI Local Test ===${NC}"

# Check if kind is installed
if ! command -v kind &> /dev/null; then
  echo -e "${RED}Error: kind is not installed${NC}"
  echo "Install from: https://kind.sigs.k8s.io/docs/user/quick-start/"
  exit 1
fi

# Check if tilt is installed
if ! command -v tilt &> /dev/null; then
  echo -e "${RED}Error: tilt is not installed${NC}"
  echo "Install from: https://docs.tilt.dev/install.html"
  exit 1
fi

# Check if Docker is running
if ! docker info &> /dev/null; then
  echo -e "${RED}Error: Docker is not running${NC}"
  exit 1
fi

echo -e "${YELLOW}Creating kind cluster: ${CLUSTER_NAME}${NC}"

# Create kind cluster if it doesn't exist
if ! kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
  kind create cluster --name "${CLUSTER_NAME}" --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: ${CLUSTER_NAME}
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 80
    hostPort: 8080
    listenAddress: "127.0.0.1"
  - containerPort: 443
    hostPort: 8443
    listenAddress: "127.0.0.1"
EOF
  echo -e "${GREEN}Created cluster${NC}"
else
  echo -e "${YELLOW}Cluster already exists${NC}"
fi

# Set kubectl context
kubectl cluster-info --context="kind-${CLUSTER_NAME}"

echo -e "${YELLOW}Running tilt ci...${NC}"

# Run tilt ci (note: --output flag not available in all Tilt versions, using log-source instead)
if tilt ci --file=Tiltfile --timeout="${TIMEOUT}s" --log-source=all; then
  echo -e "${GREEN}Tilt CI succeeded!${NC}"

  echo -e "${YELLOW}=== Cluster Status ===${NC}"
  echo "Cluster: ${CLUSTER_NAME}"
  echo "Nodes:"
  kubectl get nodes -o wide

  echo -e "\n${YELLOW}=== C8S Namespace ===${NC}"
  kubectl get all -n c8s-system -o wide || true

  exit 0
else
  echo -e "${RED}Tilt CI failed!${NC}"

  echo -e "${YELLOW}=== Debugging Info ===${NC}"

  echo -e "\n${YELLOW}--- Cluster Info ---${NC}"
  kubectl cluster-info || true

  echo -e "\n${YELLOW}--- Nodes ---${NC}"
  kubectl get nodes -o wide || true

  echo -e "\n${YELLOW}--- All Pods ---${NC}"
  kubectl get pods -A -o wide || true

  echo -e "\n${YELLOW}--- C8S Namespace ---${NC}"
  kubectl get all -n c8s-system -o wide || true

  echo -e "\n${YELLOW}--- Recent Events ---${NC}"
  kubectl get events -A --sort-by='.lastTimestamp' | tail -20 || true

  echo -e "\n${YELLOW}--- Tilt Logs (last 50 lines) ---${NC}"
  if [ -d ~/.tilt ]; then
    tail -50 ~/.tilt/*.log 2>/dev/null || true
  fi

  exit 1
fi
