# =============================================================================
# Lychee — Quick Docker Run Script (Windows PowerShell)
# =============================================================================
# Pulls and runs the latest lychee image from GitHub Container Registry.
# Models are stored in a named Docker volume (lychee-models) for persistence.
#
# Usage:
#   .\scripts\docker-run.ps1
# =============================================================================

docker run -d `
  -p 11434:11434 `
  -v lychee-models:/root/.lychee/models `
  --name lychee `
  ghcr.io/md-mushfiqur123/lychee:latest
