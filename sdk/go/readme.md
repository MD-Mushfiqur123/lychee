# Lychee Go SDK

Official Go client for [Lychee](https://github.com/MD-Mushfiqur123/lychee) — the universal local LLM runtime.

## Installation

```bash
go get github.com/MD-Mushfiqur123/lychee/sdk/go
```

## Example 1: Generate text

```go
package main

import (
    "context"
    "fmt"

    "github.com/MD-Mushfiqur123/lychee/sdk/go"
)

func main() {
    ctx := context.Background()
    client := lychee.NewClient("") // defaults to http://localhost:11434

    resp, err := client.Generate(ctx, &lychee.GenerateRequest{
        Model:  "gemma3",
        Prompt: "Explain quantum computing in one sentence.",
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Response)
}
```

## Example 2: Chat and list models

```go
package main

import (
    "context"
    "fmt"

    "github.com/MD-Mushfiqur123/lychee/sdk/go"
)

func main() {
    ctx := context.Background()
    client := lychee.NewClient("http://localhost:11434")

    // Chat with a model
    chatResp, err := client.Chat(ctx, &lychee.ChatRequest{
        Model: "gemma3",
        Messages: []lychee.Message{
            {Role: "user", Content: "What is 2 + 2?"},
        },
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(chatResp.Message.Content)

    // List available models
    models, err := client.ListModels(ctx)
    if err != nil {
        panic(err)
    }
    for _, m := range models {
        fmt.Printf("%s (%d bytes)\n", m.Name, m.Size)
    }
}
```

## API Reference

| Method | Description |
|---|---|
| `Generate(ctx, req)` | Generate text (non-streaming) |
| `GenerateStream(ctx, req, fn)` | Generate text with streaming callback |
| `Chat(ctx, req)` | Chat with message history |
| `ChatStream(ctx, req, fn)` | Chat with streaming callback |
| `ListModels(ctx)` | List locally installed models |
| `PullModel(ctx, name, fn)` | Pull a model from HuggingFace |
| `Embeddings(ctx, req)` | Get embedding vectors |
| `DeleteModel(ctx, name)` | Delete a local model |
| `Version(ctx)` | Get server version |
| `Health(ctx)` | Check server reachability |

## Requirements

- Go 1.23+
- A running [Lychee server](https://github.com/MD-Mushfiqur123/lychee)

## License

MIT
