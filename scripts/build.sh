#!/usr/bin/env bash
set -euo pipefail

# Build durablecompose binary
# Usage: ./scripts/build.sh [output-path]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

OUTPUT_PATH="${1:-${PROJECT_ROOT}/bin/build/durablecompose}"
VERSION="${VERSION:-$(git describe --match 'v[0-9]*' --dirty='.m' --always --tags)}"

echo "Building durablecompose ${VERSION}..."

cd "${PROJECT_ROOT}"

GO111MODULE=on go build \
  -trimpath \
  -tags "e2e" \
  -ldflags "-w -X github.com/durable_oss/durablecompose/internal.Version=${VERSION}" \
  -o "${OUTPUT_PATH}" \
  ./cmd

echo "Build complete: ${OUTPUT_PATH}"
