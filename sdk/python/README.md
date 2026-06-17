# Lychee Python SDK

The official Python client for [Lychee](https://github.com/MD-Mushfiqur123/lychee) — the universal local LLM runtime.

## Installation

```bash
pip install -e .
# or
pip install lychee
```

Or install from requirements:

```bash
pip install -r requirements.txt
```

## Quick Start

```python
from lychee import LycheeClient

client = LycheeClient()  # defaults to http://localhost:11434

# Chat with a model
response = client.chat("gemma3", [
    {"role": "user", "content": "What is 2 + 2?"}
])
print(response["message"]["content"])

# Generate text (single-turn)
response = client.generate("gemma3", "Write a haiku about coding.")
print(response["response"])

# List locally installed models
models = client.list_models()
for m in models:
    print(f"{m['name']} ({m['size']} bytes)")

# Pull a model from HuggingFace
for progress in client.pull_model("bartowski/Meta-Llama-3.1-8B-Instruct-GGUF"):
    status = progress.get("status", "")
    total = progress.get("total", 0)
    completed = progress.get("completed", 0)
    print(f"\r{status}: {completed}/{total}", end="", flush=True)
print()

# Get embeddings
vector = client.embeddings("nomic-embed-text", "Hello world")
print(f"Embedding size: {len(vector)}")

# Multi-model pipeline (DAG)
result = client.compose(
    input="Translate and summarize: Hello world",
    steps=[
        {"model": "gemma3", "prompt": "Translate to French: {{input}}"},
        {"model": "llama3.2", "prompt": "Summarize: {{step[0].output}}"},
    ]
)
print(result["output"])

# Structured output with JSON schema
person = client.structured(
    model="gemma3",
    prompt="Extract the name and age from: Alice is 30 years old.",
    schema={"type": "object", "properties": {"name": {"type": "string"}, "age": {"type": "integer"}}}
)
print(person)
```

## Streaming

```python
# Stream generated text token by token
for chunk in client.generate("gemma3", "Tell me a story about a robot.", stream=True):
    print(chunk.get("response", ""), end="", flush=True)
print()

# Stream chat responses
for chunk in client.chat("gemma3", [{"role": "user", "content": "Write a poem."}], stream=True):
    print(chunk.get("message", {}).get("content", ""), end="", flush=True)
print()
```

## Batch Embeddings

```python
# Get embeddings for multiple strings at once
vectors = client.embeddings_batch("nomic-embed-text", [
    "What is machine learning?",
    "How does a neural network work?",
    "Explain transformers.",
])
for i, vec in enumerate(vectors):
    print(f"Input {i}: {len(vec)}-dimensional vector")
```

## Model Management

```python
# Show model details
info = client.show_model("gemma3")
print(info["license"])

# List running models
running = client.list_running()
print(running)

# Delete a model
client.delete_model("old-model")

# Check server health
if client.health():
    print("Lychee server is running!")
    print(client.version())
```

## Using as Context Manager

```python
with LycheeClient() as client:
    response = client.chat("gemma3", [{"role": "user", "content": "Hello!"}])
    print(response["message"]["content"])
# Session is automatically closed
```

## API Reference

### `LycheeClient(host="http://localhost:11434", timeout=120.0)`

Create a client connected to a Lychee server.

### Methods

| Method | Description |
|---|---|
| `chat(model, messages, *, stream=False, **options)` | Chat with a model using a message list |
| `generate(model, prompt, *, stream=False, **options)` | Generate text from a model (single-turn) |
| `list_models()` | List all locally installed models |
| `pull_model(name)` | Pull a model from HuggingFace (returns progress iterator) |
| `embeddings(model, input)` | Get embedding vector for a string |
| `embeddings_batch(model, inputs)` | Get embeddings for multiple strings |
| `show_model(model)` | Show model details (license, template, etc.) |
| `delete_model(model)` | Delete a local model |
| `list_running()` | List models currently loaded in memory |
| `compose(input, steps, *, stream=False)` | Execute multi-model DAG pipeline |
| `structured(model, prompt, schema, *, max_retries=3, **options)` | Generate schema-conforming JSON with auto-retry |
| `version()` | Get Lychee server version |
| `health()` | Check if the server is reachable |
| `close()` | Close the HTTP session |

## Requirements

- Python 3.8+
- `requests`

## License

MIT — see the [main repository](https://github.com/MD-Mushfiqur123/lychee/blob/main/LICENSE).
