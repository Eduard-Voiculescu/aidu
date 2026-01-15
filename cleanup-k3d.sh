#!/bin/bash

set -e

CLUSTER_NAME="aidu-cluster"

echo "=== AIDU K3D Cluster Cleanup ==="
echo ""

read -p "This will delete the cluster and all sessions. Continue? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled"
    exit 0
fi

echo "Deleting k3d cluster: $CLUSTER_NAME"
k3d cluster delete "$CLUSTER_NAME" 2>/dev/null || echo "Cluster not found or already deleted"

echo "Removing leftover Docker networks..."
docker network ls --filter name=k3d-${CLUSTER_NAME} --format "{{.Name}}" | while read network; do
    if [[ -n "$network" ]]; then
        echo "Removing network: $network"
        docker network rm "$network" 2>/dev/null || echo "Network already removed"
    fi
done

echo "Removing leftover Docker volumes..."
docker volume ls --filter name=k3d-${CLUSTER_NAME} --format "{{.Name}}" | while read volume; do
    if [[ -n "$volume" ]]; then
        echo "Removing volume: $volume"
        docker volume rm "$volume" 2>/dev/null || echo "Volume already removed"
    fi
done

echo "Removing leftover Docker containers..."
docker ps -a --filter name=k3d-${CLUSTER_NAME} --format "{{.Names}}" | while read container; do
    if [[ -n "$container" ]]; then
        echo "Removing container: $container"
        docker rm -f "$container" 2>/dev/null || echo "Container already removed"
    fi
done

echo "Removing leftover registry containers..."
docker ps -a --format "{{.ID}} {{.Image}} {{.Names}}" | grep -E "registry.*aidu-registry|k3d.*aidu.*registry" | awk '{print $1}' | while read container_id; do
    if [[ -n "$container_id" ]]; then
        echo "Removing registry container: $container_id"
        docker rm -f "$container_id" 2>/dev/null || echo "Container already removed"
    fi
done

echo ""
echo "=== Cleanup Complete ==="
echo ""
echo "You can now run ./setup-k3d.sh to create a fresh cluster"
