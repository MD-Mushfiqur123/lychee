# Lychee Ruby SDK

The official Ruby client for [Lychee](https://github.com/MD-Mushfiqur123/lychee) — the universal local LLM runtime.

## Installation

Add to your `Gemfile`:

```ruby
gem 'lychee', path: 'sdk/ruby'
```

Or install directly:

```bash
gem build lychee.gemspec
gem install lychee-0.2.0.gem
```

> 💡 Coming soon to RubyGems: `gem install lychee`

## Quick Start

```ruby
require 'lychee'

client = LycheeClient.new  # defaults to http://localhost:11434

# Chat with a model
response = client.chat("gemma3", [
  { role: "user", content: "What is 2 + 2?" }
])
puts response["message"]["content"]

# Generate text (single-turn)
response = client.generate("gemma3", "Write a haiku about coding.")
puts response["response"]

# List locally installed models
models = client.list_models
models["models"].each do |m|
  puts "#{m['name']} (#{m['size']} bytes)"
end

# Pull a model from HuggingFace
client.pull_model("bartowski/Meta-Llama-3.1-8B-Instruct-GGUF") do |progress|
  status = progress["status"] || ""
  total  = progress["total"] || 0
  completed = progress["completed"] || 0
  print "\r#{status}: #{completed}/#{total}"
end
puts

# Get embeddings
vector = client.embeddings("nomic-embed-text", "Hello world")
puts "Embedding size: #{vector.length}"

# Multi-model pipeline (DAG)
result = client.compose(
  "Translate and summarize: Hello world",
  [
    { model: "gemma3",   prompt: "Translate to French: {{input}}" },
    { model: "llama3.2", prompt: "Summarize: {{step[0].output}}" }
  ]
)
puts result["output"]

# Structured output with JSON schema
person = client.structured(
  "gemma3",
  "Extract the name and age from: Alice is 30 years old.",
  { type: "object", properties: { name: { type: "string" }, age: { type: "integer" } } }
)
puts person
```

## Streaming

```ruby
# Stream generated text token by token
client.generate("gemma3", "Tell me a story about a robot.", stream: true) do |chunk|
  print chunk["response"]
end
puts

# Stream chat responses
client.chat("gemma3", [{ role: "user", content: "Write a poem." }], stream: true) do |chunk|
  print chunk.dig("message", "content")
end
puts
```

## Batch Embeddings

```ruby
# Get embeddings for multiple strings at once
vectors = client.embeddings_batch("nomic-embed-text", [
  "What is machine learning?",
  "How does a neural network work?",
  "Explain transformers."
])
vectors.each_with_index do |vec, i|
  puts "Input #{i}: #{vec.length}-dimensional vector"
end
```

## Model Management

```ruby
# Show model details
info = client.show_model("gemma3")
puts info["license"]

# List running models
running = client.list_running
puts running

# Delete a model
client.delete_model("old-model")

# Check server health
if client.health
  puts "Lychee server is running!"
  puts client.version
end
```

## Custom Configuration

```ruby
# Connect to a remote Lychee server
client = LycheeClient.new("https://my-server.example.com:11434", timeout: 60)

# All methods work the same
resp = client.chat("gemma3", [{ role: "user", content: "Hello!" }])
puts resp["message"]["content"]
```

## API Reference

### `LycheeClient.new(host = "http://localhost:11434", timeout: 120)`

Create a client connected to a Lychee server.

### Methods

| Method | Description |
|---|---|
| `chat(model, messages, stream:, **opts)` | Chat with a model using a message list |
| `generate(model, prompt, stream:, **opts)` | Generate text from a model (single-turn) |
| `list_models` | List all locally installed models |
| `show_model(model)` | Show model details (license, template, etc.) |
| `delete_model(model)` | Delete a local model |
| `list_running` | List models currently loaded in memory |
| `pull_model(name, insecure:)` | Pull a model from HuggingFace (yields progress with block) |
| `embeddings(model, input)` | Get embedding vector for a string |
| `embeddings_batch(model, inputs)` | Get embeddings for multiple strings |
| `compose(input, steps, stream:)` | Execute multi-model DAG pipeline |
| `structured(model, prompt, schema, max_retries:, **opts)` | Generate schema-conforming JSON with auto-retry |
| `version` | Get Lychee server version |
| `health` | Check if the server is reachable |

## Requirements

- Ruby 2.7+
- No external dependencies (stdlib only: `net/http`, `json`, `uri`)

## License

MIT — see the [main repository](https://github.com/MD-Mushfiqur123/lychee/blob/main/LICENSE).
