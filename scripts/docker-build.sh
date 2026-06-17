#!/bin/sh
# Multi-arch Docker build script
# Builds for linux/amd64 and linux/arm64
# Pushes to GitHub Container Registry (ghcr.io)

set -e

IMAGE="ghcr.io/md-mushfiqur123/lychee"
VERSION="${1:-latest}"

echo "Building Lychee Docker image: $IMAGE:$VERSION"

# Build multi-arch
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  --tag "$IMAGE:$VERSION" \
  --tag "$IMAGE:latest" \
  --push \
  .

echo "✅ Pushed: $IMAGE:$VERSION"
