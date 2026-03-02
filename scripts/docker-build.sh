#!/usr/bin/env bash
set -euo pipefail

# Build durablecompose Docker image
# Usage: ./scripts/docker-build.sh [image-tag]

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

IMAGE_TAG="${1:-durablecompose:latest}"
VERSION="${VERSION:-$(git describe --match 'v[0-9]*' --dirty='.m' --always --tags)}"

echo "Building Docker image: ${IMAGE_TAG}"
echo "Version: ${VERSION}"

cd "${PROJECT_ROOT}"

docker buildx build \
  --build-arg VERSION="${VERSION}" \
  --tag "${IMAGE_TAG}" \
  --load \
  .

echo "Docker image built successfully: ${IMAGE_TAG}"
