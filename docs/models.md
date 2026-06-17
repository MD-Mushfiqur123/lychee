---
title: Models Quick Start
---

A curated guide to getting started with models in Lychee — what to pull, how to run them, and how quantization works.

## Table of Contents

- [Recommended Starter Models](#recommended-starter-models)
  - [Tiny (<1 GB)](#tiny-1-gb)
  - [Small (1–3 GB)](#small-13-gb)
  - [Medium (4–8 GB)](#medium-48-gb)
  - [Large (8 GB+)](#large-8-gb)
- [Pulling a Model](#pulling-a-model)
- [Running a Model](#running-a-model)
- [Hugging Face vs Lychee Native Models](#hugging-face-vs-lychee-native-models)
- [Quantization Options](#quantization-options)
  - [What Is Quantization?](#what-is-quantization)
  - [Available Quantization Levels](#available-quantization-levels)
  - [Choosing the Right Quantization](#choosing-the-right-quantization)
  - [How to Pull a Specific Quant](#how-to-pull-a-specific-quant)
- [Managing Your Models](#managing-your-models)
- [Next Steps](#next-steps)

---

## Recommended Starter Models

These models are known to work well with Lychee out of the box. Pick the tier that fits your hardware.

### Tiny (<1 GB)

Best for CPU-only machines, low-RAM environments, or quick experiments. Fast but limited reasoning.

| Model | Size | Best For |
|-------|------|----------|
| **Qwen2.5-0.5B** | ~0.4 GB | Quick prototyping, embedded systems |
| **SmolLM2-135M** | ~0.1 GB | Smoke tests, CI pipelines, constrained environments |

```bash
lychee pull qwen2.5:0.5b
lychee pull smollm2:135m
```

### Small (1–3 GB)

Good balance of speed and capability. Runs comfortably on most laptops with 8 GB RAM.

| Model | Size | Best For |
|-------|------|----------|
| **Qwen2.5-3B** | ~2 GB | Strong multilingual, strong reasoning for its size |
| **Phi-3-mini** | ~2 GB | Compact, instruction-tuned, excellent code generation |

```bash
lychee pull qwen2.5:3b
lychee pull phi3:mini
```

### Medium (4–8 GB)

The sweet spot for local development. Capable enough for coding assistants and serious chat.

| Model | Size | Best For |
|-------|------|----------|
| **Llama 3.2-8B** | ~5 GB | Latest Llama generation, great tool use and instruction following |
| **Mistral-7B** | ~4 GB | Fast inference, strong general-purpose reasoning |

```bash
lychee pull llama3.2
lychee pull mistral
```

### Large (8 GB+)

For GPU-equipped machines. These models rival cloud-hosted AI in capability.

| Model | Size | Best For |
|-------|------|----------|
| **Llama 3.3-70B** | ~40 GB | Frontier open-weight model, deep reasoning |
| **Qwen2.5-72B** | ~43 GB | Best-in-class multilingual, extremely capable |

```bash
lychee pull llama3.3:70b
lychee pull qwen2.5:72b
```

> **💡 Tip:** Not sure what to start with? `llama3.2` is the default recommendation for most users.

---

## Pulling a Model

Lychee downloads models from its registry using the `pull` command:

```bash
lychee pull <org>/<model>
# or the shorthand:
lychee pull <model>
```

Examples:

```bash
# Pull the latest (default) tag
lychee pull llama3.2

# Pull a specific model variant
lychee pull qwen2.5:0.5b

# Pull by full org/model path
lychee pull meta-llama/llama-3.2-8b
```

After pulling, the model is stored locally and ready to use.

```bash
# List all downloaded models
lychee list
```

---

## Running a Model

Once pulled, run a model with:

```bash
lychee run <model> "your prompt"
```

Examples:

```bash
# One-shot completion
lychee run llama3.2 "Explain quantum computing in one sentence."

# Interactive chat session
lychee run llama3.2

# Run a specific variant
lychee run qwen2.5:3b "Write a haiku about coding."
```

In interactive mode, type your messages and press Enter. Use `/bye` to exit or Ctrl+C.

---

## Hugging Face vs Lychee Native Models

Lychee supports models from two sources:

### Lychee Native Models

Pre-converted, tested, and optimized for Lychee. Pull with a simple name:

```bash
lychee pull llama3.2
```

**Advantages:**
- No conversion needed — just pull and run
- Pre-configured with correct templates and parameters
- Community-tested and verified

### Hugging Face GGUF Models

Any GGUF-format model on Hugging Face can be imported:

```bash
# Pull a GGUF from Hugging Face
lychee pull hf.co/bartowski/Meta-Llama-3.1-8B-Instruct-GGUF:Q4_K_M

# Or use the full Hugging Face URL
lychee pull huggingface.co/bartowski/Meta-Llama-3.1-8B-Instruct-GGUF:Q4_K_M
```

**Advantages:**
- Access to thousands of community-converted models
- Fine-grained quantization selection (every quant available)
- Bleeding-edge models available immediately

### Which Should You Use?

| Scenario | Recommendation |
|----------|---------------|
| Getting started, want simplicity | **Lychee native** — `lychee pull llama3.2` |
| Need a specific quantization | **Hugging Face GGUF** — `lychee pull hf.co/bartowski/...` |
| Want bleeding-edge or niche models | **Hugging Face GGUF** |
| Building a Modelfile with a custom system prompt | Either — both work with `FROM` |

---

## Quantization Options

### What Is Quantization?

Quantization reduces the precision of a model's weights to make it smaller and faster, at a small cost to quality. Think of it like compression: a JPEG is smaller than a RAW file but still looks great.

### Available Quantization Levels

Quantization levels use the **K-quant** naming convention (from `llama.cpp`), ranging from smallest/lowest quality to largest/highest quality:

| Level | Suffix | Size vs FP16 | Quality | Best For |
|-------|--------|-------------|---------|----------|
| **2-bit** | `Q2_K` | ~15% | ⚠️ Significant loss | Extreme compression, tiny devices |
| **3-bit** | `Q3_K_S`, `Q3_K_M`, `Q3_K_L` | ~20% | ⚠️ Noticeable loss | Very constrained systems |
| **4-bit** | `Q4_0`, `Q4_K_S`, `Q4_K_M` | ~28% | ✅ Good — sweet spot | **Default recommendation** for most users |
| **5-bit** | `Q5_0`, `Q5_K_S`, `Q5_K_M` | ~35% | ✅ Very good | When you can afford slightly more RAM |
| **6-bit** | `Q6_K` | ~43% | ✅ Excellent | Near-perfect quality, decent size |
| **8-bit** | `Q8_0` | ~55% | ✅ Near-lossless | High quality, moderate compression |
| **FP16** | `F16` | 100% | ✅ Lossless | Full original quality, no compression |

**Suffix meanings:**

- **`S` (Small)** — Maximum compression for the bit level
- **`M` (Medium)** — Balanced size vs. quality **(recommended default)**
- **`L` (Large)** — Larger than M, better quality

### Choosing the Right Quantization

```
Quality ↑
  |
  |  Q8_0, F16    ← "I want the best quality possible"
  |
  |  Q5_K_M       ← "I have extra RAM to spare"
  |  Q6_K
  |
  |  Q4_K_M       ← ★ START HERE for most users
  |
  |  Q3_K_M       ← "I'm tight on memory"
  |
  |  Q2_K         ← "I just need it to run"
  |
  +------------------------------------------------→ Size ↓
```

> **Rule of thumb:** `Q4_K_M` is the default for most Lychee native models. It delivers ~95% of FP16 quality at ~28% of the size.

### How to Pull a Specific Quant

When using Hugging Face GGUF models, append the quantization tag:

```bash
# Pull Q4_K_M (balanced — recommended)
lychee pull hf.co/bartowski/Meta-Llama-3.1-8B-Instruct-GGUF:Q4_K_M

# Pull Q8_0 (near-lossless)
lychee pull hf.co/bartowski/Meta-Llama-3.1-8B-Instruct-GGUF:Q8_0

# Pull FP16 (full quality)
lychee pull hf.co/bartowski/Meta-Llama-3.1-8B-Instruct-GGUF:F16
```

For Lychee native models, the default tag usually ships with a well-chosen quantization. Check available tags:

```bash
lychee show llama3.2
```

---

## Managing Your Models

```bash
# List all downloaded models
lychee list

# Show model details (parameters, quantization, template)
lychee show llama3.2

# Remove a model to free up disk space
lychee rm llama3.2

# Copy a model (e.g., to create a custom version)
lychee cp llama3.2 my-custom-llama

# Pull a model silently (no progress output)
lychee pull --quiet llama3.2
```

---

## Next Steps

- **[Modelfile Reference](/docs/modelfile)** — Customize models with system prompts, parameters, and templates
- **[CLI Reference](/docs/cli-reference)** — Full command list and options
- **[GPU Guide](/docs/gpu)** — Configure GPU acceleration
- **[Importing Models](/docs/import)** — Bring your own GGUF or Safetensors models
- **[API Documentation](/docs/api)** — Integrate Lychee into your applications
