# Lychee Architecture

Lychee is a local-first LLM server and CLI toolkit built in Go. It wraps the `llama.cpp` C++ inference engine via CGO, provides OpenAI- and Anthropic-compatible REST APIs, and includes its own model packaging system based on OCI-like manifests.

---

## Project Tree

```
lychee/
├── main.go                    # Entrypoint: cobra CLI
│
├── cmd/                       # CLI commands (cobra-based)
│   ├── cmd.go                 # Root command, TUI selector wiring
│   ├── cmd_serve.go           # `lychee serve` — starts HTTP server
│   ├── cmd_run.go             # `lychee run` — one-shot prompt
│   ├── cmd_pull.go            # `lychee pull` — download models
│   ├── cmd_create.go          # `lychee create` — build from Modelfile
│   ├── cmd_list.go            # `lychee list` — list local models
│   ├── cmd_show.go            # `lychee show` — inspect model details
│   ├── cmd_delete.go          # `lychee delete` — remove models
│   ├── cmd_copy.go            # `lychee cp` — copy models
│   ├── cmd_push.go            # `lychee push` — push to registry
│   ├── cmd_embed.go           # `lychee embed` — generate embeddings
│   ├── cmd_bench.go/          # Benchmarking
│   ├── cmd_benchmark.go
│   ├── cmd_compare.go         # Model comparison
│   ├── cmd_compose.go         # Multi-step model pipelines
│   ├── cmd_completion.go      # Shell completions
│   ├── cmd_export.go          # Conversations export
│   ├── cmd_import.go          # Conversations import
│   ├── cmd_modelfile.go       # Modelfile creation helper
│   ├── cmd_publish.go         # Community model publishing
│   ├── cmd_scan.go            # Scan models for issues
│   ├── cmd_search.go          # Search community models
│   ├── cmd_stats.go           # Show server stats
│   ├── cmd_stop.go            # Stop running models
│   ├── cmd_update.go          # Update models
│   ├── cmd_agent.go           # Agent subcommand
│   ├── cmd_auth.go            # Auth management
│   ├── cmd_catalog.go         # Model catalog
│   ├── cmd_community.go       # Community registry
│   ├── cmd_compare.go         # Model comparison
│   ├── cmd_interactive.go     # Interactive chat mode
│   ├── cmd_launcher_test.go   # Launcher integration tests
│   ├── cmd_quantize.go        # Model quantization
│   ├── start*.go              # Server startup helpers (platform-specific)
│   ├── hf.go, hf_catalog.go   # HuggingFace catalog integration
│   ├── interactive.go         # Terminal chat UI logic
│   ├── config/                # Configuration loading
│   ├── launch/                # Launcher integrations (Claude Code, VS Code, etc.)
│   ├── tui/                   # Terminal UI (Bubbletea-based selectors, sign-in)
│   ├── runner/                # Subprocess runner dispatch (imagegen, MLX)
│   ├── bench/                 # Benchmark framework
│   └── internal/fileutil/     # File utilities for commands
│
├── server/                    # HTTP server — the heart of Lychee
│   ├── routes.go              # Route registration, `Serve()`, `GenerateRoutes()`
│   ├── routes_openai.go       # OpenAI-compatible endpoints (`/v1/chat/completions`, etc.)
│   ├── routes_anthropic.go    # Anthropic-compatible endpoint (`/v1/messages`)
│   ├── handler_chat.go        # Chat completion handler
│   ├── handler_generate.go    # Text generation handler
│   ├── handler_embed.go       # Embedding handler
│   ├── handler_compose.go     # Multi-step compose handler
│   ├── handler_structured.go  # Structured output handler
│   ├── handler_router.go      # Model router handler
│   ├── handler_memory.go      # Conversation memory handler
│   ├── handler_admin.go       # Admin/status handlers
│   ├── handler_aliases.go     # Model alias management
│   ├── create.go              # Model creation from GGUF/safetensors
│   ├── download.go            # Model download orchestration
│   ├── upload.go              # Model upload to registries
│   ├── huggingface.go         # HuggingFace model discovery + download
│   ├── model.go               # Model abstraction, GGML parsing, layer management
│   ├── model_router.go        # Multi-backend model routing (local/cloud)
│   ├── model_resolver.go      # Model name resolution
│   ├── model_aliases.go       # Model alias system
│   ├── model_caches.go        # Cached model metadata
│   ├── model_list_cache.go    # Model list caching
│   ├── model_recommendations.go # Model recommendation engine
│   ├── model_show_cache.go    # Model detail caching
│   ├── prompt.go              # Prompt construction and template application
│   ├── prompt_cache.go        # Prompt caching
│   ├── memory_store.go        # Persistent conversation storage (JSON files)
│   ├── sched.go               # Model scheduler — loads/evicts GPU runners
│   ├── sched_vram.go          # VRAM-aware scheduling
│   ├── sched_evict.go         # Runner eviction logic
│   ├── sched_load.go          # Runner loading logic
│   ├── quantization.go        # Quantization registry
│   ├── structured_output.go   # JSON schema-driven structured output
│   ├── schema_validator.go    # JSON Schema validator
│   ├── renderer_resolution.go # Chat template renderer resolution
│   ├── images.go              # Image/media handling for multimodal
│   ├── logprob.go             # Log probability computation
│   ├── fixblobs.go            # Blob integrity repair
│   ├── sparse_*.go            # Sparse file support (Windows)
│   ├── auth.go                # ED25519 key-based auth
│   ├── dashboard.go           # Built-in dashboard UI
│   ├── cloud_proxy.go         # Cloud relay proxy
│   ├── inference_request_log.go # Per-request logging
│   ├── internal/
│   │   ├── cache/blob/        # Content-addressable blob storage
│   │   ├── client/lychee/     # Lychee registry client
│   │   ├── internal/backoff/  # Exponential backoff utility
│   │   ├── internal/names/    # Model name parsing
│   │   ├── internal/stringsx/ # Extended string utilities
│   │   ├── internal/syncs/    # Synchronization primitives
│   │   ├── manifest/          # OCI-compatible manifest management
│   │   ├── registry/          # Local model registry server
│   │   └── testutil/          # Test helpers
│   └── webui/                 # Embedded web UI (Svelte/React frontend)
│
├── api/                       # Go client library + shared API types
│   ├── client.go              # HTTP client for Lychee API
│   ├── types.go               # Core request/response types (Generate, Chat, Embed)
│   ├── types_compose.go       # Compose API types
│   ├── types_memory.go        # Conversation memory types
│   ├── types_router.go        # Model routing types
│   ├── types_structured.go    # Structured output types
│   └── examples/              # API usage examples (chat, generate, multimodal, etc.)
│
├── middleware/                # HTTP middleware for API compatibility
│   ├── openai.go              # Transform Lychee API ↔ OpenAI format
│   ├── anthropic.go           # Transform Lychee API ↔ Anthropic format
│   └── test_home_test.go      # Test helpers
│
├── llm/                       # LLM inference backend management
│   ├── server.go              # `LlamaServer` interface + configuration
│   ├── llama_server.go        # llama.cpp server subprocess management
│   ├── llama_binary.go        # llama.cpp binary discovery
│   ├── status.go              # Server health/status tracking
│   ├── media.go               # Multimodal media handling
│   ├── exit_status.go         # Process exit code handling
│   ├── metal_retry.go         # Apple Metal GPU retry logic
│   ├── rocm*.go               # AMD ROCm support
│   ├── vulkan_*.go            # Vulkan backend support
│   └── llm_*.go               # Platform-specific build constraints
│
├── discover/                  # Hardware discovery
│   ├── gpu.go                 # GPU enumeration, system memory discovery
│   ├── amd.go                 # AMD GPU detection
│   ├── cuda_compat.go         # CUDA compatibility detection
│   ├── vulkan.go              # Vulkan device detection
│   ├── cpu_*.go               # Platform-specific CPU info
│   ├── native_probe*.go       # Native GPU probing
│   ├── llama_server.go        # llama-server capability probing
│   ├── models_registry.go     # Hardware→model capability mapping
│   └── runner.go              # Runner selection logic
│
├── model/                     # Model architecture definitions
│   ├── model.go               # Core model interface
│   ├── chain.go               # Multi-model chaining
│   ├── input/input.go         # Unified input construction
│   ├── models/                # Full model implementations
│   │   └── gemma4/            # Gemma 4 multimodal model (text+vision+audio)
│   │   └── laguna/            # Laguna model
│   ├── parsers/               # Chat template parsers (per model family)
│   │   ├── parsers.go         # Parser registry + dispatcher
│   │   ├── deepseek3.go       # DeepSeek 3
│   │   ├── gemma4.go          # Gemma 4
│   │   ├── glm46.go, glm47.go # GLM-4-6, GLM-4-7
│   │   ├── qwen3.go, qwen35.go, qwen3vl.go # Qwen family
│   │   ├── olmo3.go, olmo3_think.go # OLMo 3
│   │   ├── lfm2.go            # LFM-2
│   │   ├── functiongemma.go   # FunctionGemma
│   │   ├── nemotron3nano.go   # Nemotron-3-Nano
│   │   ├── ministral.go       # Ministral
│   │   ├── cogito.go          # Cogito
│   │   └── laguna.go, glmocr.go, qwen3coder.go
│   └── renderers/             # Prompt template renderers
│       ├── renderer.go        # Renderer interface + registry
│       ├── json.go            # JSON schema template rendering
│       ├── image_tags.go      # Image tag handling
│       └── [matching files for each parser]
│
├── convert/                   # Model format conversion
│   ├── convert.go             # Conversion orchestrator
│   ├── reader.go              # Safetensors/PyTorch tensor readers
│   ├── tensor.go              # Tensor processing utilities
│   ├── tokenizer.go           # Tokenizer conversion (SentencePiece)
│   ├── json_compat.go         # JSON format compatibility
│   ├── convert_llama.go       # Llama family converter
│   ├── convert_mistral.go     # Mistral/Mixtral converter
│   ├── convert_gemma*.go      # Gemma family converters
│   ├── convert_qwen*.go       # Qwen family converters
│   ├── convert_deepseek2.go   # DeepSeek v2 converter
│   ├── convert_phi3.go        # Phi-3 converter
│   ├── convert_bert.go        # BERT converter
│   ├── convert_olmo.go        # OLMo converter
│   ├── convert_gptoss.go      # GPT-OSS converter
│   ├── convert_laguna.go      # Laguna converter
│   ├── convert_lfm2.go        # LFM-2 converter
│   ├── convert_nemotron_h.go  # Nemotron-H converter
│   ├── convert_commandr.go    # Command-R converter
│   ├── convert_mllama.go      # Multi-modal Llama converter
│   ├── convert_*.go           # Additional architecture converters
│   └── testdata/              # Reference model configs for testing
│
├── ml/                        # Machine learning primitives
│   ├── backend.go             # Compute backend interface
│   ├── device.go              # Device abstraction (GPU, CPU)
│   ├── path.go                # Model path resolution
│   ├── backend/ggml/          # GGML backend integration
│   └── nn/                    # Neural network layers
│       ├── attention.go, linear.go, embedding.go
│       ├── normalization.go, rope.go, convolution.go
│       └── pooling/
│
├── anthropic/                 # Anthropic API compatibility layer
│   ├── anthropic.go           # Types: messages, content blocks, tool use
│   ├── trace.go               # Request tracing
│   └── README.md
│
├── openai/                    # OpenAI API compatibility layer
│   ├── openai.go              # Types: chat completions, responses
│   └── responses.go           # Responses API types
│
├── x/                         # Extended / experimental features
│   ├── agent/                 # Agent sandboxing and approval
│   ├── imagegen/              # Image generation (Flux, ZImage via MLX)
│   ├── mlxrunner/             # MLX backend runner
│   ├── safetensors/           # Safetensors format support
│   ├── create/                # Advanced model creation (safetensors import)
│   ├── models/                # Experimental model implementations
│   ├── tokenizer/             # Tokenizer backends
│   ├── transfer/              # Model transfer utilities
│   ├── tools/                 # Tool integration
│   └── internal/              # Internal x/ utilities
│
├── sdk/                       # Language SDKs
│   ├── go/                    # Go SDK (minimal client)
│   ├── python/                # Python SDK (lychee package)
│   ├── javascript/            # JavaScript/TypeScript SDK
│   ├── ruby/                  # Ruby SDK (lychee gem)
│   ├── rust/                  # Rust SDK (lychee crate)
│   ├── js/                    # Legacy JS SDK
│   └── vscode/                # VS Code extension
│
├── types/                     # Shared type definitions
│   ├── model/                 # Model metadata types
│   │   ├── capability.go      # Model capabilities (completion, embedding, vision, etc.)
│   │   ├── config.go          # Model configuration (OS, arch, rootfs)
│   │   ├── families.go        # Known model family enumerations
│   │   └── name.go            # Model name parsing and validation
│   ├── errtypes/              # Standardized error types
│   └── syncmap/               # Generic concurrent map
│
├── fs/                        # File system utilities
│   ├── config.go              # Config path helpers
│   ├── ggml/                  # GGML binary format parser
│   └── gguf/                  # GGUF metadata format parser
│
├── template/                  # Chat template engine
├── parser/                    # Modelfile parser
├── tools/                     # Tool calling support
├── thinking/                  # Thinking/reasoning block parsing
├── manifest/                  # OCI-compatible model manifest system
│   ├── manifest.go            # Manifest parsing and management
│   ├── layer.go               # Content-addressable blob layers
│   └── paths.go               # Storage path resolution
├── progress/                  # Terminal progress bars and spinners
├── format/                    # Human-readable formatting (bytes, duration)
├── envconfig/                 # Environment variable configuration
├── auth/                      # ED25519 key-based authentication
├── kvcache/                   # KV cache implementations
│   ├── causal.go              # Causal attention KV cache
│   ├── encoder.go             # Encoder KV cache
│   ├── recurrent.go           # Recurrent model KV cache + checkpoints
│   └── wrapper.go             # KV cache wrapper/adaptor
├── readline/                  # Terminal line editing
├── harmony/                   # Harmony: LLM prompt format parser
├── runner/                    # Subprocess runner dispatch (platform-specific)
├── logutil/                   # Structured logging (slog)
├── version/                   # Version injection point
├── internal/                  # Shared internal utilities
│   ├── cloud/                 # Cloud connectivity policy
│   ├── modelref/              # Model reference parsing
│   └── orderedmap/            # Order-preserving map
│
├── llama/                     # llama.cpp build integration
│   ├── server/                # llama-server build with CGO
│   └── compat/                # Compatibility layer patches
│
├── integration/               # End-to-end integration tests
├── scripts/                   # Build + install scripts for all platforms
├── app/                       # Desktop app (macOS + Windows tray app)
├── community/                 # Community tools (Discord bot)
├── docs/                      # Documentation site (Mintlify)
├── cmake/                     # CMake build system (llama.cpp, MLX)
├── CMakeLists.txt             # Root CMake project
├── Dockerfile                 # Multi-stage Docker build
├── docker-compose.yml         # Docker Compose setup
└── go.mod / go.sum            # Go module dependencies
```

---

## Request Flow

### 1. HTTP Request Entry

All requests enter through the Go standard library's HTTP server with HTTP/2 Cleartext (h2c) support. The server is configured with `http2.Server{}` wrapped in `h2c.NewHandler` to accept both HTTP/1.1 and HTTP/2.

### 2. Gin Router

`server.GenerateRoutes()` creates a Gin engine with CORS middleware, host allowlisting, and a version header injection. Three route families coexist:

**Native Lychee API** (`/api/*`) — direct endpoints for the Lychee CLI and first-party clients.

**OpenAI-compatible API** (`/v1/*`) — transparently translates OpenAI-format requests/responses. The `middleware.OpenAI*` middleware intercepts both incoming request bodies and outgoing response bodies, performing format conversion bidirectionally.

**Anthropic-compatible API** (`/v1/messages`) — translates Anthropic Messages API to/from Lychee's internal format using `middleware.Anthropic*`.

### 3. Middleware Layer

Route handlers are wrapped in middleware chains. For inference endpoints:
- `cloudPassthroughMiddleware` — optionally forwards to Lychee Cloud if a remote model is requested.
- Format middleware (`ChatMiddleware`, `CompletionsMiddleware`, etc.) — transforms external API formats to Lychee's internal `api.ChatRequest` / `api.GenerateRequest`.
- `withInferenceRequestLogging` — captures per-request metadata for observability.

### 4. Handler → Scheduler → LLM Backend

1. **Handler** receives the validated request and calls `scheduleRunner()`.
2. **Model Resolution** looks up the model in local storage, resolves aliases, checks capabilities.
3. **Scheduler** (`sched.go`) manages GPU VRAM allocation. It decides whether to:
   - Use an already-loaded runner (hot model).
   - Load a new runner, possibly evicting others.
   - Reject if insufficient resources.
4. **LLM Server** (`llm/server.go`) wraps a `llama-server` subprocess (from llama.cpp). Communication happens via HTTP to the subprocess on a dynamically allocated port.
5. **Streaming**: Responses flow from llama-server → handler → middleware → HTTP client as streaming JSON or SSE.

### 5. Chat Template Pipeline

For chat requests:
- `server/prompt.go` constructs the prompt using the model's chat template.
- **Parsers** (`model/parsers/`) understand each model family's chat format (thinking tags, tool call delimiters, etc.).
- **Renderers** (`model/renderers/`) produce the final prompt string with image tags, JSON schemas, and tool definitions.
- **Harmony** (`harmony/`) provides a generic tagged-format parser for models using `<|start|>/<|end|>` delimiters.

### Visual Flow Summary

```
HTTP Request
    │
    ▼
h2c Handler (HTTP/1.1 + HTTP/2)
    │
    ▼
Gin Engine + CORS Middleware
    │
    ├─ /api/* ─────────────────────────────────────────────► Native Handler
    │                                                            │
    ├─ /v1/chat/completions ─► OpenAI Middleware (translate) ────┤
    │                                                            │
    └─ /v1/messages ─────────► Anthropic Middleware (translate) ─┤
                                                                 │
                                                    ┌────────────┘
                                                    ▼
                                              scheduleRunner()
                                                    │
                                          ┌─────────┼─────────┐
                                          │         │         │
                                     Model        GPU       Cloud
                                   Resolution   Scheduler   Proxy
                                          │         │         │
                                          └────┬────┘         │
                                               ▼              │
                                        LlamaServer           │
                                      (llama.cpp subprocess)  │
                                               │              │
                                               ▼              ▼
                                          Streaming Response
                                               │
                                               ▼
                                    Middleware (if compat API)
                                               │
                                               ▼
                                          HTTP Response
```

---

## Multi-Backend Overview

### Local Inference Backends

Lychee uses `llama.cpp` (`llama-server`) as its primary local inference backend, discovered through the `discover/` and `llm/` packages.

**GPU Backend Detection** (`discover/gpu.go`, platform-specific files):
- **CUDA** — NVIDIA GPUs via CUDA libraries
- **ROCm** — AMD GPUs via ROCm/HIP
- **Metal** — Apple Silicon via Metal Performance Shaders
- **Vulkan** — Cross-platform GPU compute

**Backend Selection**: At startup, `discover.GPUDevices()` probes available GPUs and logs their capabilities. The scheduler uses this information (VRAM, compute capability) to decide where to place models.

### Cloud Backend

Lychee integrates with **Lychee Cloud** for remote inference. When a model is marked as cloud-hosted, requests are proxied to the cloud service instead of loading locally. This is managed by `server/cloud_proxy.go` and `internal/cloud/policy.go`.

### HuggingFace Integration

`server/huggingface.go` provides direct model discovery and download from HuggingFace Hub. It:
- Queries the HuggingFace API for repository file listings.
- Detects GGUF files suitable for local inference.
- Handles LFS (Large File Storage) downloads with progress tracking.
- Manages authentication via HuggingFace tokens.

### Model Types and Architecture Support

The `convert/` package supports importing models from multiple formats:
- **PyTorch** (`.bin`, `.pt`)
- **Safetensors** (`.safetensors`)
- **GGUF/GGML** — natively supported

Model architectures with dedicated converters include: Llama, Mistral, Mixtral, Gemma (1–4), Qwen (2, 2.5-VL, 3, 3-Next, 3-VL, 3-Coder), DeepSeek v2, Phi-3, BERT, OLMo, Command-R, Nemotron-H, GPT-OSS, Laguna, LFM-2, GLM-4-MoE, GLM-OCR, EmbeddingGemma, and others.

### Extended Backend (x/)

The `x/` directory contains experimental and platform-specific backends:
- **Image Generation** (`x/imagegen/`) — Flux and ZImage models via MLX (Apple Silicon only).
- **MLX Runner** (`x/mlxrunner/`) — Apple MLX tensor framework for inference.
- **Agent Sandboxing** (`x/agent/`) — Secure code execution with approval workflows.

---

## Key Packages and Responsibilities

| Package | Responsibility |
|---|---|
| `cmd/` | CLI entry points. All `lychee` subcommands. TUI integration. Launcher config generation. |
| `server/` | HTTP API server. Route registration. Model management. Scheduling. HuggingFace integration. |
| `api/` | Go client library for the Lychee REST API. Request/response type definitions. |
| `llm/` | LLM server lifecycle. Spawns and manages llama-server subprocesses. Platform-specific GPU config. |
| `discover/` | Hardware discovery. GPU enumeration, system memory, backend capability probing. |
| `middleware/` | API compatibility translation layer. OpenAI and Anthropic format converters. |
| `model/` | Model architecture definitions. Chat template parsers and renderers per model family. |
| `convert/` | Model format conversion (PyTorch/safetensors → GGUF). Tokenizer conversion. |
| `ml/` | Machine learning primitives: backends, devices, NN layer types. |
| `manifest/` | OCI-compatible content-addressable storage for model files and layers. |
| `types/` | Shared type definitions for model metadata: capabilities, configuration, names, errors. |
| `fs/` | Low-level file format parsers for GGML and GGUF binary formats. |
| `template/` | Chat/prompt template engine. Jinja2-style template processing. |
| `parser/` | Modelfile parser. Reads Lychee model definition files. |
| `tools/` | Tool/function calling support. Tool registry and execution. |
| `thinking/` | Thinking/reasoning content extraction from model outputs. |
| `anthropic/` | Anthropic API data types and streaming SSE converter. |
| `openai/` | OpenAI API data types and streaming format converter. |
| `auth/` | ED25519 key-based authentication for cloud services. |
| `envconfig/` | Environment variable configuration with defaults and validation. |
| `format/` | Human-readable formatting utilities (bytes, durations). |
| `progress/` | Terminal progress bars and spinners for downloads and operations. |
| `readline/` | Terminal line editing for interactive mode. |
| `harmony/` | Generic tagged-format prompt parser (`<|start|>/<|end|>` delimiters). |
| `kvcache/` | KV cache memory management. Causal, encoder, and recurrent variants with checkpointing. |
| `logutil/` | Structured logging configuration using Go's `log/slog`. |
| `x/` | Extended features: image generation, MLX runner, agent sandboxing, safetensors. |
| `runner/` | Platform-specific subprocess runner dispatch for imagegen and MLX engines. |
| `internal/` | Shared internal utilities: cloud policy, model ref parsing, ordered maps, sync primitives. |

---

## Build System

### Go Module

The project is a single Go module (`github.com/MD-Mushfiqur123/lychee`) requiring **Go 1.25**. Building the pure-Go components is straightforward:

```bash
go build ./...
```

### CGO and llama.cpp

The real build complexity comes from CGO integration with `llama.cpp`. The `llama/` directory contains:
- `llama/server/CMakeLists.txt` — builds `llama-server` binary via `FetchContent` from the pinned llama.cpp source.
- `llama/compat/` — compatibility patches and C++ shims.

**Root `CMakeLists.txt`** orchestrates the build, configuring GGML backends (CUDA, ROCm, Metal, Vulkan) and the MLX framework on macOS.

### Platform Build Scripts

The `scripts/` directory has platform-specific build scripts:
- `build_darwin.sh` — macOS (Metal + MLX)
- `build_linux.sh` — Linux (CUDA/ROCm/Vulkan)
- `build_windows.ps1` — Windows (CUDA/Vulkan)
- `build_docker.sh` / `docker-build.ps1` — Docker images
- `install.sh` / `install.ps1` / `install.bat` — Installers

### Docker

Multi-stage Docker build:
1. **Builder stage** — `golang:1.25-alpine` with CGO toolchain (gcc, g++, cmake, make).
2. **Runtime stage** — Minimal `alpine:latest` with compiled binary.

### CI/CD

GitHub Actions (`ci.yml`, `release.yml`, `docker-publish.yml`):
- **CI** — `go mod tidy`, `go build ./...`, `go vet ./...` on every push/PR.
- **Release** — Cross-platform builds with CGO for macOS, Linux, Windows.
- **Docker Publish** — Builds and pushes multi-arch Docker images.

### Desktop App

The `app/` directory contains a native desktop application (macOS menu bar app and Windows system tray) built with:
- Go for the backend server management.
- C++ WebView for the UI shell.
- TypeScript/React frontend (`app/ui/app/`) using TanStack Router, Tailwind CSS, and Vite.
- InnoSetup for Windows installer packaging.

---

## Storage Layout

Models and metadata are stored in `~/.lychee/`:

```
~/.lychee/
├── models/              # Model manifests (OCI-compatible JSON)
├── blobs/               # Content-addressable blob storage (SHA256)
├── id_ed25519           # Private key for cloud auth
├── id_ed25519.pub       # Public key
├── conversations/       # Chat history (JSON files)
├── routes.json          # Model routing configuration
├── aliases.json         # Model aliases
└── config.json          # Cloud configuration
```

---

## Design Patterns

**Content-Addressable Storage**: Models are stored as layers with SHA256 digests, inspired by OCI container images. This enables deduplication, layer sharing between models, and integrity verification.

**Template-Driven Architecture**: Each model family has dedicated parser and renderer implementations. The registry in `model/parsers/parsers.go` maps model families to their parsing/render logic, making it easy to add support for new architectures.

**Backend Abstraction**: The `ml/backend.go` interface and `llm/server.go` `LlamaServer` interface abstract away hardware differences. Platform-specific code lives in `_darwin.go`, `_linux.go`, `_windows.go` files with build tags.

**Compatibility as Middleware**: OpenAI and Anthropic API compatibility is achieved through Gin middleware that sits between the HTTP layer and Lychee's native handlers, converting request/response formats without modifying core business logic.

**Scheduler-Based Resource Management**: Models are loaded on-demand and evicted under memory pressure. The scheduler in `server/sched.go` tracks VRAM across all GPUs, supports prediction-based eviction, and handles OOM recovery via retry with full eviction.
