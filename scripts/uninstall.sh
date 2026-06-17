#!/bin/sh
set -eu

# ──────────────────────────────────────────────
# 🍒 Lychee — Uninstall Script
# Removes the Lychee CLI binary, data, and
# services from Linux and macOS.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/uninstall.sh | sh
# ──────────────────────────────────────────────

RED="$( (tput bold 2>/dev/null || :; tput setaf 1 2>/dev/null || :) )"
GREEN="$( (tput bold 2>/dev/null || :; tput setaf 2 2>/dev/null || :) )"
YELLOW="$( (tput setaf 3 2>/dev/null || :) )"
PLAIN="$( (tput sgr0 2>/dev/null || :) )"

log()     { printf '%s\n' ">>> $*" >&2; }
success() { printf '%s\n' "${GREEN}✔${PLAIN} $*" >&2; }
warn()    { printf '%s\n' "${YELLOW}⚠${PLAIN} $*" >&2; }
err()     { printf '%s\n' "${RED}✘${PLAIN} $*" >&2; exit 1; }

SUDO=""
if [ "$(id -u)" -ne 0 ]; then
    if command -v sudo >/dev/null 2>&1; then
        SUDO="sudo"
    else
        warn "Not running as root and sudo not available — some items may not be removable."
    fi
fi

# ── Stop services ──────────────────────────────
log "Stopping Lychee service (if running)..."

if command -v systemctl >/dev/null 2>&1; then
    $SUDO systemctl stop lychee 2>/dev/null || true
    $SUDO systemctl disable lychee 2>/dev/null || true
    if [ -f /etc/systemd/system/lychee.service ]; then
        $SUDO rm -f /etc/systemd/system/lychee.service
        $SUDO systemctl daemon-reload 2>/dev/null || true
        success "Systemd service removed."
    fi
fi

# Kill any running lychee processes
if command -v pkill >/dev/null 2>&1; then
    pkill -x lychee 2>/dev/null || true
fi

# ── Remove binary ──────────────────────────────
log "Removing Lychee binary..."

BINS_REMOVED=0
for dir in /usr/local/bin /usr/bin /usr/local/sbin; do
    if [ -f "${dir}/lychee" ]; then
        $SUDO rm -f "${dir}/lychee"
        success "Removed ${dir}/lychee"
        BINS_REMOVED=$((BINS_REMOVED + 1))
    fi
done

# ── Remove Go-installed binary ─────────────────
if command -v go >/dev/null 2>&1; then
    GOPATH="$(go env GOPATH 2>/dev/null || echo "${HOME}/go")"
    if [ -f "${GOPATH}/bin/lychee" ]; then
        rm -f "${GOPATH}/bin/lychee"
        success "Removed Go-installed binary: ${GOPATH}/bin/lychee"
        BINS_REMOVED=$((BINS_REMOVED + 1))
    fi

    log "Cleaning Go module cache (optional)..."
    go clean -modcache 2>/dev/null || warn "Could not clean Go mod cache — you can skip this."
fi

# ── Remove data directory ──────────────────────
DATA_DIR="${HOME}/.lychee"
if [ -d "$DATA_DIR" ]; then
    log "Lychee data directory found at ${DATA_DIR}"
    printf "Remove data directory (includes downloaded models)? [y/N] " >&2
    read -r CONFIRM
    case "$CONFIRM" in
        [Yy]*)
            rm -rf "$DATA_DIR"
            success "Removed data directory: ${DATA_DIR}"
            ;;
        *)
            warn "Data directory kept: ${DATA_DIR}"
            log "You can remove it manually later:  rm -rf ~/.lychee"
            ;;
    esac
else
    log "No data directory found at ${DATA_DIR}"
fi

# ── Remove model directory (if exists) ─────────
MODEL_DIR="${HOME}/.lychee-models"
if [ -d "$MODEL_DIR" ]; then
    log "Model cache found at ${MODEL_DIR}"
    printf "Remove model cache? [y/N] " >&2
    read -r CONFIRM
    case "$CONFIRM" in
        [Yy]*)
            rm -rf "$MODEL_DIR"
            success "Removed model cache: ${MODEL_DIR}"
            ;;
        *)
            warn "Model cache kept: ${MODEL_DIR}"
            ;;
    esac
fi

# ── Remove macOS .app bundle ───────────────────
if [ "$(uname -s)" = "Darwin" ]; then
    if [ -d "/Applications/Lychee.app" ]; then
        log "Lychee Desktop app found."
        printf "Remove Lychee Desktop app? [y/N] " >&2
        read -r CONFIRM
        case "$CONFIRM" in
            [Yy]*)
                rm -rf "/Applications/Lychee.app"
                success "Removed Lychee Desktop app."
                ;;
            *)
                warn "Desktop app kept."
                ;;
        esac
    fi
fi

# ── Summary ────────────────────────────────────
if [ $BINS_REMOVED -gt 0 ]; then
    success "Lychee has been uninstalled."
else
    warn "Lychee binary not found — it may already be uninstalled, or installed in a non-standard location."
fi

log "Done."
exit 0
