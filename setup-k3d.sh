#!/bin/bash

set -e

echo "=== AIDU K3D Cluster Setup ==="
echo ""

if ! command -v k3d &> /dev/null; then
    echo "Error: k3d is not installed"
    echo "Install with: brew install k3d or check installation instructions here: https://github.com/k3d-io/k3d?tab=readme-ov-file#get"
    exit 1
fi

if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "Error: docker is not installed"
    exit 1
fi

CLUSTER_NAME="aidu-cluster"
REGISTRY_NAME="aidu-registry"
REGISTRY_PORT="5050"
STORAGE_DIR="$(pwd)/k3d-storage"

if k3d cluster list | grep -q "$CLUSTER_NAME"; then
    echo "Cluster '$CLUSTER_NAME' already exists"
    read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Deleting existing cluster..."
        k3d cluster delete "$CLUSTER_NAME"
    else
        echo "Using existing cluster"
        k3d cluster start "$CLUSTER_NAME" 2>/dev/null || true
        kubectl config use-context "k3d-$CLUSTER_NAME"
        exit 0
    fi
fi

echo "Creating storage directory: $STORAGE_DIR"
mkdir -p "$STORAGE_DIR"

echo "Creating k3d cluster: $CLUSTER_NAME"
if ! k3d cluster create "$CLUSTER_NAME" \
  --agents 2 \
  --registry-create "${REGISTRY_NAME}:0.0.0.0:${REGISTRY_PORT}" \
  --volume "${STORAGE_DIR}:/var/lib/rancher/k3s/storage@all"; then
    echo ""
    echo "Error: Failed to create cluster"
    echo "This may be due to leftover Docker resources."
    echo ""
    echo "Run './cleanup-k3d.sh' to clean up and try again."
    exit 1
fi

echo ""
echo "Building Docker image..."
docker build -t localhost:${REGISTRY_PORT}/aidu-claude:latest .

echo ""
echo "Pushing image to local registry..."
docker push localhost:${REGISTRY_PORT}/aidu-claude:latest

echo ""
echo "Importing image into cluster nodes..."
k3d image import localhost:${REGISTRY_PORT}/aidu-claude:latest -c "$CLUSTER_NAME"

echo ""
echo "Creating namespace..."
kubectl apply -f k8s/namespace.yaml

echo ""
echo "=== Setup Complete ==="
echo ""
echo "Next steps:"
echo "1. Create a session:"
echo "   ./aidu-session create my-project"
echo ""
echo "2. Connect to the session:"
echo "   ./aidu-session connect my-project"
echo ""
echo "3. Authenticate Claude (first time only):"
echo "   Inside the session, run: claude"
echo "   Log in with your Claude.ai Pro account"
echo ""
echo "See K3D-SETUP.md for detailed documentation"
