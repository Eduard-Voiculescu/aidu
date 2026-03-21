#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
BINARY="$PROJECT_DIR/bin/aidu"
NAMESPACE="aidu-sessions"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

fail() {
    echo -e "${RED}FAIL: $1${NC}"
    exit 1
}

pass() {
    echo -e "${GREEN}PASS: $1${NC}"
}

echo "=== AIDU Smoke Test ==="
echo ""

if ! k3d cluster list 2>/dev/null | grep -q "aidu-cluster"; then
    fail "k3d cluster 'aidu-cluster' is not running. Run ./setup-k3d.sh first."
fi
pass "k3d cluster is running"

if ! kubectl get secret aidu-api-key -n "$NAMESPACE" &>/dev/null; then
    fail "API key secret not found. Run 'bin/aidu setup' first."
fi
pass "API key secret exists"

if [[ ! -f "$BINARY" ]]; then
    fail "aidu binary not found at $BINARY. Run 'make build' first."
fi
pass "aidu binary exists"

echo ""
echo "Running task: What is 2+2? Reply with just the number."
echo "---"

"$BINARY" run "What is 2+2? Reply with just the number." --namespace "$NAMESPACE"
RUN_EXIT=$?

if [[ $RUN_EXIT -ne 0 ]]; then
    fail "aidu run exited with code $RUN_EXIT"
fi
pass "aidu run completed successfully"

LOG_FILE=$(ls -t "$PROJECT_DIR/logs/"aidu-*.log 2>/dev/null | head -1)
if [[ -z "$LOG_FILE" ]]; then
    fail "No log file found in $PROJECT_DIR/logs/"
fi
pass "Log file created: $LOG_FILE"

if [[ ! -s "$LOG_FILE" ]]; then
    fail "Log file is empty"
fi
pass "Log file has content"

echo ""
echo "=== Log output ==="
cat "$LOG_FILE"
echo ""
echo "=== All smoke tests passed ==="
