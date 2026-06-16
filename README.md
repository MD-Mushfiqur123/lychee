# 🍒 Lychee

### The Orchestration Layer for Local LLMs — Built on Ollama

Lychee extends [Ollama](https://github.com/ollama/ollama) with features the upstream doesn't have:
multi-model pipelines, structured output with auto-retry, persistent conversation memory,
and multi-instance load balancing. Everything Ollama does, Lychee does too — plus more.

<p align="center">
  <a href="https://github.com/MD-Mushfiqur123/lychee/stargazers">
    <img src="https://img.shields.io/github/stars/MD-Mushfiqur123/lychee?style=flat-square&color=yellow" alt="GitHub Stars"/>
  </a>
  <a href="https://github.com/MD-Mushfiqur123/lychee/releases">
    <img src="https://img.shields.io/github/v/release/MD-Mushfiqur123/lychee?style=flat-square&color=green&label=release" alt="Latest Release"/>
  </a>
  <a href="https://github.com/MD-Mushfiqur123/lychee/blob/main/go.mod">
    <img src="https://img.shields.io/github/go-mod/go-version/MD-Mushfiqur123/lychee?style=flat-square&color=00ADD8" alt="Go Version"/>
  </a>
  <a href="https://github.com/MD-Mushfiqur123/lychee/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/license-MIT-blue?style=flat-square" alt="MIT License"/>
  </a>
  <a href="https://md-mushfiqur123.github.io/lychee-docs/">
    <img src="https://img.shields.io/badge/docs-vitepress-646cff?style=flat-square" alt="Documentation"/>
  </a>
</p>

---

## ⚡ What Lychee Adds (Original Features)

Lychee extends Ollama with advanced orchestration and developer utilities:

1. **Model Composer**
   Chain multiple local models together sequentially into multi-step pipelines. Pass outputs from one step as inputs to the next, using DAG-based conditional logic and routing.
2. **Structured Output with Auto-Retry**
   Go beyond basic JSON modes. Lychee validates output against a schema and automatically retries with error-correction prompting if the model produces invalid JSON.
3. **Conversation Memory Store**
   Persist chats locally (using SQLite and JSON). Save, list, resume, and delete conversations across sessions with simple API calls.
4. **Model Router with Load Balancing**
   Define virtual model names that route requests across multiple local or remote Ollama/Lychee instances using round-robin, random, or least-loaded strategies.
5. **Official Python & JS SDKs**
   Modern SDKs featuring first-class support for Model Composer, Structured Output, Memory, and Router configurations.

## 🤝 Inherited from Ollama (Full Credit)

The core engine and main capabilities are built directly on top of the excellent work by the Ollama team:
- **Core Inference Engine**: Powered by `llama.cpp` and Apple's MLX for high-performance GPU/CPU inference.
- **Universal API Layer**: OpenAI and Anthropic endpoint compatibility.
- **Stateful Agent Mode**: Embedded agent runtime with local tool execution and approval gates.
- **Model Registry & Management**: Seamlessly pull, run, and customize models.
- **Prompt Caching & KV Sharing**: Automated prefix hashing and KV cache management.

---

## 🔌 API Compatibility Layer

Lychee acts as a local proxy that translates industry-standard APIs into optimized backend executions.

| Endpoint | Standard | Features Supported | Status |
| --- | --- | --- | --- |
| `/api/chat` | Lychee Native | Streaming, Tools, Multi-turn Chat | ✅ Production |
| `/api/generate` | Lychee Native | Streaming, Format, Options | ✅ Production |
| `/v1/chat/completions` | OpenAI | Chat completions, System prompts | ✅ Production |
| `/v1/messages` | Anthropic | Messages API, Streaming, System prompt | ✅ Production |
| `/v1/responses` | OpenAI Responses | Structured responses, validation | ✅ Production |

---

## 📦 Installation

> [!NOTE]
> Lychee is currently in active early alpha phase. Requires Go 1.22+.

### go install (Recommended)
```bash
go install github.com/MD-Mushfiqur123/lychee@v0.1.1-alpha
```

### Docker
```bash
docker build -t lychee https://github.com/MD-Mushfiqur123/lychee.git
docker run -d -v lychee:/root/.lychee -p 11434:11434 --name lychee lychee
```

### Build from Source
```bash
git clone https://github.com/MD-Mushfiqur123/lychee.git
cd lychee
go build -o lychee .
sudo mv lychee /usr/local/bin/
```

### Windows (PowerShell)
```powershell
irm https://raw.githubusercontent.com/MD-Mushfiqur123/lychee/main/scripts/install.ps1 | iex
```

### Client SDKs Installation

Official client SDKs are available to integrate Lychee into your applications:

* **Python SDK:**
  ```bash
  pip install lychee-python
  ```
* **JavaScript SDK:**
  ```bash
  npm install lychee-js
  ```
* **Rust SDK (coming soon):**
  Add the dependency in your `Cargo.toml`:
  ```toml
  [dependencies]
  lychee-rs = { git = "https://github.com/MD-Mushfiqur123/lychee.git", branch = "main" }
  ```


---

## 💡 Core Features Showcase

### 1. Direct HuggingFace Pull (`lychee hf`)
Search and download any GGUF format model directly from HuggingFace Hub with resume support, concurrent multi-shard downloads, and SHA256 integrity verification:
```bash
# Search catalog models
lychee hf search llama

# Download the recommended quantization of a HuggingFace model
lychee hf pull microsoft/Phi-3-mini-4k-instruct-gguf

# List all available quantizations for a model repository
lychee hf pull bartowski/Meta-Llama-3.1-8B-Instruct-GGUF --list

# Download a specific quantization variant
lychee hf pull bartowski/Meta-Llama-3.1-8B-Instruct-GGUF --quant q5_k_m
```

### 2. Drop-in Anthropic SDK Support (`/v1/messages`)
Deploy Claude-compatible workflows locally. Point the official Anthropic client to Lychee:
```python
from anthropic import Anthropic

client = Anthropic(
    base_url="http://localhost:11434/v1",
    api_key="lychee-local"
)

message = client.messages.create(
    model="gemma3",
    max_tokens=1000,
    messages=[
        {"role": "user", "content": "Explain quantum computing in one sentence."}
    ]
)
print(message.content[0].text)
```

### 3. OpenAI Responses API (`/v1/responses`)
Generate structured data conforming to a JSON Schema directly from local models:
```bash
curl http://localhost:11434/v1/responses \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemma3",
    "prompt": "Extract information: Jane Doe is a 28-year-old engineer.",
    "response_format": {
      "type": "json_schema",
      "json_schema": {
        "name": "profile",
        "schema": {
          "type": "object",
          "properties": {
            "name": {"type": "string"},
            "age": {"type": "integer"},
            "occupation": {"type": "string"}
          },
          "required": ["name", "age", "occupation"]
        }
      }
    }
  }'
```

### 4. Interactive Agent Sandbox (`lychee run --experimental`)
Run LLM-driven agents with local tool support. Lychee prompts you with terminal UI approval before executing commands, searching the web, or accessing the filesystem:
```bash
# Start an interactive agent loop with bash, web search, and web fetch capabilities
lychee run gemma3 --experimental

# Inside the prompt, ask:
# >>> Search the web for latest Rust news and compile a list of highlights in rust_news.txt
#
# Lychee will prompt you in real time to approve/reject the web search and bash write actions!
```

### 5. Multi-Model Benchmarking & Comparison (`lychee compare`)
Evaluate multiple local models head-to-head on the same prompt with real-time, side-by-side streaming output and response metrics (tokens/sec, time-to-first-token):
```bash
# Compare gemma3 and phi3 side-by-side
lychee compare "Explain quantum computing in three sentences." gemma3 phi3
```

### 6. Interactive Terminal Performance Dashboard (`lychee stats --tui`)
Monitor active models, VRAM usage, token throughput, and model context usage in real time, directly from your shell, using a beautiful Bubbletea-powered terminal UI:
```bash
lychee stats --tui
```

### 7. Hardware Detection & Code Generation
Optimizing and integrating local models has never been simpler:
* **Hardware Scan (`lychee scan`)**: Scan your hardware configuration and receive personalized local model sizing and execution recommendations.
* **Code Boilerplate Generator (`lychee generate-client`)**: Generate ready-to-run, native API client code in Python, JavaScript, Rust, or Go for any loaded model.

---

## 🚀 Quickstart

1. **Start the Lychee server**
   ```bash
   lychee serve
   ```

2. **Run your first model**
   ```bash
   lychee run gemma3
   ```

3. **Interact via API**
   
   *Using the Anthropic Messages API:*
   ```bash
   curl http://localhost:11434/v1/messages \
     -H "Content-Type: application/json" \
     -d '{
       "model": "gemma3",
       "messages": [{"role": "user", "content": "Hello!"}],
       "max_tokens": 1024
     }'
   ```

   *Structured Output with Auto-Retry (`/api/structured`):*
   ```bash
   curl http://localhost:11434/api/structured \
     -H "Content-Type: application/json" \
     -d '{
       "model": "gemma3",
       "prompt": "Extract the country name and population: France has a population of 67 million.",
       "schema": {
         "type": "object",
         "properties": {
           "country": {"type": "string"},
           "population": {"type": "string"}
         },
         "required": ["country", "population"]
       },
       "max_retries": 3
     }'
   ```

   *Conversation Memory Store (`/api/conversations`):*
   ```bash
   # Create a new persistent conversation session
   curl http://localhost:11434/api/conversations \
     -H "Content-Type: application/json" \
     -d '{
       "model": "gemma3",
       "messages": [
         {"role": "user", "content": "Hello! I am planning a trip to Paris."}
       ]
     }'
   
   # Retrieve conversation details and message history using the returned session ID
   curl http://localhost:11434/api/conversations/<conversation_id>

   # Resume conversation by passing the conversation_id parameter to the chat endpoint
   curl http://localhost:11434/api/chat \
     -H "Content-Type: application/json" \
     -d '{
       "model": "gemma3",
       "conversation_id": "<conversation_id>",
       "messages": [
         {"role": "user", "content": "What is the first thing I should visit?"}
       ]
     }'
   ```

   *Virtual Model Router (`/api/routes`):*
   ```bash
   # Register a virtual model route called 'fast-route' distributed across two backends
   curl http://localhost:11434/api/routes \
     -H "Content-Type: application/json" \
     -d '{
       "name": "fast-route",
       "endpoints": [
         {"host": "http://localhost:11434", "model": "gemma3", "weight": 2},
         {"host": "http://localhost:11435", "model": "phi3", "weight": 1}
       ],
       "strategy": "weighted_round_robin"
     }'

   # Call the virtual route like a standard model
   curl http://localhost:11434/api/chat \
     -H "Content-Type: application/json" \
     -d '{
       "model": "fast-route",
       "messages": [{"role": "user", "content": "Hello!"}]
     }'
   ```

   *Model Composer Chaining (`/api/compose`):*
   ```bash
   # Execute a multi-model pipeline where outputs pass sequentially
   curl http://localhost:11434/api/compose \
     -H "Content-Type: application/json" \
     -d '{
       "input": "Life is like a box of chocolates.",
       "steps": [
         {
           "model": "gemma3",
           "prompt": "Translate this sentence to French: {{input}}"
         },
         {
           "model": "phi3",
           "prompt": "Explain the meaning of this French translation: {{step[0].output}}"
         }
       ],
       "stream": false
     }'
   ```

---

## 🛠️ Supported Backends

Lychee leverages the best-in-class local inference engines under the hood:
* **llama.cpp** — optimized CPU/GPU execution for Apple Silicon (Metal), NVIDIA (CUDA), and AMD (ROCm).
* **MLX** — native Apple Silicon machine learning framework for maximum performance on macOS.

---

## 🔗 Links

- 📖 **Documentation**: [md-mushfiqur123.github.io/lychee-docs](https://md-mushfiqur123.github.io/lychee-docs/)
- 🌐 **Landing Page**: [md-mushfiqur123.github.io/lychee-landing-page](https://md-mushfiqur123.github.io/lychee-landing-page/)
- 💻 **GitHub**: [MD-Mushfiqur123/lychee](https://github.com/MD-Mushfiqur123/lychee)
- 📦 **Releases**: [github.com/MD-Mushfiqur123/lychee/releases](https://github.com/MD-Mushfiqur123/lychee/releases)

---

## 📄 License

Lychee is open-source software licensed under the [MIT License](LICENSE).

[![Star History Chart](https://api.star-history.com/chart?repos=MD-Mushfiqur123/lychee&type=timeline&logscale&legend=bottom-right)](https://www.star-history.com/?repos=MD-Mushfiqur123%2Flychee&type=timeline&logscale=&legend=bottom-right)
