# AIDU Orchestration Quick Start

This guide will help you get started with the AIDU manager/worker orchestration system.

## Prerequisites

- k3d cluster running (see K3D-SETUP.md)
- kubectl configured
- Go 1.21 or later (for building)
- Docker
- aidu-claude:latest image available

## Building the Manager

```bash
# From the project root
go build -o bin/aidu-manager ./cmd/aidu-manager
```

This creates the `bin/aidu-manager` binary that orchestrates worker pods.

## Basic Usage

### 1. Create an AGENTS.md file (optional)

Create an `AGENTS.md` file in your project directory to provide context for the workers:

```markdown
# My Project Agent Context

## Project Overview
This is a REST API built with Go that handles user management.

## Architecture
- Handler layer for HTTP endpoints
- Service layer for business logic
- Repository layer for database access

## Coding Standards
- Follow Go best practices
- Use proper error handling
- Write table-driven tests
```

If you don't create one, the default `AGENTS.md` from the repo will be used.

### 2. Run a task using the orchestrate command

```bash
# Simple question
./aidu-session orchestrate "what is 2+2"

# Data fetching
./aidu-session orchestrate "fetch the latest stock price for AAPL"

# Code implementation
./aidu-session orchestrate "implement a JWT authentication middleware in Go"

# Complex task
./aidu-session orchestrate "build a REST API endpoint for user registration with validation and tests"
```

### 3. Using the manager directly

You can also use the `aidu-manager` binary directly for more control:

```bash
./bin/aidu-manager \
  --namespace aidu-sessions \
  --max-retries 3 \
  --timeout 10m \
  --validation-score 0.75 \
  --agents-file ./AGENTS.md \
  "implement user authentication"
```

## How It Works

1. **Task Analysis**: The manager analyzes your task to determine complexity
2. **Context Injection**: Your AGENTS.md and task description are packaged into a ConfigMap
3. **Worker Pod Creation**: A worker pod is spawned with Claude Code CLI in yolo mode
4. **Execution**: The worker executes the task autonomously
5. **Validation**: The manager validates the output using heuristics
6. **Retry if Needed**: If validation fails, a new worker is spawned with feedback
7. **Success**: When satisfied, the output is available for review

## Configuration Options

### Manager Flags

- `--namespace`: Kubernetes namespace (default: `aidu-sessions`)
- `--max-retries`: Maximum retry attempts (default: `3`)
- `--timeout`: Worker execution timeout (default: `10m`)
- `--validation-score`: Minimum acceptance threshold (default: `0.75`)
- `--worker-image`: Container image to use (default: `aidu-claude:latest`)
- `--agents-file`: Path to AGENTS.md (default: `./AGENTS.md`)
- `--output-dir`: Output directory (default: `./output`)

## Task Complexity Levels

The manager automatically categorizes tasks:

- **SIMPLE**: Direct questions (e.g., "what is 1+1")
- **DATA_FETCH**: External data retrieval (e.g., "get stock price")
- **IMPLEMENT**: Code implementation (e.g., "add feature X")
- **COMPLEX**: Multi-step orchestration (e.g., "build microservice")

## Output and Results

Worker outputs are stored in PersistentVolumeClaims and validated before being considered complete. The manager checks for:

- Presence of expected files
- Error indicators in logs
- Completeness of implementation
- Test coverage (for code tasks)

## Validation Scoring

The manager calculates a validation score (0.0 to 1.0):

- **1.0**: Perfect output, no issues
- **0.75+**: Acceptable (default threshold)
- **< 0.75**: Retry with feedback

Score deductions:
- Errors found: -0.3
- Non-zero exit code: -0.2
- Missing code files: -0.4
- Missing tests: -0.2

## Troubleshooting

### Manager binary not found

```bash
# Build the manager
go build -o bin/aidu-manager ./cmd/aidu-manager
```

### K8s connection issues

```bash
# Check your kubectl context
kubectl config current-context

# Verify namespace exists
kubectl get namespace aidu-sessions
```

### Worker pod failures

```bash
# Check worker pod logs
kubectl logs -n aidu-sessions worker-<task-id>

# Check pod events
kubectl describe pod -n aidu-sessions worker-<task-id>
```

### Image not found

Make sure the `aidu-claude:latest` image is available:

```bash
# Build the image (see Dockerfile)
docker build -t aidu-claude:latest .

# Import into k3d
k3d image import aidu-claude:latest
```

## Examples

### Example 1: Simple Implementation

```bash
./aidu-session orchestrate "create a Go function that calculates factorial"
```

The manager will:
1. Detect complexity: IMPLEMENT
2. Spawn worker with task
3. Worker creates factorial function with tests
4. Manager validates output
5. Output available in results

### Example 2: Feature with Tests

```bash
./aidu-session orchestrate "implement password hashing using bcrypt with tests"
```

The manager will:
1. Create worker with full context
2. Worker implements password hashing
3. Worker creates comprehensive tests
4. Manager validates (checks for code + tests)
5. May retry if tests missing

### Example 3: Custom Context

```bash
# Create project-specific AGENTS.md
cat > AGENTS.md <<EOF
# E-commerce API Agent Context
## Tech Stack
- Go 1.21
- PostgreSQL
- Chi router
## Standards
- Use repository pattern
- All endpoints need auth middleware
- Write integration tests
EOF

./aidu-session orchestrate "add product CRUD endpoints"
```

Worker will use your custom context to build endpoints following your patterns.

## Next Steps

- Read [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed design
- Customize your AGENTS.md for project-specific needs
- Experiment with different task complexities
- Monitor worker performance and adjust validation thresholds

## Advanced Usage

### Interactive Retries

Watch the manager's decisions in real-time:

```bash
./bin/aidu-manager --max-retries 5 "complex task" 2>&1 | tee orchestration.log
```

### Custom Validation Threshold

For stricter validation:

```bash
./bin/aidu-manager --validation-score 0.90 "implement feature"
```

### Extended Timeout

For long-running tasks:

```bash
./bin/aidu-manager --timeout 30m "build entire microservice"
```

## Support

- Check [ARCHITECTURE.md](./ARCHITECTURE.md) for system design
- See [README.md](./README.md) for project overview
- Review [K3D-SETUP.md](./K3D-SETUP.md) for cluster setup
