# 🗺️ Lychee Roadmap

> Lychee: Universal AI model puller & manager. Half app, half engine — full power.

---

## 🟢 Now (v0.2.x) — Stable

| Status | Feature | Description |
|--------|---------|-------------|
| ✅ | Universal pull (auto HF detect) | Pull models from Hugging Face with automatic model-type detection |
| ✅ | Interactive REPL | Built-in interactive shell for exploring and managing models |
| ✅ | Built-in dashboard | Web-based dashboard for visual model management |
| ✅ | Shell completions | Tab completions for bash, zsh, fish, and PowerShell |
| ✅ | Self-update | `lychee update` keeps the CLI current |
| ✅ | Benchmark | `lychee bench` for model inference performance testing |
| ✅ | Desktop app | Cross-platform desktop GUI (Electron/Tauri) |
| ✅ | Python/JS/Go/Ruby SDKs | Native SDKs for programmatic model management |
| ✅ | Docker multi-arch | Official images for linux/amd64 and linux/arm64 |
| ✅ | Config file support | `lychee.toml` / `~/.lychee/config.toml` for persistent settings |

---

## 🟡 Next (v0.3.x) — Under Development

| Status | Feature | Description |
|--------|---------|-------------|
| 🚧 | CGO binaries (llama.cpp backend) | Native compiled binaries using llama.cpp for zero-dependency inference |
| 🚧 | GPU acceleration (CUDA, ROCm, Metal) | Hardware-accelerated inference across NVIDIA, AMD, and Apple Silicon |
| 🚧 | Model quantization | Built-in GGUF, AWQ, GPTQ quantization workflows |
| 🚧 | RAG integration | First-class retrieval-augmented generation pipeline |
| 🚧 | Web search plugin | Plug-and-play web search augmentation for models |
| 📋 | LangChain integration | Native LangChain model wrappers and tool support |
| 📋 | Multi-modal (vision models) | Pull, run, and benchmark vision-language models |
| 📋 | Voice interface | Voice-to-text and text-to-voice interaction layer |

---

## 🔵 Future (v1.0) — Planned

| Status | Feature | Description |
|--------|---------|-------------|
| 📋 | Lychee Cloud (serverless API) | Hosted inference API with pay-per-token pricing |
| 📋 | Model marketplace | Discover, share, and rate community-curated models |
| 📋 | Team collaboration | Shared model registries, access controls, and audit logs |
| 📋 | Enterprise SSO | SAML/OIDC single sign-on for organizations |
| 📋 | Mobile app | iOS and Android companion for remote model management |

---

**Legend:** ✅ Done &nbsp;|&nbsp; 🚧 In Progress &nbsp;|&nbsp; 📋 Planned
