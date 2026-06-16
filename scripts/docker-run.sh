#!/bin/sh
# =============================================================================
# Lychee — Quick Docker Run Script (Linux / macOS / WSL)
# =============================================================================
# Pulls and runs the latest lychee image from GitHub Container Registry.
# Models are stored in a named Docker volume (lychee-models) for persistence.
#
# Usage:
#   chmod +x scripts/docker-run.sh
#   ./scripts/docker-run.sh
# =============================================================================

docker run -d \
  -p 11434:11434 \
  -v lychee-models:/root/.lychee/models \
  --name lychee \
  ghcr.io/md-mushfiqur123/lychee:latest
