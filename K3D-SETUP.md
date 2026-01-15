## K3D Setup for Isolated Claude Code Sessions

This guide explains how to run isolated Claude Code sessions in a k3d Kubernetes environment. Each session runs in its own pod with dedicated persistent storage, completely isolated from other sessions.

### Prerequisites

1. **Docker** - Install from https://docs.docker.com/get-docker/
2. **k3d** - Install with:
   ```bash
   curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
   ```
3. **kubectl** - Install from https://kubernetes.io/docs/tasks/tools/

### Initial Setup

#### 1. Create k3d Cluster

Create a k3d cluster with local registry for storing the Claude Code image:

```bash
k3d cluster create aidu-cluster \
  --agents 2 \
  --registry-create aidu-registry:0.0.0.0:5050 \
  --volume /tmp/k3d-storage:/var/lib/rancher/k3s/storage@all
```

This creates:

- A cluster named `aidu-cluster`
- 2 agent nodes for running sessions
- A local registry on port 5050
- Persistent storage mounted from `/tmp/k3d-storage`

#### 2. Build and Push Docker Image

Build the Claude Code image and push to the local registry:

```bash
docker build -t localhost:5050/aidu-claude:latest .
docker push localhost:5050/aidu-claude:latest
```

Update the image name in `k8s/session-template.yaml`:

```yaml
image: localhost:5050/aidu-claude:latest
```

#### 3. Create Namespace

```bash
kubectl apply -f k8s/namespace.yaml
```

Or the CLI will create it automatically on first run.

#### 4. Set Anthropic API Key

Export your Anthropic API key:

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

For persistent configuration, add to your shell profile:

```bash
echo 'export ANTHROPIC_API_KEY="your-api-key-here"' >> ~/.bashrc
```

### Usage

#### Create a New Session

```bash
./aidu-session create my-project
```

This creates:

- A PersistentVolumeClaim (10Gi) for workspace storage
- A Secret containing the Anthropic API key
- A Pod running Claude Code

#### List Active Sessions

```bash
./aidu-session list
```

Shows all running sessions with their status and age.

#### Connect to a Session

```bash
./aidu-session connect my-project
```

This opens an interactive shell in the session pod. Once connected:

```bash
cd /workspace
claude
```

Type `exit` to disconnect from the session without stopping it.

#### Check Session Status

```bash
./aidu-session status my-project
```

Shows detailed information about the session pod.

#### View Session Logs

```bash
./aidu-session logs my-project
```

Streams the last 100 lines of logs from the session.

#### Delete a Session

```bash
./aidu-session delete my-project
```

Removes the pod, secret, and PersistentVolumeClaim. All data will be lost.

#### Cleanup Failed Sessions

```bash
./aidu-session cleanup
```

Removes all pods in Succeeded or Failed state.

### Architecture

```
┌─────────────────────────────────────────┐
│          k3d Cluster                    │
│                                         │
│  ┌────────────────────────────────┐    │
│  │  Namespace: aidu-sessions      │    │
│  │                                │    │
│  │  ┌──────────────────────────┐ │    │
│  │  │  Session: project-1      │ │    │
│  │  │  - Pod (claude-code)     │ │    │
│  │  │  - PVC (10Gi storage)    │ │    │
│  │  │  - Secret (API key)      │ │    │
│  │  └──────────────────────────┘ │    │
│  │                                │    │
│  │  ┌──────────────────────────┐ │    │
│  │  │  Session: project-2      │ │    │
│  │  │  - Pod (claude-code)     │ │    │
│  │  │  - PVC (10Gi storage)    │ │    │
│  │  │  - Secret (API key)      │ │    │
│  │  └──────────────────────────┘ │    │
│  │                                │    │
│  └────────────────────────────────┘    │
│                                         │
└─────────────────────────────────────────┘
         ▲
         │
    kubectl exec
         │
   ┌─────────────┐
   │ aidu-session│
   │     CLI     │
   └─────────────┘
```

### Session Isolation

Each session is completely isolated:

1. **Process Isolation**: Each session runs in its own pod with dedicated container
2. **Storage Isolation**: Each session has its own PersistentVolumeClaim
3. **Network Isolation**: Pods can communicate only if explicitly configured
4. **Resource Isolation**: CPU and memory limits prevent resource exhaustion
   - Requests: 500m CPU, 512Mi RAM
   - Limits: 2000m CPU, 2Gi RAM

### Workflow Example

```bash
# Create session for frontend work
export ANTHROPIC_API_KEY="sk-..."
./aidu-session create ttt-frontend

# Connect and work
./aidu-session connect ttt-frontend
cd /workspace
git clone https://github.com/user/repo.git
cd repo
claude

# In another terminal, create backend session
./aidu-session create ttt-backend
./aidu-session connect ttt-backend
cd /workspace
git clone https://github.com/user/repo.git
cd repo
claude

# List all sessions
./aidu-session list

# When done, cleanup
./aidu-session delete ttt-frontend
./aidu-session delete ttt-backend
```

### Advanced Configuration

#### Adjusting Resources

Edit `k8s/session-template.yaml` to modify resource limits:

```yaml
resources:
  requests:
    memory: "1Gi"
    cpu: "1000m"
  limits:
    memory: "4Gi"
    cpu: "4000m"
```

#### Changing Storage Size

Modify the PVC size in `k8s/session-template.yaml`:

```yaml
resources:
  requests:
    storage: 20Gi
```

#### Adding Custom Environment Variables

Add to the pod spec in `k8s/session-template.yaml`:

```yaml
env:
  - name: MY_CUSTOM_VAR
    value: "custom-value"
```

### Troubleshooting

#### Pod Won't Start

```bash
./aidu-session status <session-name>
kubectl get events -n aidu-sessions
```

#### Storage Issues

Check PVC status:

```bash
kubectl get pvc -n aidu-sessions
```

#### Image Pull Errors

Ensure image is in local registry:

```bash
docker images | grep aidu-claude
k3d image import localhost:5050/aidu-claude:latest -c aidu-cluster
```

#### Cannot Connect to Session

Verify pod is running:

```bash
kubectl get pods -n aidu-sessions
```

### Cleanup

#### Delete Specific Session

```bash
./aidu-session delete <session-name>
```

#### Delete All Sessions

```bash
kubectl delete namespace aidu-sessions
kubectl create namespace aidu-sessions
```

#### Delete Cluster

```bash
k3d cluster delete aidu-cluster
```

### Integration with AIDU

This setup provides the foundation for your AIDU orchestration system. The AAA Manager can:

1. Spawn sub-agents by calling `./aidu-session create`
2. Distribute tasks to different sessions
3. Monitor progress via `./aidu-session status`
4. Collect results from each isolated environment
5. Clean up completed sessions with `./aidu-session delete`

Each sub-agent works in complete isolation, preventing context pollution and enabling parallel execution of multiple tasks.
