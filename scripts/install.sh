#!/bin/sh
set -eu

# ──────────────────────────────────────────────
# 🍒 Lychee — One-Liner Install Script
# Installs the Lychee CLI binary on Linux & macOS.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/install.sh | sh
#
# Or specify a version:
#   curl -fsSL https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/install.sh | LYCHEE_VERSION=v0.1.0 sh
# ──────────────────────────────────────────────

# ── Helpers ───────────────────────────────────
RED="$( (tput bold 2>/dev/null || :; tput setaf 1 2>/dev/null || :) )"
GREEN="$( (tput bold 2>/dev/null || :; tput setaf 2 2>/dev/null || :) )"
YELLOW="$( (tput setaf 3 2>/dev/null || :) )"
CYAN="$( (tput setaf 6 2>/dev/null || :) )"
PLAIN="$( (tput sgr0 2>/dev/null || :) )"

log()     { printf '%s\n' ">>> $*" >&2; }
success() { printf '%s\n' "${GREEN}✔${PLAIN} $*" >&2; }
warn()    { printf '%s\n' "${YELLOW}⚠${PLAIN} $*" >&2; }
err()     { printf '%s\n' "${RED}✘${PLAIN} $*" >&2; exit 1; }

TEMP_DIR="$(mktemp -d)"
cleanup() { rm -rf "$TEMP_DIR"; }
trap cleanup EXIT

available() { command -v "$1" >/dev/null 2>&1; }

# ── OS / Arch detection ───────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)             err "Unsupported architecture: $ARCH" ;;
esac

case "$OS" in
    Linux)  OS="linux" ;;
    Darwin) OS="darwin" ;;
    *)      err "Unsupported OS: $OS (only Linux and macOS are supported)" ;;
esac

log "Detected: ${OS}/${ARCH}"

# ── Determine install path ─────────────────────
INSTALL_DIR="${LYCHEE_INSTALL_DIR:-/usr/local/bin}"

# ── Install via Go (preferred if available) ────
if available go; then
    GO_VERSION="$(go version 2>/dev/null | grep -oP 'go\K[0-9]+\.[0-9]+' || echo "0.0")"
    GO_MAJOR="$(echo "$GO_VERSION" | cut -d. -f1)"
    GO_MINOR="$(echo "$GO_VERSION" | cut -d. -f2)"

    if [ "$GO_MAJOR" -gt 1 ] || { [ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -ge 22 ]; }; then
        log "Go ${GO_VERSION} detected — installing via 'go install'..."

        VERSION_FLAG="@latest"
        if [ -n "${LYCHEE_VERSION:-}" ]; then
            VERSION_FLAG="@${LYCHEE_VERSION}"
        fi

        go install "github.com/MD-Mushfiqur123/lychee${VERSION_FLAG}"

        GOPATH="$(go env GOPATH 2>/dev/null || echo "$HOME/go")"
        LYCHEE_BIN="${GOPATH}/bin/lychee"

        if [ -x "$LYCHEE_BIN" ]; then
            if [ "$LYCHEE_BIN" != "${INSTALL_DIR}/lychee" ]; then
                if [ -w "$INSTALL_DIR" ]; then
                    ln -sf "$LYCHEE_BIN" "${INSTALL_DIR}/lychee"
                else
                    sudo ln -sf "$LYCHEE_BIN" "${INSTALL_DIR}/lychee"
                fi
            fi
            success "Lychee installed via Go — $(lychee version 2>/dev/null || echo "done")"
            exit 0
        fi
        warn "go install succeeded but binary not found at ${LYCHEE_BIN} — falling back to binary download..."
    else
        warn "Go ${GO_VERSION} detected but 1.22+ required — falling back to binary download..."
    fi
fi

# ── Install via pre-built binary ───────────────
log "Installing via pre-built binary..."

if ! available curl; then
    err "curl is required but not installed. Install curl and try again."
fi

REPO="MD-Mushfiqur123/lychee"
if [ -n "${LYCHEE_VERSION:-}" ]; then
    RELEASE_URL="https://github.com/${REPO}/releases/download/${LYCHEE_VERSION}"
else
    RELEASE_URL="https://github.com/${REPO}/releases/latest/download"
fi

BINARY_NAME="lychee-${OS}-${ARCH}"
ARCHIVE_EXT=""
DOWNLOAD_URL=""

# Try different archive formats
for EXT in ".tar.gz" ".tgz" ".tar.zst" ""; do
    if [ -n "$EXT" ]; then
        TEST_URL="${RELEASE_URL}/${BINARY_NAME}${EXT}"
    else
        TEST_URL="${RELEASE_URL}/${BINARY_NAME}"
    fi

    if curl --fail --silent --head --location "$TEST_URL" >/dev/null 2>&1; then
        DOWNLOAD_URL="$TEST_URL"
        if [ -n "$EXT" ]; then
            ARCHIVE_EXT="$EXT"
        fi
        break
    fi
done

if [ -z "$DOWNLOAD_URL" ]; then
    err "No pre-built binary found for ${OS}/${ARCH} at ${RELEASE_URL}\n       Try installing via Go:  go install github.com/MD-Mushfiqur123/lychee@latest"
fi

log "Downloading ${DOWNLOAD_URL} ..."
curl --fail --show-error --location --progress-bar -o "${TEMP_DIR}/lychee${ARCHIVE_EXT}" "$DOWNLOAD_URL"

# Extract if archive
if [ -n "$ARCHIVE_EXT" ]; then
    case "$ARCHIVE_EXT" in
        .tar.gz|.tgz)
            tar -xzf "${TEMP_DIR}/lychee${ARCHIVE_EXT}" -C "$TEMP_DIR"
            ;;
        .tar.zst)
            if ! available zstd; then
                err "zstd is required to extract ${ARCHIVE_EXT}. Install zstd and try again."
            fi
            zstd -d < "${TEMP_DIR}/lychee${ARCHIVE_EXT}" | tar -xf - -C "$TEMP_DIR"
            ;;
    esac
    # Find the lychee binary in extracted files
    find "$TEMP_DIR" -type f -name "lychee" -exec mv {} "${TEMP_DIR}/lychee" \; 2>/dev/null || true
else
    mv "${TEMP_DIR}/lychee" "${TEMP_DIR}/lychee.bin" 2>/dev/null || true
    mv "${TEMP_DIR}/lychee.bin" "${TEMP_DIR}/lychee" 2>/dev/null || true
fi

chmod +x "${TEMP_DIR}/lychee" 2>/dev/null || true

if [ ! -x "${TEMP_DIR}/lychee" ]; then
    err "Failed to extract binary from download."
fi

# Install to destination
if [ -w "$INSTALL_DIR" ]; then
    mv -f "${TEMP_DIR}/lychee" "${INSTALL_DIR}/lychee"
else
    sudo mv -f "${TEMP_DIR}/lychee" "${INSTALL_DIR}/lychee"
fi

# ── Verify ─────────────────────────────────────
if available lychee; then
    success "Lychee installed successfully!"
    lychee version 2>/dev/null || true
    echo ""
    log "Run 'lychee serve' to start the server."
    log "Run 'lychee --help' for all commands."
else
    warn "Lychee installed to ${INSTALL_DIR} but it's not in your PATH."
    log "Add it to PATH or run:  ${INSTALL_DIR}/lychee"
fi

exit 0
