# Linux Build Guide

## Overview

Lychee builds for Linux with **CGO enabled** to support the embedded `llama.cpp` inference engine. This enables native CPU inference without any external runtime dependencies.

## Prerequisites

### Build Dependencies

```bash
sudo apt-get update
sudo apt-get install -y gcc g++ make cmake
```

### Go

Go **1.25+** is required:

```bash
# Via official Go installer
wget https://go.dev/dl/go1.25.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

## Building

### Native (amd64)

```bash
CGO_ENABLED=1 go build -ldflags="-s -w" -o dist/lychee .
```

### Cross-compile for arm64

Install the ARM64 cross-compilation toolchain first:

```bash
sudo apt-get install -y gcc-aarch64-linux-gnu g++-aarch64-linux-gnu
```

Then build:

```bash
CC=aarch64-linux-gnu-gcc \
CXX=aarch64-linux-gnu-g++ \
GOOS=linux GOARCH=arm64 CGO_ENABLED=1 \
  go build -ldflags="-s -w" -o dist/lychee .
```

### Build flags explained

| Flag | Purpose |
|------|---------|
| `CGO_ENABLED=1` | Enables CGo — required for llama.cpp |
| `-ldflags="-s -w"` | Strip debug info + symbol table → smaller binary |
| `CC=/CXX=` | Cross-compiler paths (arm64 only) |

## CI/CD

Linux builds are automated via `.github/workflows/build-linux.yml`:

| Trigger | What happens |
|---------|-------------|
| Push to `main` | Builds amd64 + arm64, uploads as artifacts |
| Tag `v*` pushed | Builds both archs, creates GitHub Release with tarballs |
| Manual `workflow_dispatch` | Builds both archs, uploads artifacts |

### Artifacts

- `lychee-linux-amd64` — x86_64 binary
- `lychee-linux-arm64` — aarch64 binary

All artifacts are available for 90 days from the Actions run.

## Release

On version tags (e.g., `v1.2.3`), the workflow:

1. Builds both amd64 and arm64 binaries
2. Creates tarballs
3. Attaches them to a GitHub Release with auto-generated release notes

Download the latest from the [Releases page](https://github.com/MD-Mushfiqur123/lychee/releases).

## System Requirements

| Requirement | Minimum |
|-------------|---------|
| Kernel | Linux 4.x+ |
| glibc | 2.31+ |
| RAM | 4 GB (model-dependent) |
| Disk | ~100 MB (binary + models) |
