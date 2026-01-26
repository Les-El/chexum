#!/bin/bash
# test-with-cleanup.sh
# Wrapper script to run tests with automatic cleanup
# Usage: bash scripts/test-with-cleanup.sh [test-args...]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

echo "=== Pre-test Cleanup ==="
bash scripts/cleanup-before-tests.sh || true

echo ""
echo "=== Running Tests ==="

# Default to ./... if no args provided
TEST_ARGS="${@:-./.../cmd...,./...internal}"

# Run tests with the provided arguments or defaults
if [ $# -eq 0 ]; then
    echo "Running: go test ./cmd/... ./internal/checkpoint"
    go test ./cmd/... ./internal/checkpoint
else
    echo "Running: go test $@"
    go test "$@"
fi

EXIT_CODE=$?

echo ""
echo "=== Post-test Cleanup ==="
bash scripts/cleanup-before-tests.sh || true

exit $EXIT_CODE
