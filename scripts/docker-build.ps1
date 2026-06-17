# Multi-arch Docker build script (PowerShell)
# Builds for linux/amd64 and linux/arm64
# Pushes to GitHub Container Registry (ghcr.io)

param(
    [string]$Version = "latest"
)

$ErrorActionPreference = "Stop"

$IMAGE = "ghcr.io/md-mushfiqur123/lychee"

Write-Host "Building Lychee Docker image: ${IMAGE}:${Version}"

# Build multi-arch
docker buildx build `
  --platform linux/amd64,linux/arm64 `
  --tag "${IMAGE}:${Version}" `
  --tag "${IMAGE}:latest" `
  --push `
  .

Write-Host "✅ Pushed: ${IMAGE}:${Version}"
