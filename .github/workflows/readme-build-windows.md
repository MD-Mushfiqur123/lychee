# Windows CGO Build & Release Workflow

## File: `.github/workflows/build-windows.yml`

This GitHub Actions workflow builds Lychee for **Windows AMD64** with **CGO enabled**, required because Lychee links against native C/C++ libraries (llama.cpp, MLX) at build time.

## Triggers

| Trigger              | When                                        |
|----------------------|---------------------------------------------|
| `push` to `main`     | Every commit/merge to main                  |
| `push` tag `v*`      | Version tags (e.g. `v1.2.3`) → also creates a GitHub Release |
| `workflow_dispatch`  | Manual trigger from the Actions tab         |

## What it does

1. **Checks out** the repository
2. **Sets up Go** — reads the required Go version from `go.mod` (currently Go 1.25+)
3. **Installs C++ build tools** — downloads `llvm-mingw` (clang-based MinGW toolchain) which provides a gcc-compatible C/C++ compiler on Windows. This is the **same toolchain** used by Lychee's official release pipeline. Clang is required (not MSVC) for proper UTF-16 handling in CGO.
4. **Builds** the binary with CGO enabled:
   ```
   go build -ldflags="-s -w" -o dist/lychee-windows-amd64.exe .
   ```
   - `CGO_ENABLED=1` — links native llama.cpp/MLX libraries
   - `-s -w` — strips debug symbols and DWARF tables to reduce binary size
   - Optimizations: `CGO_CFLAGS=-O3`, `CGO_CXXFLAGS=-O3`
5. **Verifies** the binary exists and logs its size + SHA256 hash
6. **Uploads** the binary + a zip archive as a workflow artifact (available for 90 days)
7. **On release tags only** — attaches `lychee-windows-amd64.zip` to the GitHub Release using `softprops/action-gh-release@v2`

## Why CGO?

Lychee integrates with:
- **llama.cpp** — native C++ inference engine
- **MLX** — Apple ML framework (macOS) / CUDA bindings (Windows)

These are compiled from C/C++ source at build time via CGO. A pure-Go cross-compile (with `CGO_ENABLED=0`) would **not** include the inference runtime. The official release pipeline (`release.yaml`) uses a similar native build approach.

## Artifacts

After each run, download from the **Actions → Summary** page:

| File                                | Description                           |
|-------------------------------------|---------------------------------------|
| `lychee-windows-amd64.exe`          | Standalone Windows executable         |
| `lychee-windows-amd64.zip`          | Zipped executable (for release attachment) |

## Requirements

- GitHub-hosted runner `windows-latest` (Windows Server 2022 with 4 vCPUs, 16 GB RAM)
- Internet access to download llvm-mingw (~80 MB, one-time per run)

## Manual Trigger

Go to **Actions → Windows CGO Build + Release → Run workflow**, choose the branch, and click **Run workflow**.
