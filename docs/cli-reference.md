# 🍒 Lychee CLI Reference

Complete command-line reference for Lychee v0.x. All commands, flags, examples, and output in one place.

---

## Overview

Lychee ships with **35+ CLI commands** organized into model management, inference, benchmarking, pipeline orchestration, and developer tooling.

```
lychee [command] [flags]
```

| Category | Commands |
|:---|:---|
| **Server** | `serve` |
| **Model Management** | `pull`, `create`, `list`, `show`, `rm`, `cp`, `push`, `ps`, `stop` |
| **Inference** | `run`, `interactive`, `agent`, `embed`, `compose` |
| **HuggingFace** | `hf pull`, `hf search`, `hf list` |
| **Benchmarking** | `bench`, `benchmark`, `compare`, `stats`, `scan`, `inspect` |
| **Configuration** | `config`, `search`, `modelfile`, `quantize` |
| **Distribution** | `export`, `import`, `serve-catalog`, `push`, `signin`, `signout` |
| **Developer** | `completion`, `update`, `generate-client`, `community` |

---

## Global Flags

| Flag | Description |
|:---|:---|
| `-v`, `--version` | Show version information |
| `--verbose` | Show timings for response |
| `--nowordwrap` | Don't wrap words to next line automatically |

### Environment Variables

| Variable | Description |
|:---|:---|
| `LYCHEE_HOST` | Server address (default: `http://localhost:11434`) |
| `LYCHEE_EDITOR` | Editor for interactive sessions |
| `LYCHEE_NOHISTORY` | Disable command history |

---

## Server

### `lychee serve`

Start the Lychee inference server. This is the core daemon that handles all model loading, inference, and API requests.

```
lychee serve [flags]
```

| Flag | Environment | Default | Description |
|:---|:---|:---|:---|
| `--parallel` | `LYCHEE_NUM_PARALLEL` | auto | Number of parallel request slots |
| `--slots` | (alias) | auto | Alias for `--parallel` |

**Environment variables** (also apply):

| Variable | Description |
|:---|:---|
| `LYCHEE_DEBUG` | Enable debug logging |
| `LYCHEE_HOST` | Listen address (`host:port`) |
| `LYCHEE_CONTEXT_LENGTH` | Default context length for loaded models |
| `LYCHEE_KEEP_ALIVE` | Default keep-alive duration (e.g. `5m`) |
| `LYCHEE_MAX_LOADED_MODELS` | Max simultaneously loaded models |
| `LYCHEE_MAX_QUEUE` | Max pending requests in queue |
| `LYCHEE_MODELS` | Path to models directory |
| `LYCHEE_ORIGINS` | Allowed CORS origins |
| `LYCHEE_NOPRUNE` | Disable automatic blob pruning |
| `LYCHEE_FLASH_ATTENTION` | Enable flash attention |
| `LYCHEE_KV_CACHE_TYPE` | KV cache type (f16, f32, q8_0, q4_0) |
| `LYCHEE_LLM_LIBRARY` | Backend library override |
| `LYCHEE_GPU_OVERHEAD` | VRAM overhead per GPU |
| `LYCHEE_LOAD_TIMEOUT` | Load timeout duration |

**Examples:**

```bash
# Start on default port 11434
lychee serve

# Start with 4 parallel slots
lychee serve --parallel 4

# Start on custom host/port via env
LYCHEE_HOST=0.0.0.0:8080 lychee serve
```

**Output:**
```
time=2025-01-15T10:30:00.000Z level=INFO msg="server config" ...
time=2025-01-15T10:30:00.000Z level=INFO msg="Listening on 127.0.0.1:11434 (version 0.x.x)"
```

> 🌐 Open **http://localhost:11434** in your browser for the built-in Dashboard UI.

---

## Model Management

### `lychee pull MODEL`

Pull a model from a registry (Lychee registry or HuggingFace auto-detected).

```
lychee pull MODEL [flags]
```

| Flag | Type | Description |
|:---|:---|:---|
| `--insecure` | bool | Use an insecure registry |
| `--quant` | string | Prefer specific quantization (e.g. `q4_k_m`, `q5_k_m`) — for HuggingFace models |
| `--list` | bool | List all available quantizations then exit — for HuggingFace models |
| `--run` | bool | After pull, automatically start an interactive chat session |

**Examples:**

```bash
# Pull a model from Lychee registry
lychee pull llama3.2

# Pull from HuggingFace (auto-detected)
lychee pull bartowski/Meta-Llama-3.1-8B-Instruct-GGUF

# Pull a specific quantization
lychee pull bartowski/Meta-Llama-3.1-8B-Instruct-GGUF --quant q5_k_m

# List available quantizations for a HuggingFace model
lychee pull bartowski/Mixtral-8x7B-Instruct-v0.1-GGUF --list

# Pull and immediately start chatting
lychee pull microsoft/Phi-3-mini-4k-instruct-gguf --run
```

**Output (pull):**
```
pulling manifest
pulling 6a5e8cf... 100% ▕██████████████████████████████▏ 4.7 GB
verifying sha256 digest
writing manifest
success
```

### `lychee create MODEL`

Create a new model from a Modelfile.

```
lychee create MODEL [flags]
```

| Flag | Type | Default | Description |
|:---|:---|:---|:---|
| `-f`, `--file` | string | `Modelfile` | Path to Modelfile |
| `-q`, `--quantize` | string | | Quantize model to this level (e.g. `q4_K_M`) |
| `--draft-quantize` | string | | Quantize draft model to this level |
| `--experimental` | bool | false | Enable experimental safetensors model creation |

**Examples:**

```bash
# Create from Modelfile in current directory
lychee create my-custom-model

# Create from specific Modelfile
lychee create my-model -f ./Modelfile.custom

# Create with quantization
lychee create my-model -q q4_K_M
```

**Modelfile format:**
```
FROM llama3.2
SYSTEM "You are a helpful assistant."
PARAMETER temperature 0.7
PARAMETER num_ctx 4096
```

### `lychee list`

List all downloaded models. Alias: `ls`.

```
lychee list
```

**Examples:**
```bash
lychee list
lychee ls
```

**Output:**
```
NAME                    ID              SIZE      MODIFIED
llama3.2:latest         123abc456def    2.0 GB    2 days ago
phi-3:mini              789ghi012jkl    2.2 GB    5 hours ago
qwen3:0.6b              a1b2c3d4e5f6    394 MB    1 week ago
```

### `lychee show MODEL`

Show detailed information about a model.

```
lychee show MODEL [flags]
```

| Flag | Type | Description |
|:---|:---|:---|
| `--license` | bool | Show license of a model |
| `--modelfile` | bool | Show Modelfile of a model |
| `--parameters` | bool | Show parameters of a model |
| `--template` | bool | Show template of a model |
| `--system` | bool | Show system message of a model |
| `-v`, `--verbose` | bool | Show detailed model information |

**Examples:**
```bash
lychee show llama3.2
lychee show llama3.2 --modelfile
lychee show llama3.2 --parameters
lychee show llama3.2 --verbose
```

**Output:**
```
  Model:
    architecture        llama
    parameters          3.2B
    context length      131072
    embedding length    3072
    quantization        Q4_K_M
```

### `lychee rm MODEL [MODEL...]`

Remove one or more downloaded models.

```
lychee rm MODEL [MODEL...]
```

**Examples:**
```bash
lychee rm phi-3:mini
lychee rm llama3.2:latest qwen3:0.6b
```

**Output:**
```
deleted 'phi-3:mini'
```

### `lychee cp SOURCE DESTINATION`

Copy a model (creates a duplicate with a new name).

```
lychee cp SOURCE DESTINATION
```

**Examples:**
```bash
lychee cp llama3.2 my-llama-copy
```

**Output:**
```
copied 'llama3.2' to 'my-llama-copy'
```

### `lychee push MODEL`

Push a model to a registry.

```
lychee push MODEL [flags]
```

| Flag | Type | Description |
|:---|:---|:---|
| `--insecure` | bool | Use an insecure registry |

**Examples:**
```bash
lychee push myuser/my-custom-model
```

### `lychee ps`

List currently running (loaded) models with memory usage.

```
lychee ps
```

**Output:**
```
NAME                    ID              SIZE      PROCESSOR    UNTIL
llama3.2:latest         123abc456def    4.5 GB    100% GPU     4 minutes from now
```

### `lychee stop MODEL`

Unload a running model.

```
lychee stop MODEL
```

**Examples:**
```bash
lychee stop llama3.2
```

---

## Inference

### `lychee run MODEL [PROMPT]`

Run a model interactively or with a single prompt.

```
lychee run MODEL [PROMPT] [flags]
```

| Flag | Type | Default | Description |
|:---|:---|:---|:---|
| `--keepalive` | string | | Keep model loaded for duration (e.g. `5m`) |
| `--draft` | string | | Draft model for speculative decoding |
| `--verbose` | bool | false | Show generation timings |
| `--insecure` | bool | false | Use insecure registry |
| `--nowordwrap` | bool | false | Don't wrap output |
| `--format` | string | | Response format (e.g. `json`) |
| `--think` | string | | Thinking mode: `true` / `false` / `high` / `medium` / `low` |
| `--hidethinking` | bool | false | Hide thinking output |
| `--truncate` | bool | true | Truncate inputs exceeding context length |
| `--dimensions` | int | 0 | Truncate output embeddings to dimension (embed models) |
| `--experimental` | bool | false | Enable experimental agent loop with tools |
| `--experimental-yolo` | bool | false | Skip tool approval prompts |
| `--experimental-websearch` | bool | false | Enable web search in experimental mode |
| `--imagegen` | bool | false | Use image generation runner (hidden) |
| `--width`, `--height`, `--steps`, `--seed` | — | | Image generation parameters |

**Examples:**

```bash
# Interactive chat
lychee run llama3.2

# Single prompt
lychee run llama3.2 "Explain quantum computing in one paragraph"

# With thinking mode
lychee run qwen3:8b --think high "Solve this problem step by step: ..."

# With JSON output format
lychee run llama3.2 --format json "List 3 colors"

# Experimental agent mode
lychee run llama3.2 --experimental

# Keep model loaded for 10 minutes after exit
lychee run llama3.2 --keepalive 10m
```

### `lychee interactive [MODEL]`

Start an interactive chat REPL session.

```
lychee interactive [MODEL]
```

Default model: `qwen2.5:3b`

**REPL commands:**

| Command | Description |
|:---|:---|
| `/exit`, `/quit` | Exit the REPL |
| `/help` | Show help |
| `/model <name>` | Switch to a different model |
| `/models` | List available models |
| `/clear` | Clear the screen |
| `/system` | Show system info (GPU, RAM, CPU) |

**Examples:**
```bash
lychee interactive
lychee interactive llama3.2:1b
```

**Output:**
```
🤖 Lychee REPL — model: qwen2.5:3b
Type /help for commands, /exit to quit, /model to switch

> What is machine learning?
  Machine learning is a subset of artificial intelligence that enables systems
  to learn and improve from experience without being explicitly programmed.
```

### `lychee agent MODEL`

Run a model in agentic tool-use mode with interactive tools.

```
lychee agent MODEL [flags]
```

| Flag | Type | Description |
|:---|:---|:---|
| `--keepalive` | string | Keep model loaded for duration |
| `--think` | string | Thinking mode |
| `--hidethinking` | bool | Hide thinking output |
| `--yolo` | bool | Skip all tool approval prompts |
| `--websearch` | bool | Enable web search tool |
| `--verbose` | bool | Show timings |
| `--nowordwrap` | bool | Don't wrap words |

**Examples:**
```bash
lychee agent llama3.2
lychee agent llama3.2 --websearch --yolo
lychee agent qwen3:8b --think high
```

### `lychee embed [text...]`

Generate vector embeddings for text.

```
lychee embed [text...] [flags]
```

| Flag | Type | Default | Description |
|:---|:---|:---|:---|
| `-m`, `--model` | string | `nomic-embed-text` | Embedding model |
| `-o`, `--output` | string | `text` | Output format: `text`, `json`, `csv` |
| `--truncate` | bool | `true` | Truncate inputs exceeding context length |

**Examples:**
```bash
# Single text
lychee embed "hello world" --model nomic-embed-text

# Multiple texts
lychee embed "first" "second" "third" --model nomic-embed-text

# JSON output
lychee embed "hello" --model nomic-embed-text --output json

# Pipe from stdin
echo "some text" | lychee embed --model nomic-embed-text
```

**Output (text):**
```
Dimensions: 768
Vector:     [0.0234, -0.0145, 0.0089, 0.0456, -0.0321, 0.0178, -0.0056, 0.0290] ... (760 more)
Tokens: 2
```

### `lychee compose`

Execute a multi-model DAG pipeline from a JSON steps file.

```
lychee compose [flags]
```

| Flag | Type | Default | Description |
|:---|:---|:---|:---|
| `-s`, `--steps` | string | (required) | Path to JSON steps file |
| `-i`, `--input` | string | | Initial input text |
| `--stream` | bool | `true` | Stream step outputs dynamically |

**Steps JSON format:**

```json
[
  {
    "model": "gemma3:4b",
    "prompt": "Translate the following to French: {{input}}",
    "output": "french"
  },
  {
    "model": "llama3.2:3b",
    "prompt": "Summarize this translation in 1 sentence: {{step[0].output}}"
  }
]
```

**Templating variables:**
- `{{input}}` — pipeline input text
- `{{step[N].output}}` — output from step N (0-indexed)
- `{{step[N].model}}` — model name used in step N

**Examples:**
```bash
lychee compose --steps pipeline.json --input "Hello World"
```

**Output:**
```
[Step 1] Executing model 'gemma3:4b'...
[Step 1] Complete. Output:
Bonjour le monde

[Step 2] Executing model 'llama3.2:3b'...
[Step 2] Complete. Output:
The French translation of "Hello World" is "Bonjour le monde."

Pipeline completed successfully.
Final Output:
The French translation of "Hello World" is "Bonjour le monde."
```

---

## HuggingFace Integration

### `lychee hf pull <org/repo>`

Pull a GGUF model directly from HuggingFace Hub. No account required.

```
lychee hf pull <org/repo> [flags]
```

| Flag | Type | Description |
|:---|:---|:---|
| `--quant` | string | Prefer specific quantization (e.g. `q4_k_m`, `q5_k_m`, `q8_0`, `f16`) |
| `--list` | bool | List all available quantizations then exit |

**Examples:**

```bash
# Pull default (largest) quantization
lychee hf pull microsoft/Phi-3-mini-4k-instruct-gguf

# Pull specific quant
lychee hf pull bartowski/Meta-Llama-3.1-8B-Instruct-GGUF --quant q5_k_m

# List all available quants
lychee hf pull bartowski/Mixtral-8x7B-Instruct-v0.1-GGUF --list
```

**Output (--list):**
```
Available quantizations for bartowski/Meta-Llama-3.1-8B-Instruct-GGUF:
  Q2_K    2.9 GB
  Q3_K_M  3.8 GB
  Q4_K_M  4.9 GB
  Q5_K_M  5.7 GB
  Q6_K    6.6 GB
  Q8_0    8.5 GB
  F16    16.1 GB

Use --quant to select one: lychee hf pull <repo> --quant q4_k_m
```

### `lychee hf search [query]`

Search the built-in Lychee model catalog (50+ models).

```
lychee hf search [query]
```

**Examples:**
```bash
lychee hf search
lychee hf search code
lychee hf search llama
```

**Output:**
```
NAME                             REGISTRY/REPOS ID
Llama 3.2 3B Instruct           bartowski/Llama-3.2-3B-Instruct-GGUF
CodeLlama 7B Instruct            TheBloke/CodeLlama-7B-Instruct-GGUF
...
```

### `lychee hf list`

List all known HuggingFace models in the catalog.

```
lychee hf list
```

---

## Benchmarking & Analysis

### `lychee bench`

Precision benchmarking of models with configurable parameters. Measures prefill speed, generation speed, TTFT, load time, and total time.

```
lychee bench [flags]
```

| Flag | Type | Default | Description |
|:---|:---|:---|:---|
| `--model` | string | (required) | Model(s) to benchmark (comma-separated) |
| `--epochs` | int | `6` | Number of epochs per model |
| `--max-tokens` | int | `200` | Max tokens per response |
| `--temperature` | float | `0` | Temperature |
| `--seed` | int | `0` | Random seed |
| `--timeout` | int | `300` | Timeout in seconds |
| `-p`, `--prompt` | string | (default story) | Benchmark prompt |
| `--image` | string | | Image file for multimodal benchmarks |
| `-k`, `--keepalive` | float | `0` | Keep-alive duration between epochs |
| `--format` | string | `benchstat` | Output format: `benchstat` or `csv` |
| `-o`, `--output` | string | | Output file (stdout if empty) |
| `--verbose` | bool | false | Show system info |
| `--debug` | bool | false | Show debug output |
| `--warmup` | int | `1` | Warmup requests before timing |
| `--prompt-tokens` | int | `0` | Generate prompt targeting ~N tokens |
| `--num-ctx` | int | `0` | Context size override |

**Examples:**
```bash
# Basic bench
lychee bench --model llama3.2:3b

# Compare multiple models
lychee bench --model llama3.2:3b,phi-3:mini --epochs 10

# CSV output to file
lychee bench --model llama3.2:3b --format csv -o results.csv

# With custom prompt
lychee bench --model llama3.2:3b -p "Write a poem about AI" --epochs 3

# With image (multimodal)
lychee bench --model llava:13b --image photo.jpg
```

**Output (benchstat):**
```
# Model: llama3.2:3b | Params: 3.2B | Quant: Q4_K_M | Family: llama | Size: 1999999999 | VRAM: 1999999999
BenchmarkModel/name=llama3.2:3b/step=prefill 1 1234.56 ns/token 810.00 token/sec
BenchmarkModel/name=llama3.2:3b/step=generate 1 45.67 ns/token 21895.00 token/sec
BenchmarkModel/name=llama3.2:3b/step=ttft 1 345678901 ns/op
BenchmarkModel/name=llama3.2:3b/step=load 1 2500000000 ns/op
BenchmarkModel/name=llama3.2:3b/step=total 1 5234567890 ns/op
```

### `lychee benchmark [PROMPT]`

Benchmark multiple models against the same prompt and compare performance side-by-side with a ranked leaderboard.

```
lychee benchmark [PROMPT] [flags]
```

| Flag | Type | Default | Description |
|:---|:---|:---|:---|
| `--models` | string | (all installed) | Comma-separated list of models |
| `--runs` | int | `1` | Runs per model for averaging |

**Examples:**
```bash
# Compare all installed models
lychee benchmark "Explain quantum computing"

# Compare specific models with 3 runs each
lychee benchmark --models llama3.2:3b,phi-3:mini --runs 3 "What is AI?"

# Custom prompt
lychee benchmark "Compare procedural vs OOP programming"
```

**Output:**
```
🔬 Benchmarking 3 model(s) with 3 run(s) each...
Prompt: Explain quantum computing

⏳ Benchmarking llama3.2:3b...
  Run 1/3...
  Run 2/3...
  Run 3/3...
  ✅ 45.2 tok/s | TTFT: 234.50ms | Total: 4.42s | Tokens: 200

⏳ Benchmarking qwen3:4b...
  ...

#  MODEL              TOK/S    TTFT        TOTAL TIME  TOKENS  STATUS
🥇 qwen3:4b           52.3     198.20ms    3.82s       200     ✅
🥈 llama3.2:3b        45.2     234.50ms    4.42s       200     ✅
🥉 phi-3:mini         38.1     345.10ms    5.25s       200     ✅

💡 Higher TOK/S is better. TTFT = Time To First Token.
```

### `lychee compare`

Compare model outputs side-by-side for quality evaluation.

```
lychee compare [flags]
```

Run multiple models with the same prompt and display their outputs side by side for quality comparison.

**Examples:**
```bash
lychee compare --models llama3.2:3b,qwen3:4b "Write a poem about spring"
```

### `lychee stats`

Show live performance stats for all running models.

```
lychee stats [flags]
```

| Flag | Type | Default | Description |
|:---|:---|:---|:---|
| `--interval` | int | `1` | Refresh interval in seconds |
| `--tui` | bool | false | Use rich Bubbletea TUI dashboard |

**Examples:**
```bash
# Live stats updating every second
lychee stats

# Slower refresh
lychee stats --interval 2

# Rich TUI dashboard
lychee stats --tui
```

**Output:**
```
  Lychee Stats  [15:30:45]  Ctrl+C to exit
  ────────────────────────────────────────────────────────────
  MODEL                        VRAM        RAM         CTX        EXPIRES
  ────────────────────────────────────────────────────────────
  llama3.2:latest              4.5 GB      ─           128k       4m23s
```

### `lychee scan`

Scan system hardware and recommend models that fit your setup.

```
lychee scan [flags]
```

| Flag | Type | Description |
|:---|:---|:---|
| `--json` | bool | Output as JSON |

**Examples:**
```bash
lychee scan
lychee scan --json
```

**Output:**
```
  Scanning system hardware...

  Platform:         windows/amd64 (16 CPUs)
  System RAM:       31.8 GB
  GPU VRAM:         8.0 GB

  Recommended models for your hardware:

  MODEL                        SIZE       FITS     NOTES
  ───────────────────────────────────────────────────────────
  qwen3:0.6b                   394 MB     ✓        any hardware, fastest
  llama3.2:1b                  800 MB     ✓        edge device / low RAM
  qwen3:1.7b                   1.1 GB     ✓        good quality, very fast
  llama3.2:3b                  2.0 GB     ✓        great balance
  qwen3:4b                     2.5 GB     ✓        solid quality
  mistral:7b                   4.3 GB     ✓        workhorse model
  llama3.1:8b                  4.8 GB     ✓        best 8B model
  qwen3:8b                     5.0 GB     ✓        reasoning capable
  phi4:14b                     8.4 GB     ✗        Microsoft, very capable
  qwen3:14b                    9.1 GB     ✗        high quality reasoning

  Pull any model with:  lychee hf pull <repo>
```

### `lychee inspect MODEL`

Inspect model metadata and predict VRAM usage for different context lengths.

```
lychee inspect MODEL
```

**Examples:**
```bash
lychee inspect llama3.2:3b
```

**Output:**
```
🍒 Lychee Model Inspector: llama3.2:3b

Model Properties:
  Architecture:   llama
  Parameters:     3.2B
  Quantization:   Q4_K_M
  File Format:    gguf
  File Size:      2.0 GB

Estimated VRAM Memory Layout Profiles:
┌────────────────┬──────────────┬──────────────┬───────────────────────┐
│ CONTEXT WINDOW │ WEIGHTS SIZE │ KV CACHE SIZE│ ESTIMATED TOTAL VRAM  │
├────────────────┼──────────────┼──────────────┼───────────────────────┤
│ 2048 tokens    │ 2.0 GB       │ 49 MB        │ 2.4 GB                │
│ 4096 tokens    │ 2.0 GB       │ 98 MB        │ 2.5 GB                │
│ 8192 tokens    │ 2.0 GB       │ 196 MB       │ 2.7 GB                │
│ 16384 tokens   │ 2.0 GB       │ 392 MB       │ 3.2 GB                │
│ 32768 tokens   │ 2.0 GB       │ 784 MB       │ 4.1 GB                │
└────────────────┴──────────────┴──────────────┴───────────────────────┘

* Note: Predictions include base model weights, KV Cache context scaling, and a 300 MiB compute graph overhead.
```

---

## Configuration & Tools

### `lychee config`

Show/edit configuration settings.

```
lychee config
```

Shows current Lychee configuration including host, model paths, and runtime settings.

### `lychee search QUERY`

Search for models across registries and catalogs.

```
lychee search QUERY [flags]
```

| Flag | Type | Description |
|:---|:---|:---|
| `--catalog` | string | Custom catalog server URL |
| `-t`, `--token` | string | Auth token for custom catalog |

**Examples:**
```bash
lychee search llama
lychee search code --catalog http://192.168.1.100:9090/
```

### `lychee modelfile`

Manage and analyze Modelfiles.

**Subcommands:**

#### `lychee modelfile lint [filename]`

Lint a Modelfile for syntax errors and deprecated directives.

```bash
lychee modelfile lint
lychee modelfile lint ./Modelfile.custom
```

**Output:**
```
✅ Lint passed for Modelfile. No issues found.
```

#### `lychee modelfile init`

Interactively create a new Modelfile with a wizard.

```bash
lychee modelfile init
```

**Output:**
```
🍒 Welcome to the Lychee Modelfile Builder wizard! Let's build your Modelfile.
Enter base model name or path [default: llama3.2]:
Enter system prompt (instruction) [default: None]:
Enter temperature [default: 0.7]:
Enter context length [default: 2048]:
✅ Success! Formatted Modelfile saved to ./Modelfile
```

### `lychee quantize INPUT OUTPUT TYPE`

Quantize a local GGUF model file.

```
lychee quantize INPUT OUTPUT TYPE
```

**Quantization types:** `Q2_K`, `Q3_K_S`, `Q3_K_M`, `Q3_K_L`, `Q4_0`, `Q4_K_S`, `Q4_K_M`, `Q5_0`, `Q5_K_S`, `Q5_K_M`, `Q6_K`, `Q8_0`, `F16`

**Examples:**
```bash
lychee quantize ./model-f16.gguf ./model-q4.gguf Q4_K_M
```

**Output:**
```
Using quantize binary: /path/to/llama-quantize
Quantizing ./model-f16.gguf -> ./model-q4.gguf (Q4_K_M)
Quantizing tensors  100% ▕██████████████████████████████▏ 200/200
```

---

## Distribution & Sharing

### `lychee export MODEL OUTPUT_FILE.lychee`

Export a model and all its blobs into a portable archive for sharing.

```
lychee export MODEL OUTPUT_FILE.lychee
```

**Examples:**
```bash
lychee export my-custom-model my-model.lychee
```

**Output:**
```
Exporting model my-custom-model -> my-model.lychee...
Packing layer blob sha256:abc123... (2.0 GB)
✅ Model successfully exported!
```

### `lychee import ARCHIVE_FILE.lychee`

Import a portable model archive into your local catalog.

```
lychee import ARCHIVE_FILE.lychee
```

**Examples:**
```bash
lychee import my-model.lychee
```

**Output:**
```
Importing model from my-model.lychee...
Extracting blob sha256:abc123...
✅ Model my-custom-model successfully imported!
```

### `lychee serve-catalog`

Serve local model list as a shareable JSON catalog over HTTP.

```
lychee serve-catalog [flags]
```

| Flag | Type | Default | Description |
|:---|:---|:---|:---|
| `-p`, `--port` | int | `9090` | Port to run on |
| `-t`, `--token` | string | | Optional bearer token for authentication |

**Examples:**
```bash
lychee serve-catalog
lychee serve-catalog --port 8080
lychee serve-catalog --token my-secret-token
```

**Output:**
```
🍒 Lychee Registry Catalog Server listening on http://localhost:9090/catalog
Other users on your network can query this using:
  lychee search --catalog http://<your-ip>:9090/ <query>
```

### `lychee signin` / `lychee login` (hidden)

Sign in to lychee.com for cloud features.

```bash
lychee signin
lychee login  # alias (hidden)
```

### `lychee signout` / `lychee logout` (hidden)

Sign out from lychee.com.

```bash
lychee signout
lychee logout  # alias (hidden)
```

---

## Developer Tools

### `lychee completion [bash|zsh|fish|powershell]`

Generate shell autocompletion scripts.

```
lychee completion <shell>
```

**Examples:**
```bash
# Bash
source <(lychee completion bash)
lychee completion bash > /etc/bash_completion.d/lychee

# Zsh
source <(lychee completion zsh)

# Fish
lychee completion fish > ~/.config/fish/completions/lychee.fish

# PowerShell
lychee completion powershell | Out-String | Invoke-Expression
```

### `lychee update`

Check for and install the latest version of Lychee.

```
lychee update [flags]
```

| Flag | Type | Description |
|:---|:---|:---|
| `--check` | bool | Check for updates without installing |
| `--binary` | bool | Download pre-built binary instead of `go install` |

**Examples:**
```bash
# Update to latest
lychee update

# Check only
lychee update --check

# Use pre-built binary
lychee update --binary
```

**Output:**
```
Current: 0.1.0
Latest:  0.2.0

Updating...
✅ Updated to 0.2.0
```

### `lychee community`

Display community links, Discord invite, and open-source contribution opportunities.

```bash
lychee community
```

**Subcommands:**
- `lychee community claim ISSUE_ID` — Claim and get setup guide for a GitHub issue
- `lychee community feedback MODEL_ID` — View community reviews for a model

### `lychee generate-client`

Generate API client boilerplate code in various languages.

```
lychee generate-client [flags]
```

(Generates client code snippets for Python, JavaScript, Go, and curl based on the Lychee API)

### `lychee version`

Show version information (also available via `-v` flag).

```bash
lychee --version
lychee -v
```

**Output:**
```
lychee version is 0.1.0
```

---

## Quick Reference — Common Workflows

### Start Developing with a Model

```bash
# 1. Start the server
lychee serve

# 2. Pull a model (auto HF detect)
lychee pull microsoft/Phi-3-mini-4k-instruct-gguf

# 3. Pull and immediately chat
lychee pull microsoft/Phi-3-mini-4k --run

# 4. Or list & run separately
lychee list
lychee run phi-3:mini "Hello!"
```

### Benchmark Pipeline

```bash
# Scan hardware for recommendations
lychee scan

# Pull recommended models
lychee pull qwen3:0.6b
lychee pull llama3.2:3b

# Compare performance
lychee benchmark --models qwen3:0.6b,llama3.2:3b --runs 3

# Detailed bench with CSV export
lychee bench --model llama3.2:3b --epochs 10 --format csv -o bench.csv
```

### Custom Model Workflow

```bash
# Create a Modelfile
lychee modelfile init

# Lint it
lychee modelfile lint

# Create the model
lychee create my-assistant -f ./Modelfile

# Run it
lychee run my-assistant
```

### Multi-Model Pipeline

```bash
# Create steps JSON
cat > pipeline.json << 'EOF'
[
  {"model": "gemma3:4b", "prompt": "Translate to French: {{input}}"},
  {"model": "llama3.2:3b", "prompt": "Summarize: {{step[0].output}}"}
]
EOF

# Execute
lychee compose --steps pipeline.json --input "Hello World"
```

### Share Models

```bash
# Export a model
lychee export my-custom-model my-model.lychee

# On another machine, import it
lychee import my-model.lychee
```

---

## See Also

- [API Guide](./API-GUIDE.md) — Full REST API reference
- [Documentation Index](./INDEX.md) — All docs
- [README](../README.md) — Project overview
