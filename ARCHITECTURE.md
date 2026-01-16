# AIDU Orchestration Architecture

## Overview

The AIDU orchestration system implements a manager/worker pattern where a manager process orchestrates isolated Claude Code sessions running in Kubernetes pods. The manager analyzes tasks, spawns worker pods with appropriate context, evaluates their outputs, and iteratively refines until achieving satisfactory results.

## Core Principles

1. **Isolation First**: Workers run in disposable K8s pods with Claude Code CLI in yolo mode
2. **Manager as Orchestrator**: The manager is NOT an LLM agent - it's a coordination system
3. **Iterative Refinement**: Manager spawns workers until output meets acceptance criteria
4. **Context Injection**: Tasks and project context (AGENTS.md) are injected into worker pods
5. **Well-Scoped Tasks**: User requests are assumed to be clear and well-defined

## Architecture Components

### 1. Manager (Go Application)

**Responsibilities:**

- Analyze incoming task complexity
- Create worker pods with injected context
- Monitor worker execution and outputs
- Validate results using heuristics
- Decide on: accept, retry with refinements, or spawn new worker
- Commit accepted outputs to host machine

**Does NOT:**

- Run LLM inference directly
- Execute code itself
- Make subjective decisions (uses heuristics)

**Location:** `cmd/aidu-manager/`

### 2. Worker Pods (K8s Pods with Claude Code CLI)

**Characteristics:**

- Ephemeral and isolated
- Run Claude Code CLI with `--yolo` mode enabled
- Receive task via mounted ConfigMap or environment
- Produce output to shared volume
- Self-contained with all dependencies

**Pod Configuration:**

- Uses existing `aidu-claude:latest` image
- Mounts task context and AGENTS.md
- Dedicated storage for output artifacts
- Resource limits to prevent runaway processes

**Location:** `k8s/worker-pod-template.yaml`

### 3. Context System

**AGENTS.md:**

- Project-specific context and instructions
- Injected into every worker pod
- Contains:
  - Project overview
  - Coding standards
  - Architecture guidelines
  - Domain knowledge
  - Constraints and requirements

**Task Context:**

- Specific task description from user
- Previous attempt feedback (for retries)
- Expected output format
- Success criteria hints

**Location:** `pkg/config/`

### 4. Task Analyzer

**Complexity Levels:**

```txt
SIMPLE       → Direct questions (e.g., "what is 1+1")
DATA_FETCH   → External data retrieval (e.g., "fetch stock data for AAPL")
IMPLEMENT    → Code implementation (e.g., "add authentication")
COMPLEX      → Multi-step orchestration (e.g., "build microservice")
```

**Heuristics:**

- Keyword matching (implement, build, create, fetch, etc.)
- Task length and structure
- Presence of technical terms
- Mentioned technologies/frameworks

**Location:** `pkg/analyzer/`

### 5. Output Validator

**Validation Strategies:**

- File presence checks
- Build/test success indicators
- Error log analysis
- Completeness heuristics
- Format validation

**Decision Logic:**

```txt
IF output contains errors → RETRY with error context
IF output incomplete → RETRY with completion request
IF output format invalid → RETRY with format guidance
IF all checks pass → ACCEPT
IF max retries reached → ESCALATE to user
```

**Location:** `pkg/validator/`

## System Flow

### High-Level Flow

```txt
┌──────────┐
│   User   │
└────┬─────┘
     │ task
     ▼
┌─────────────────┐
│  aidu-session   │
│  orchestrate    │
└────┬────────────┘
     │
     ▼
┌──────────────────────────────────────────┐
│           Manager Process                │
│                                          │
│  1. Analyze task complexity              │
│  2. Load AGENTS.md context               │
│  3. Create worker pod spec               │
│  4. Submit to K8s                        │
│  5. Monitor execution                    │
│  6. Retrieve outputs                     │
│  7. Validate results                     │
│  8. Decision: accept / retry / escalate  │
└──────────────┬───────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│         Worker Pod (Isolated)           │
│                                         │
│  ┌────────────────────────────────┐   │
│  │  Claude Code CLI (yolo mode)   │   │
│  │                                 │   │
│  │  - Reads task from ConfigMap    │   │
│  │  - Reads AGENTS.md context      │   │
│  │  - Executes task autonomously   │   │
│  │  - Writes output to volume      │   │
│  └────────────────────────────────┘   │
└─────────────────────────────────────────┘
               │
               │ output
               ▼
         ┌──────────┐
         │  Volume  │
         │ (output) │
         └──────────┘
               │
               ▼ read back
         [Manager validates]
               │
        ┌──────┴────────┐
        │               │
     Accept          Retry
        │               │
        ▼               └──> (spawn new pod with feedback)
   [Commit to host]
```

### Detailed Flow

**Phase 1: Task Submission**

```bash
aidu-session orchestrate "implement user authentication"
```

**Phase 2: Manager Analysis**

1. Parse task text
2. Run through complexity analyzer
3. Determine: IMPLEMENT
4. Load AGENTS.md for project context
5. Prepare task context package

**Phase 3: Worker Spawn**

1. Generate unique worker ID: `worker-auth-abc123`
2. Create ConfigMap with task + AGENTS.md
3. Create PVC for output storage
4. Generate pod spec from template
5. Submit to K8s in `aidu-sessions` namespace
6. Wait for pod Ready state

**Phase 4: Execution**

1. Pod starts, Claude Code CLI initializes
2. CLI reads task from `/task/input.md`
3. CLI reads context from `/task/AGENTS.md`
4. CLI executes in yolo mode (no confirmations)
5. CLI writes outputs to `/output/`
6. Pod signals completion (exit code)

**Phase 5: Validation**

1. Manager retrieves output from PVC
2. Run validation checks:
   - Parse output structure
   - Check for error indicators
   - Verify expected artifacts present
   - Analyze completeness
3. Calculate confidence score

**Phase 6: Decision**

```txt
IF score >= threshold:
    - Accept output
    - Copy to host machine
    - Clean up worker pod
    - Report success to user
ELSE IF retry_count < max_retries:
    - Analyze failure reason
    - Generate refined task context
    - Spawn new worker with feedback
    - Increment retry_count
ELSE:
    - Escalate to user
    - Preserve all worker outputs for debugging
    - Request human intervention
```

## Directory Structure

```bash
aidu/
├── cmd/
│   └── aidu-manager/          # Manager entry point
│       └── main.go
│
├── pkg/
│   ├── analyzer/              # Task complexity analysis
│   │   ├── analyzer.go
│   │   └── heuristics.go
│   │
│   ├── manager/               # Core manager logic
│   │   ├── manager.go         # Main orchestration
│   │   ├── pod_builder.go     # Pod spec generation
│   │   └── lifecycle.go       # Pod lifecycle management
│   │
│   ├── validator/             # Output validation
│   │   ├── validator.go
│   │   ├── checks.go
│   │   └── scoring.go
│   │
│   ├── config/                # Configuration management
│   │   ├── agents.go          # AGENTS.md loader
│   │   ├── task.go            # Task context builder
│   │   └── k8s.go             # K8s client config
│   │
│   └── types/                 # Shared types
│       ├── task.go
│       ├── worker.go
│       └── result.go
│
├── internal/
│   └── k8s/                   # K8s client utilities
│       ├── client.go
│       └── helpers.go
│
├── k8s/
│   ├── worker-pod-template.yaml    # Worker pod template
│   └── configmap-template.yaml     # Task context template
│
├── scripts/
│   └── integrate-manager.sh        # Integration with aidu-session
│
├── AGENTS.md                  # Default agent context
├── ARCHITECTURE.md            # This file
├── go.mod
└── go.sum
```

## Configuration

### AGENTS.md Structure

```markdown
# Agent Context for [Project Name]

## Project Overview
Brief description of what this project does.

## Architecture
Key architectural decisions and patterns.

## Coding Standards
- Language-specific best practices
- Naming conventions
- Error handling patterns

## Domain Knowledge
- Business rules
- Key concepts
- Important constraints

## Testing Requirements
- Expected test coverage
- Testing frameworks
- Test patterns

## Success Criteria
What defines a successful implementation.
```

### Manager Configuration

```go
type ManagerConfig struct {
    Namespace       string        // K8s namespace (default: aidu-sessions)
    MaxRetries      int           // Max worker spawns per task (default: 3)
    Timeout         time.Duration // Worker execution timeout (default: 10m)
    ValidationScore float64       // Acceptance threshold (default: 0.75)
    WorkerImage     string        // Container image (default: aidu-claude:latest)
    StorageClass    string        // PVC storage class
}
```

## Integration with Existing System

### Enhanced aidu-session CLI

```bash
# Existing commands continue to work
aidu-session create my-session
aidu-session connect my-session

# New orchestration command
aidu-session orchestrate [task] [flags]

Flags:
    --agents-file    Path to AGENTS.md (default: ./AGENTS.md)
    --max-retries    Maximum worker retries (default: 3)
    --timeout        Worker timeout (default: 10m)
    --interactive    Ask for approval at each retry
```

### Workflow Integration

```bash
# Example: Implement a feature
cd /my/project
cat > AGENTS.md <<EOF
# My Project Agent Context
## Overview
This is a REST API in Go...
EOF

aidu-session orchestrate "implement JWT authentication middleware"

# Manager spawns worker pod
# Worker produces implementation
# Manager validates and iterates if needed
# Final output committed to current directory
```

## Worker Pod Specification

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: worker-{TASK_ID}
  namespace: aidu-sessions
  labels:
    app: aidu-worker
    manager: aidu-orchestrator
spec:
  restartPolicy: Never
  containers:
  - name: claude-code
    image: aidu-claude:latest
    command: ["/bin/bash", "-c"]
    args:
      - |
        claude --yolo chat < /task/input.md > /output/result.md 2> /output/logs.txt
        echo $? > /output/exit_code
    volumeMounts:
    - name: task-context
      mountPath: /task
      readOnly: true
    - name: output
      mountPath: /output
    resources:
      limits:
        memory: "4Gi"
        cpu: "2"
  volumes:
  - name: task-context
    configMap:
      name: task-{TASK_ID}
  - name: output
    persistentVolumeClaim:
      claimName: worker-{TASK_ID}-output
```

## Future Enhancements

### Phase 2: Specialized Workers

- Golang environment worker with build/test capabilities
- Python environment worker
- Database-enabled workers

### Phase 3: Worker Communication

- Workers can collaborate on complex tasks
- Pair programming mode between workers

### Phase 4: Advanced Validation

- LLM-based output validation
- Test execution validation
- Security scanning integration

### Phase 5: Resource Optimization

- Worker pool management
- Reusable worker pods
- Incremental context updates

## Security Considerations

1. **Isolation**: Workers run in isolated namespaces with no network access by default
2. **Resource Limits**: CPU and memory limits prevent resource exhaustion
3. **Yolo Mode Safety**: Safe because environments are disposable and isolated
4. **Output Sanitization**: Manager validates outputs before committing to host
5. **Secret Management**: Sensitive data injected via Kubernetes secrets, not ConfigMaps

## Monitoring and Observability

### Metrics to Track

- Task success rate
- Average retries per task
- Worker execution time
- Validation scores
- Resource utilization

### Logging Strategy

- Manager logs: Orchestration decisions and validation results
- Worker logs: Preserved in output volume for debugging
- Structured logging with task correlation IDs

## Error Handling

### Manager Errors

- K8s API failures → Retry with exponential backoff
- Timeout → Kill worker pod, retry with reduced scope
- Validation failure → Spawn new worker with feedback

### Worker Errors

- Exit code != 0 → Capture logs, analyze error, retry
- OOM killed → Reduce task scope or increase resources
- Timeout → Simplify task or extend timeout

### User Escalation

- Max retries exceeded → Present all attempts to user
- Ambiguous task → Request clarification
- Resource constraints → Suggest task breakdown

## Success Metrics

A task is considered successful when:

1. Worker completes without errors
2. Expected artifacts are present
3. Validation score >= threshold
4. Output format matches expectations
5. No security concerns detected

## Development Roadmap

### MVP (Phase 1)

- [x] Architecture design
- [ ] Manager implementation
- [ ] Task analyzer with simple heuristics
- [ ] Pod lifecycle management
- [ ] Output validator
- [ ] Integration with aidu-session CLI

### Phase 2

- [ ] Enhanced validation logic
- [ ] Interactive retry approval
- [ ] Better error analysis
- [ ] Performance optimizations

### Phase 3

- [ ] Specialized worker environments
- [ ] Worker-to-worker communication
- [ ] Advanced monitoring dashboard
- [ ] Cost optimization features
