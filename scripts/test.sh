#!/usr/bin/env bash
set -euo pipefail

# Run tests for durablecompose
# Usage: ./scripts/test.sh [test-flags]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${PROJECT_ROOT}"

# Run unit tests
echo "Running unit tests..."
go test -v -race -coverprofile=coverage.out -covermode=atomic "$@" ./...

# Display coverage summary
echo ""
echo "Coverage summary:"
go tool cover -func=coverage.out | tail -1
