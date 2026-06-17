# 🍒 Lychee

<p align="center">
  <img src="https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/docs/assets/lychee-banner.svg" alt="Lychee Logo" width="600"/>
</p>

<p align="center">
  <strong>The Universal Local LLM Runtime & Orchestration Layer</strong><br/>
  <em>Pull. Run. Orchestrate. — Any model, any API, anywhere.</em>
</p>

<p align="center">
  <a href="https://github.com/MD-Mushfiqur123/lychee/stargazers">
    <img src="https://img.shields.io/github/stars/MD-Mushfiqur123/lychee?style=for-the-badge&color=yellow" alt="GitHub Stars"/>
  </a>
  <a href="https://github.com/MD-Mushfiqur123/lychee/releases">
    <img src="https://img.shields.io/github/v/release/MD-Mushfiqur123/lychee?style=for-the-badge&color=green&label=release" alt="Latest Release"/>
  </a>
  <a href="https://github.com/MD-Mushfiqur123/lychee/blob/main/go.mod">
    <img src="https://img.shields.io/github/go-mod/go-version/MD-Mushfiqur123/lychee?style=for-the-badge&color=00ADD8" alt="Go Version"/>
  </a>
  <a href="https://github.com/MD-Mushfiqur123/lychee/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/license-MIT-blue?style=for-the-badge" alt="MIT License"/>
  </a>
  <a href="https://goreportcard.com/report/github.com/MD-Mushfiqur123/lychee">
    <img src="https://goreportcard.com/badge/github.com/MD-Mushfiqur123/lychee?style=for-the-badge" alt="Go Report Card"/>
  </a>
  <a href="https://md-mushfiqur123.github.io/lychee-docs/">
    <img src="https://img.shields.io/badge/docs-vitepress-646cff?style=for-the-badge" alt="Documentation"/>
  </a>
  <a href="https://pkg.go.dev/github.com/MD-Mushfiqur123/lychee">
    <img src="https://pkg.go.dev/badge/github.com/MD-Mushfiqur123/lychee.svg" alt="Go Reference"/>
  </a>
  <a href="https://github.com/MD-Mushfiqur123/lychee/actions/workflows/docker-publish.yml">
    <img src="https://github.com/MD-Mushfiqur123/lychee/actions/workflows/docker-publish.yml/badge.svg" alt="Docker Build"/>
  </a>
  <a href="https://github.com/MD-Mushfiqur123/lychee/actions/workflows/ci.yml">
    <img src="https://github.com/MD-Mushfiqur123/lychee/actions/workflows/ci.yml/badge.svg" alt="CI"/>
  </a>
</p>

<p align="center">
  <a href="https://github.com/MD-Mushfiqur123/lychee-desktop">
    <img src="https://img.shields.io/badge/DESKTOP-AVAILABLE-Crimson?style=for-the-badge&logo=windows&logoColor=white" alt="Lychee Desktop Available"/>
  </a>
  &nbsp;
  <a href="https://github.com/MD-Mushfiqur123/lychee-desktop/releases">
    <img src="https://img.shields.io/github/v/release/MD-Mushfiqur123/lychee-desktop?style=for-the-badge&color=Crimson&label=desktop%20release" alt="Desktop Release"/>
  </a>
  &nbsp;
  <a href="https://github.com/MD-Mushfiqur123/lychee/releases">
    <img src="https://img.shields.io/badge/download-binary-28a745?style=for-the-badge&logo=github&logoColor=white" alt="Download Binary"/>
  </a>
</p>

<p align="center">
  <a href="https://github.com/MD-Mushfiqur123/lychee/pkgs/container/lychee">
    <img src="https://img.shields.io/badge/docker-ghcr.io-blue?style=for-the-badge&logo=docker&logoColor=white" alt="Docker GHCR"/>
  </a>
  &nbsp;
  <a href="https://github.com/MD-Mushfiqur123/lychee/blob/main/Dockerfile">
    <img src="https://img.shields.io/badge/dockerfile-multi--arch-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Dockerfile"/>
  </a>
  &nbsp;
  <a href="https://github.com/MD-Mushfiqur123/lychee/actions/workflows/docker-publish.yml">
    <img src="https://img.shields.io/github/actions/workflow/status/MD-Mushfiqur123/lychee/docker-publish.yml?style=for-the-badge&label=docker%20build" alt="Docker Build Status"/>
  </a>
</p>

<p align="center">
  <a href="https://github.com/MD-Mushfiqur123/lychee">
    <img src="https://img.shields.io/badge/Try%20it-go%20install%20%40latest-28a745?style=for-the-badge&logo=go&logoColor=white" alt="Try it: go install @latest"/>
  </a>
</p>

## 🍒 Try It Now

```bash
# 1. Install
go install github.com/MD-Mushfiqur123/lychee@latest

# 2. Start server
lychee serve

# 3. Chat with a model
lychee run qwen2.5:3b "Hello, Lychee!"
```

---

## 📥 Install

<p align="center">
  <strong>Copy & paste into your terminal — one command, zero friction.</strong>
</p>

<table align="center">
  <tr>
    <th align="center">🪟 Windows</th>
    <th align="center">🍎 macOS</th>
    <th align="center">🐧 Linux</th>
    <th align="center">🔧 Go</th>
  </tr>
  <tr>
    <td align="center">
      <sub>PowerShell</sub><br/>
      <code>irm https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/install.ps1 | iex</code>
    </td>
    <td align="center">
      <sub>curl | sh</sub><br/>
      <code>curl -fsSL https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/install.sh | sh</code>
    </td>
    <td align="center">
      <sub>curl | sh</sub><br/>
      <code>curl -fsSL https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/install.sh | sh</code>
    </td>
    <td align="center">
      <sub>go install (latest)</sub><br/>
      <code>go install github.com/MD-Mushfiqur123/lychee@latest</code>
    </td>
  </tr>
</table>

```bash
# 🍒 Install Lychee (requires Go 1.22+)
go install github.com/MD-Mushfiqur123/lychee@latest

# Verify it works
lychee version
```

> 💡 The one-liner scripts auto-detect your OS and architecture. They try `go install` first (if Go 1.22+ is available), otherwise they download the pre-built binary from GitHub Releases. **Docker images** are also available — see [Installation](#-installation) section below.

---

## 🚀 Quick Start

### One-Liner Install

| Platform | Command |
|:---|:---|
| 🐧 **Linux** / 🍎 **macOS** | `curl -fsSL https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/install.sh \| sh` |
| 🪟 **Windows** | `irm https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/install.ps1 \| iex` |
| 🔧 **Go users** | `go install github.com/MD-Mushfiqur123/lychee@latest` |

```bash
# After install, start the server
lychee serve

# Pull and chat with any HuggingFace model — instantly
lychee pull microsoft/Phi-3-mini-4k-instruct-gguf --run
```

> 🌐 Open **http://localhost:11434** in your browser for the built-in Dashboard UI.

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🌍 Universal Model Pull
Pull **any HuggingFace GGUF model** natively. No registry lock-in.  
`lychee pull org/model` — auto-detects HF repos.

```bash
lychee pull bartowski/Meta-Llama-3.1-8B-Instruct-GGUF
```

### ⚡ Auto-Run Mode
Pull + chat in one command. Models start the moment they're ready.

```bash
lychee pull microsoft/Phi-3-mini-4k --run
```

### 🖥️ Built-in Dashboard
Browser-based UI at `localhost:11434` — manage models, monitor performance, chat interactively.

</td>
<td width="50%">

### 🎛️ 24 CLI Commands
Full toolkit for model management, benchmarking, pipelines, and more.

```bash
lychee help    # See all 24 commands
```

### 🔌 Dual API Support
OpenAI-compatible Chat Completions **and** Anthropic Messages API — point any SDK at Lychee.

```python
# OpenAI SDK → Lychee
client = OpenAI(base_url="http://localhost:11434/v1")

# Anthropic SDK → Lychee
client = Anthropic(base_url="http://localhost:11434/v1")
```

### 📦 50+ HuggingFace Models
Access every GGUF model on HuggingFace. Concurrent multi-shard downloads with SHA256 verification.

</td>
</tr>
<tr>
<td width="50%">

### 🔗 Multi-Model Pipelines (DAG)
Chain models sequentially. Output flows through composable steps.

```bash
curl http://localhost:11434/api/compose \
  -d '{"input":"Hello","steps":[
    {"model":"gemma3","prompt":"Translate: {{input}}"},
    {"model":"phi3","prompt":"Analyze: {{step[0].output}}"}
  ]}'
```

### 💾 Persistent Memory
SQLite-backed conversation store. Save, resume, and search across sessions.

```bash
curl http://localhost:11434/api/conversations \
  -d '{"model":"gemma3","messages":[...]}'
```

</td>
<td width="50%">

### ⚖️ Load Balancing Router
Route requests across multiple instances with weighted round-robin.

```bash
curl http://localhost:11434/api/routes \
  -d '{"name":"fast","endpoints":[...],"strategy":"weighted_round_robin"}'
```

### ✅ Structured Output + Auto-Retry
JSON schema validation with automatic error-correction retries.

```bash
curl http://localhost:11434/api/structured \
  -d '{"model":"gemma3","prompt":"...","schema":{...},"max_retries":3}'
```

</td>
</tr>
</table>

---

## 📊 Lychee vs Ollama

| Feature | 🦙 Ollama | 🍒 Lychee |
|:---|:---:|:---:|
| **Core inference engine** | ✅ llama.cpp | ✅ llama.cpp + MLX |
| **OpenAI-compatible API** | ✅ `/v1/chat/completions` | ✅ `/v1/chat/completions` |
| **Anthropic Messages API** | ❌ | ✅ `/v1/messages` |
| **OpenAI Responses API** | ❌ | ✅ `/v1/responses` (structured) |
| **HuggingFace model pull** | ❌ Registry-only | ✅ `lychee pull org/model` — **50+ models** |
| **Universal pull (auto HF detect)** | ❌ | ✅ Any `org/model` format |
| **Auto-run (pull + chat)** | ❌ | ✅ `lychee pull model --run` |
| **Built-in browser dashboard** | ❌ | ✅ `http://localhost:11434` |
| **Native desktop app** | ❌ | ✅ Lychee Desktop (Windows) |
| **CLI commands** | ~12 | ✅ **24 commands** |
| **Multi-model pipelines** | ❌ | ✅ DAG-based Model Composer |
| **Persistent conversation memory** | ❌ | ✅ SQLite + JSON store |
| **Structured output + auto-retry** | ❌ | ✅ JSON schema validation |
| **Multi-instance load balancing** | ❌ | ✅ Weighted round-robin router |
| **Interactive TUI dashboard** | ❌ | ✅ `lychee stats --tui` |
| **Model benchmarking** | ❌ | ✅ `lychee compare` |
| **Interactive agent sandbox** | ❌ | ✅ `lychee run --experimental` |
| **Hardware scanner & optimizer** | ❌ | ✅ `lychee scan` |
| **Code boilerplate generator** | ❌ | ✅ `lychee generate-client` |
| **Official Python SDK** | ❌ | ✅ `lychee-python` |
| **Official JavaScript SDK** | ❌ | ✅ `lychee-js` |
| **Official Ruby SDK** | ❌ | ✅ `lychee` gem |
| **Official Go SDK** | ❌ | ✅ `sdk/go` |

> **TL;DR**: Use Ollama for barebones local inference. Use Lychee when you need **orchestration, multi-model pipelines, universal HuggingFace pulls, a browser dashboard, structured output, persistent memory, load balancing, and a full 24-command developer toolkit.**

---

## 🖥️ Lychee Desktop

<p align="center">
  <a href="https://github.com/MD-Mushfiqur123/lychee-desktop/releases">
    <img src="https://img.shields.io/badge/⬇️%20Download-Lychee%20Desktop%20v0.1.0%20Alpha-Crimson?style=for-the-badge&logo=windows&logoColor=white" alt="Download Lychee Desktop"/>
  </a>
</p>

Lychee Desktop is the native GUI companion app for Windows. It wraps the full Lychee engine with a polished chat interface, pipeline builder, and model manager — no terminal required.

| 🎯 **Chat Interface** | 🔗 **Pipeline Builder** | 📦 **Model Manager** |
|:---:|:---:|:---:|
| Beautiful multi-turn chat | Visual DAG pipeline composer | Pull, list, and manage models |
| Supports all 24 CLI features | Drag-and-drop step chaining | Real-time download progress |
| Conversation memory browser | Save & share pipelines | 50+ HuggingFace models |

### Quick Install

```powershell
# Download the latest release
Invoke-WebRequest -Uri "https://github.com/MD-Mushfiqur123/lychee-desktop/releases/latest/download/Lychee-Setup.exe" -OutFile "Lychee-Setup.exe"

# Install
Start-Process -FilePath "Lychee-Setup.exe" -Wait
```

> 📖 See the [Lychee Desktop repo](https://github.com/MD-Mushfiqur123/lychee-desktop) for full documentation and release notes.

---

## 📦 Installation

### 🍒 One-Liner (Recommended — All Platforms)

```bash
# Linux & macOS
curl -fsSL https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/install.sh | sh

# Windows (PowerShell)
irm https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/install.ps1 | iex

# Go users (requires Go 1.22+)
go install github.com/MD-Mushfiqur123/lychee@latest
```

> 💡 The one-liner automatically detects your OS and architecture. It tries `go install` first (if Go 1.22+ is available), then falls back to downloading the pre-built binary from GitHub Releases.

### Uninstall

```bash
# Linux & macOS
curl -fsSL https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/uninstall.sh | sh

# Windows
# Delete %LOCALAPPDATA%\Programs\lychee\lychee.exe and remove from PATH manually
```

### Pre-built Binaries

Download the latest binary for your platform from the [Releases page](https://github.com/MD-Mushfiqur123/lychee/releases):

| Platform | Format |
|:---|:---|
| 🪟 **Windows** | `.exe` (x64) |
| 🍎 **macOS** | `.tar.gz` (Intel + Apple Silicon) |
| 🐧 **Linux** | `.tar.gz` (x64 + ARM64) |

### Docker

```bash
# Build from GitHub
docker build -t lychee https://github.com/MD-Mushfiqur123/lychee.git

# Run with persistent model volume
docker run -d -v lychee-models:/root/.lychee/models -p 11434:11434 --name lychee lychee

# Quick one-liner test (build, start, verify, cleanup)
./scripts/docker-test.sh
```

> 💡 Use `docker compose up -d` with the included `docker-compose.yml` for health checks and automatic restarts.

### Build from Source

```bash
git clone https://github.com/MD-Mushfiqur123/lychee.git
cd lychee
go build -o lychee .
sudo mv lychee /usr/local/bin/
```

### Client SDKs

Official SDKs for programmatic access to all Lychee features.

```bash
# Python
pip install lychee-python

# JavaScript / TypeScript
npm install lychee-js

# Go
go get github.com/MD-Mushfiqur123/lychee/sdk/go

# Ruby
# gem install lychee
# (or add to Gemfile: gem 'lychee')
```

#### 🐍 Python Quick Example

```python
from lychee import LycheeClient

client = LycheeClient()  # http://localhost:11434

# Chat
resp = client.chat("gemma3", [{"role": "user", "content": "Explain AI in one sentence."}])
print(resp["message"]["content"])

# Generate with streaming
for chunk in client.generate("gemma3", "Write a haiku.", stream=True):
    print(chunk.get("response", ""), end="", flush=True)

# List all models
for m in client.list_models():
    print(m["name"], m["size"])

# Pull from HuggingFace
for progress in client.pull_model("bartowski/Meta-Llama-3.1-8B-Instruct-GGUF"):
    print(f"\r{progress.get('status')}: {progress.get('completed')}/{progress.get('total')}", end="")

# Get embeddings
vec = client.embeddings("nomic-embed-text", "Hello world")
print(len(vec))  # e.g. 768

# Multi-model DAG pipeline
result = client.compose(
    input="Hello world",
    steps=[
        {"model": "gemma3", "prompt": "Translate to French: {{input}}"},
        {"model": "llama3.2", "prompt": "Summarize: {{step[0].output}}"},
    ]
)
print(result["output"])
```

> 📦 The Python SDK is available as both the `lychee-python` package (zero-dependency) and a
> single-file `lychee.py` module (requires `requests`). See
> [sdk/python/README.md](sdk/python/README.md) for details.

#### 💎 Ruby Quick Example

```ruby
require 'lychee'

client = LycheeClient.new  # http://localhost:11434

# Chat
resp = client.chat("gemma3", [{role: "user", content: "Explain AI in one sentence."}])
puts resp["message"]["content"]

# Generate with streaming
client.generate("gemma3", "Write a haiku.", stream: true) do |chunk|
  print chunk["response"]
end

# List all models
client.list_models["models"].each { |m| puts "#{m['name']} #{m['size']}" }

# Pull from HuggingFace
client.pull_model("bartowski/Meta-Llama-3.1-8B-Instruct-GGUF") do |p|
  print "\r#{p['status']}: #{p['completed']}/#{p['total']}"
end

# Get embeddings
vec = client.embeddings("nomic-embed-text", "Hello world")
puts vec.length  # e.g. 768
```

> 💎 The Ruby SDK is a zero-dependency gem (stdlib only). See
> [sdk/ruby/README.md](sdk/ruby/README.md) for details.

---

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                      Lychee Server (:11434)                   │
│  ┌─────────────┐ ┌──────────────┐ ┌────────────────────────┐ │
│  │  OpenAI API  │ │ Anthropic API│ │   Lychee Native API    │ │
│  │/v1/chat/comp │ │ /v1/messages │ │ /api/chat, /api/compose│ │
│  └─────────────┘ └──────────────┘ └────────────────────────┘ │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │              Orchestration Layer (Lychee)                 │ │
│  │  Composer │ Router │ Memory │ Structured Output │ Pull   │ │
│  └──────────────────────────────────────────────────────────┘ │
│  ┌──────────────────────────────────────────────────────────┐ │
│  │       Inference Engine (llama.cpp / MLX)                  │ │
│  │      CUDA · ROCm · Metal · CPU · Vulkan                  │ │
│  └──────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

---

## 🔗 Ecosystem

| Resource | Description | Link |
|:---|:---|:---|
| 🖥️ **Lychee Desktop** | Native Windows GUI app | [md-mushfiqur123/lychee-desktop](https://github.com/MD-Mushfiqur123/lychee-desktop) |
| 🌐 **Landing Page** | Project homepage & showcase | [lychee-landing-page](https://md-mushfiqur123.github.io/lychee-landing-page/) |
| 📖 **Documentation** | Full VitePress docs | [lychee-docs](https://md-mushfiqur123.github.io/lychee-docs/) |
| 🐍 **Python SDK** | Official Python client | `pip install lychee-python` |
| 📦 **npm Package** | Official JS/TS client | `npm install lychee-js` |
| 💎 **Ruby Gem** | Official Ruby client | `gem 'lychee'` |
| 🐛 **Issue Tracker** | Bug reports & feature requests | [GitHub Issues](https://github.com/MD-Mushfiqur123/lychee/issues) |
| 💬 **Discussions** | Community & Q&A | [GitHub Discussions](https://github.com/MD-Mushfiqur123/lychee/discussions) |

---

## 📚 Full CLI Command Reference

Lychee ships with **24 CLI commands** — everything you need to manage local LLMs:

```bash
$ lychee help

COMMANDS:
  serve             Start the Lychee server
  run               Run a model interactively
  pull              Pull a model from HuggingFace (auto-detect)
  list              List all downloaded models
  show              Show model details
  remove            Remove a downloaded model
  compare           Benchmark models side-by-side
  stats             Show performance stats (--tui for dashboard)
  scan              Scan hardware & get model recommendations
  generate-client   Generate API client boilerplate
  hf search         Search HuggingFace for GGUF models
  hf pull           Pull specific HuggingFace model quantization
  compose           Execute multi-model DAG pipelines
  routes            Manage load-balanced model routes
  conversations     Manage persistent conversation memory
  memory            View and search conversation history
  ps                List running models
  stop              Stop a running model
  config            Show/edit configuration
  version           Show version info
  help              Show this help message
```

---

## 🤝 Contributing

We welcome contributions of all kinds — bug fixes, features, docs, and ideas.

See **[CONTRIBUTING.md](./CONTRIBUTING.md)** for:
- Development setup & build instructions
- Commit message conventions
- Pull request guidelines
- Testing requirements
- Our roadmap & priority areas

**Not sure where to start?** Check out [good first issues](https://github.com/MD-Mushfiqur123/lychee/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

### Key Areas We Need Help With

| Area | Skills Needed |
|:---|:---|
| 🧠 **LLM Inference** | llama.cpp, MLX, CUDA, ROCm optimization |
| 🎛️ **Multi-Model Pipelines** | DAG orchestration, model composition |
| 📊 **Structured Output** | JSON schema, constrained generation |
| 💾 **Conversation Memory** | SQLite, vector stores, RAG |
| ⚖️ **Load Balancing** | Multi-instance routing, health checks |
| 🖥️ **TUI / Dashboard** | Go bubbletea, web UIs |
| 🌐 **SDKs & APIs** | Python, TypeScript, API design |
| 📚 **Documentation** | Technical writing, tutorials |

---

## 👥 Community & Contributors

Thanks to everyone who has contributed to Lychee! Want to see your name here? [Open a PR](https://github.com/MD-Mushfiqur123/lychee/pulls)!

<!-- ALL-CONTRIBUTORS-LIST:START -->
<!-- prettier-ignore-start -->
<!-- markdownlint-disable -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

### 📋 Contributors Wanted!

Lychee is actively seeking contributors from the AI/LLM open-source community. See the full list of potential collaborators in **[CONTRIBUTORS.md](./CONTRIBUTORS.md)**.

---

## 📊 Stats

[![Star History](https://api.star-history.com/svg?repos=MD-Mushfiqur123/lychee&type=Date)](https://star-history.com/#MD-Mushfiqur123/lychee&Date)

[![Contributors](https://img.shields.io/github/contributors/MD-Mushfiqur123/lychee)](https://github.com/MD-Mushfiqur123/lychee/graphs/contributors)

---

## 📄 License

Lychee is open-source software licensed under the [MIT License](LICENSE).

---

## 🍒 Powered by Lychee

Show the world your project runs on Lychee — add this badge to your README:

```markdown
[![Powered by Lychee](https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/assets/powered-by-lychee.svg)](https://github.com/MD-Mushfiqur123/lychee)
```

[![Powered by Lychee](https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/assets/powered-by-lychee.svg)](https://github.com/MD-Mushfiqur123/lychee)

---

## 🏗️ Built by Xolvyn

Lychee is engineered and maintained by **[Xolvyn AI Agency](https://md-mushfiqur123.github.io/xolvyn-agency)** — _Intelligence, engineered._

[![Star History Chart](https://api.star-history.com/chart?repos=MD-Mushfiqur123/lychee&type=timeline&logscale&legend=bottom-right)](https://www.star-history.com/?repos=MD-Mushfiqur123%2Flychee&type=timeline&logscale=&legend=bottom-right)

---

## 🌐 Community

<p align="center">
  <a href="https://discord.gg/lychee">
    <img src="https://img.shields.io/badge/Discord-Join%20the%20community-5865F2?style=for-the-badge&logo=discord&logoColor=white" alt="Discord"/>
  </a>
  &nbsp;
  <a href="https://github.com/MD-Mushfiqur123/lychee/discussions">
    <img src="https://img.shields.io/badge/GitHub-Discussions-181717?style=for-the-badge&logo=github&logoColor=white" alt="GitHub Discussions"/>
  </a>
  &nbsp;
  <a href="./CONTRIBUTING.md">
    <img src="https://img.shields.io/badge/Contributing-Guidelines-28a745?style=for-the-badge" alt="Contributing"/>
  </a>
  &nbsp;
  <a href="./CODE_OF_CONDUCT.md">
    <img src="https://img.shields.io/badge/Code%20of%20Conduct-Read%20here-8B5CF6?style=for-the-badge" alt="Code of Conduct"/>
  </a>
</p>
