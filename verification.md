# Lychee Verification Checklist

Use this checklist to verify that your Lychee build is working correctly.
Run each command and check the box when it passes.

---

## Prerequisites

- [ ] **Go installed (≥1.21)**
  ```powershell
  go version
  ```
  Expected: `go version go1.21.x windows/amd64` (or newer)

- [ ] **CGO enabled**
  ```powershell
  go env CGO_ENABLED
  ```
  Expected: `1` (CGO is required for the Go ↔ C bridge to llama.cpp)

- [ ] **GCC available (for CGO builds on Windows)**
  ```powershell
  gcc --version
  ```
  Expected: version output (MinGW-w64 or MSYS2 GCC recommended)

---

## Build

- [ ] **All packages compile**
  ```powershell
  go build ./...
  ```
  Expected: no errors output

- [ ] **Binary exists and is executable**
  ```powershell
  Test-Path .\lychee.exe          # Windows
  # or
  test -f ./lychee                 # Linux / macOS
  ```
  Expected: `True` (file exists)

- [ ] **Version command works**
  ```powershell
  .\lychee.exe --version
  ```
  Expected: version string displayed (e.g., `lychee version 0.x.x`)

---

## API Server

- [ ] **Server starts**
  ```powershell
  Start-Process .\lychee.exe -ArgumentList "serve","--port","11434" -WindowStyle Hidden
  Start-Sleep 3
  ```
  Expected: no crash; process running in background

- [ ] **`GET /api/tags` responds**
  ```powershell
  Invoke-RestMethod -Uri "http://localhost:11434/api/tags" -Method Get
  ```
  Expected: JSON with `models` array (may be empty if no models pulled)

- [ ] **`GET /api/tags` returns models (after pulling)**
  ```powershell
  # Pull a tiny test model first:
  $body = @{ name = "orca-mini:3b"; stream = $false } | ConvertTo-Json
  Invoke-RestMethod -Uri "http://localhost:11434/api/pull" -Method Post -Body $body -ContentType "application/json"

  # Verify it appears:
  Invoke-RestMethod -Uri "http://localhost:11434/api/tags" -Method Get
  ```
  Expected: `orca-mini:3b` appears in the models list

---

## Model Inference

- [ ] **Text generation works**
  ```powershell
  $body = @{ model = "orca-mini:3b"; prompt = "What is 2+2?"; stream = $false } | ConvertTo-Json
  $result = Invoke-RestMethod -Uri "http://localhost:11434/api/generate" -Method Post -Body $body -ContentType "application/json"
  $result.response
  ```
  Expected: model responds with text containing "4" or an answer

- [ ] **Streaming generation works**
  ```powershell
  $body = @{ model = "orca-mini:3b"; prompt = "Hi" } | ConvertTo-Json
  Invoke-WebRequest -Uri "http://localhost:11434/api/generate" -Method Post -Body $body -ContentType "application/json"
  ```
  Expected: streaming JSON lines returned (each with `response` field)

---

## Docker (if Docker is available)

- [ ] **Docker daemon running**
  ```powershell
  docker info
  ```
  Expected: Docker system information displayed

- [ ] **Docker image builds**
  ```powershell
  docker build -t lychee:local .
  ```
  Expected: image built successfully, no errors

- [ ] **Docker container runs**
  ```powershell
  docker run -d --name lychee-test -p 11435:11434 lychee:local
  Start-Sleep 5
  Invoke-RestMethod -Uri "http://localhost:11435/api/tags" -Method Get
  docker stop lychee-test
  docker rm lychee-test
  ```
  Expected: tags response returned, container cleans up

---

## CI/CD (GitHub Actions)

- [ ] **Windows build workflow exists**
  `.github/workflows/build-windows.yml`
  Expected: builds with CGO, uploads artifacts

- [ ] **Linux build workflow exists**
  `.github/workflows/build-linux.yml`
  Expected: cross-compiles with CGO, uploads artifacts

- [ ] **macOS build workflow exists**
  `.github/workflows/build-darwin.yml`
  Expected: builds universal binary (amd64 + arm64)

- [ ] **Docker multi-arch workflow exists**
  `.github/workflows/docker-build.yml`
  Expected: builds linux/amd64 + linux/arm64 images

- [ ] **CI unit-test workflow exists**
  `.github/workflows/ci.yml`
  Expected: runs `go test ./...` on push/PR

---

## Quick One-Command Smoke Test

```powershell
.\scripts\smoke-test.ps1
```

This script automates steps 4-6 above: starts the server, tests `/api/tags` and `/api/generate`, then cleans up. Pass/Fail reported at end.

---

## Status Summary

| Check                    | Status |
|--------------------------|--------|
| Go version               | ⬜     |
| CGO enabled              | ⬜     |
| `go build ./...`         | ⬜     |
| Binary exists            | ⬜     |
| API `/api/tags`          | ⬜     |
| Model inference          | ⬜     |
| Docker build             | ⬜     |

_Last verified: 2026-06-16_
