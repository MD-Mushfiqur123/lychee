# Building Lychee with Docker

This project uses a **Docker multi-stage build** that compiles Lychee with full CGO support for llama.cpp-based local LLM inference.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed and running
- At least 4 GB of free disk space for the build

## Quick Start

### 1. Build the image

```bash
docker build -t lychee .
```

The build process:
1. **Stage 1 (builder):** Uses `golang:1.25-alpine` with gcc, g++, musl-dev, cmake, make, and linux-headers to compile Lychee with `CGO_ENABLED=1`.
2. **Stage 2 (runtime):** Produces a minimal `alpine:latest` image containing only the compiled binary and CA certificates.

### 2. Run the server

```bash
docker run -p 11434:11434 lychee
```

This starts the Lychee server on port 11434 (the default Ollama-compatible port).

### 3. Verify it works

```bash
curl http://localhost:11434/api/tags
```

## Custom Build Options

### Build for a specific platform

```bash
docker build --platform linux/amd64 -t lychee .
docker build --platform linux/arm64 -t lychee .
```

### Run in detached mode

```bash
docker run -d -p 11434:11434 --name lychee lychee
```

### Mount a volume for persistent model storage

```bash
docker run -d \
  -p 11434:11434 \
  -v lychee-models:/root/.lychee/models \
  --name lychee \
  lychee
```

### View logs

```bash
docker logs -f lychee
```

### Stop the container

```bash
docker stop lychee
docker rm lychee
```

## Why CGO?

Lychee embeds **llama.cpp**, a C++ inference engine for large language models. The Go bindings to llama.cpp require **CGO** (cgo) to link against the compiled C++ library at build time. Without `CGO_ENABLED=1`, the resulting binary would be unable to load or run any models.

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `CGO_ENABLED=0` error on build | Ensure the Dockerfile includes `apk add gcc g++ musl-dev cmake make linux-headers` in the build stage |
| `exec format error` at runtime | You may have built for the wrong architecture. Use `--platform` to match your host |
| Port already in use | Change the host port: `docker run -p 8080:11434 lychee` |
| Out of memory during build | Increase Docker's memory limit in Docker Desktop settings |
