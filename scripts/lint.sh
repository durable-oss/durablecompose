#!/usr/bin/env bash
set -euo pipefail

# Run linters on durablecompose codebase
# Usage: ./scripts/lint.sh

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${PROJECT_ROOT}"

# Check if gofumpt is installed
if ! command -v gofumpt &> /dev/null; then
    echo "gofumpt not found, installing..."
    go install mvdan.cc/gofumpt@latest
fi

echo "Running gofumpt..."
gofumpt -l -w .

echo "Running go vet..."
go vet ./...

echo "Running staticcheck..."
if ! command -v staticcheck &> /dev/null; then
    echo "staticcheck not found, installing..."
    go install honnef.co/go/tools/cmd/staticcheck@latest
fi
staticcheck ./...

echo "Linting complete!"
