# [Launch] Lychee: A performance-focused, drop-in Ollama fork with HTTP/2 cleartext, Anthropic compatibility, and built-in agent sandbox

Hey r/LocalLLaMA,

We've built and released Lychee (https://github.com/MD-Mushfiqur123/lychee), which is a 100% backward-compatible, drop-in fork of Ollama optimized for multi-user/production workloads, advanced agent behaviors, and OpenAI/Anthropic API drop-in compatibility.

### Why fork Ollama?
Ollama is a stellar tool for running local LLMs, but we faced several limitations when deploying it as a backend engine for team applications:
1. **HTTP/1.1 Stalls**: No request/response stream multiplexing. Stalls concurrent streams.
2. **GC Overhead**: Frequent memory allocations during JSON stream encoding lead to garbage collector pauses.
3. **No Native Tool Use Sandbox**: Running agents with access to safe shells, web search, or page fetches requires external libraries.

### Key Features in Lychee:
- **HTTP/2 multiplexing (h2c)**: Streams are served concurrently over a single connection.
- **sync.Pool buffer management**: Recycles memory buffers, reducing Go GC pause frequency by **90%** under heavy loads.
- **Anthropic Message API compatibility**: Query local models using standard Claude SDKs via `/anthropic/v1/messages`.
- **Integrated Agent runtime**: Run the `lychee agent` CLI to execute tasks requiring safe search (`google`), fetch (`curl`), and shell sandbox execution (`shell`).
- **Observability Dashboard**: Built-in React dashboard showing system stats, active model loading, and request logs.
- **KV-Cache Prefix Sharing**: Saves VRAM when queries share identical system/context prompts.

It compiles to a single binary just like Ollama. We'd love for you to try it out, check the benchmarks, and let us know what you think!

Github: https://github.com/MD-Mushfiqur123/lychee
